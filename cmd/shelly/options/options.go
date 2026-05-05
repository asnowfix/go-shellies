// Package options holds CLI flags shared across subcommands of the shelly
// binary. It deliberately keeps no knowledge of any home-automation daemon —
// every subcommand talks directly to devices.
package options

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asnowfix/go-shellies/types"
	"github.com/go-logr/logr"
	"sigs.k8s.io/yaml"
)

const (
	MdnsLookupDefaultTimeout time.Duration = 7 * time.Second
	MqttDefaultTimeout       time.Duration = 14 * time.Second
	CommandDefaultTimeout    time.Duration = 0 // wait indefinitely
	ShellyDefaultRateLimit   time.Duration = 200 * time.Millisecond
)

var Flags struct {
	Verbose         bool
	Debug           bool
	Quiet           bool
	Json            bool
	MqttBroker      string
	MqttTimeout     time.Duration
	MdnsTimeout     time.Duration
	Wait            time.Duration
	ShellyRateLimit time.Duration
}

// Via is the channel selected by the user (HTTP via mDNS or MQTT via broker).
var Via types.Channel

// Log is the package-level logger. The root cobra command initializes it in
// PersistentPreRun based on --verbose / --debug.
var Log logr.Logger = logr.Discard()

// InitLog configures Log based on the current Flags.
func InitLog() {
	level := slog.LevelInfo
	switch {
	case Flags.Debug:
		level = slog.LevelDebug
	case Flags.Quiet:
		level = slog.LevelWarn
	case Flags.Verbose:
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	Log = logr.FromSlogHandler(handler)
}

// CommandLineContext returns a context that is cancelled on SIGINT/SIGTERM and
// optionally has the --command-timeout applied.
func CommandLineContext(parent context.Context) context.Context {
	ctx := parent
	var cancel context.CancelFunc
	if Flags.Wait > 0 {
		ctx, cancel = context.WithTimeout(parent, Flags.Wait)
	} else {
		ctx, cancel = context.WithCancel(parent)
	}
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		select {
		case <-signals:
			Log.Info("Received signal, cancelling")
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx
}

// Args returns the trailing positional args (everything after the device id).
func Args(args []string) []string {
	if len(args) > 1 {
		return args[1:]
	}
	return []string{}
}

// PrintResult renders an RPC result to stdout, optionally prefixed by a device
// header. JSON or YAML based on --json.
func PrintResult(out any, deviceName ...string) error {
	name := ""
	if len(deviceName) > 0 {
		name = deviceName[0]
	}
	if Flags.Json {
		var toMarshal any = out
		if name != "" {
			toMarshal = map[string]any{"device": name, "result": out}
		}
		s, err := json.Marshal(toMarshal)
		if err != nil {
			return err
		}
		fmt.Println(string(s))
		return nil
	}
	if name != "" {
		fmt.Printf("--- %s ---\n", name)
	}
	s, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	fmt.Print(string(s))
	return nil
}
