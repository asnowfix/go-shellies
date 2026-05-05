// Package dispatch resolves a user-supplied device identifier (name, id,
// hostname, IP, or wildcard) to one or more devices.Device values and runs
// a per-device operation against each, in parallel.
//
// It is the standalone-CLI replacement for the myhome daemon's Foreach which
// went through the daemon's device registry; here we only use mDNS, direct
// host resolution, and MQTT topic addressing.
package dispatch

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	shelly "github.com/asnowfix/go-shellies"
	"github.com/asnowfix/go-shellies/cmd/shelly/options"
	"github.com/asnowfix/go-shellies/devices"
	"github.com/asnowfix/go-shellies/types"
	"github.com/go-logr/logr"
	"github.com/grandcat/zeroconf"
)

// Foreach resolves `name` to one or more Shelly devices and runs `do` against
// each. Resolution rules:
//
//   - If `name` parses as an IP literal, the IP is used directly.
//   - If `via` is types.ChannelMqtt and `name` contains no wildcard,
//     the name is treated as the MQTT topic id.
//   - Otherwise mDNS is browsed for "_shelly._tcp." services and matched
//     against `name` (exact id/name match, substring, or trailing-* prefix).
//
// The aggregation, parallelism and error semantics mirror shelly.Foreach.
func Foreach(ctx context.Context, log logr.Logger, name string, via types.Channel, do shelly.Do, args []string) (any, error) {
	list, err := Lookup(ctx, log, name, via)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no Shelly device matched %q", name)
	}
	return shelly.Foreach(ctx, log, list, via, do, args)
}

// Lookup resolves a name to a slice of devices without running any operation.
func Lookup(ctx context.Context, log logr.Logger, name string, via types.Channel) ([]devices.Device, error) {
	if ip := net.ParseIP(name); ip != nil {
		d, err := shelly.NewDeviceFromIp(ctx, log, ip)
		if err != nil {
			return nil, err
		}
		return []devices.Device{d}, nil
	}
	if via == types.ChannelMqtt && !strings.ContainsAny(name, "*?") {
		d, err := shelly.NewDeviceFromMqttId(ctx, log, name)
		if err != nil {
			return nil, err
		}
		return []devices.Device{d}, nil
	}
	return mdnsDiscover(ctx, log, name)
}

func mdnsDiscover(ctx context.Context, log logr.Logger, pattern string) ([]devices.Device, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("zeroconf: %w", err)
	}

	timeout := options.Flags.MdnsTimeout
	if timeout <= 0 {
		timeout = options.MdnsLookupDefaultTimeout
	}
	sctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry, 16)
	if err := resolver.Browse(sctx, shelly.MDNS_SHELLIES, "local.", entries); err != nil {
		return nil, fmt.Errorf("zeroconf browse: %w", err)
	}

	var found []devices.Device
	deadline := time.After(timeout)
loop:
	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				break loop
			}
			d, err := shelly.NewDeviceFromZeroConfEntry(ctx, log, nil, entry)
			if err != nil {
				log.V(1).Info("skipping unparseable entry", "instance", entry.Instance, "err", err)
				continue
			}
			if matches(d, pattern) {
				found = append(found, d)
			}
		case <-deadline:
			break loop
		case <-ctx.Done():
			break loop
		}
	}
	return found, nil
}

func matches(d devices.Device, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	id := strings.ToLower(d.Id())
	name := strings.ToLower(d.Name())
	pat := strings.ToLower(pattern)

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pat, "*")
		return strings.HasPrefix(id, prefix) || strings.HasPrefix(name, prefix)
	}
	return id == pat || name == pat || strings.Contains(id, pat) || strings.Contains(name, pat)
}
