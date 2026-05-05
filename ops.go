package shelly

import (
	"context"
	"reflect"
	"time"

	"github.com/asnowfix/go-shellies/ble"
	"github.com/asnowfix/go-shellies/ethernet"
	"github.com/asnowfix/go-shellies/input"
	"github.com/asnowfix/go-shellies/kvs"
	"github.com/asnowfix/go-shellies/matter"
	"github.com/asnowfix/go-shellies/mqtt"
	"github.com/asnowfix/go-shellies/ratelimit"
	"github.com/asnowfix/go-shellies/schedule"
	"github.com/asnowfix/go-shellies/shelly"
	shttp "github.com/asnowfix/go-shellies/shttp"
	"github.com/asnowfix/go-shellies/sswitch"
	"github.com/asnowfix/go-shellies/system"
	"github.com/asnowfix/go-shellies/types"
	"github.com/asnowfix/go-shellies/wifi"

	"github.com/go-logr/logr"
)

type empty struct{}

// InitExtra registers extra method handlers against the package-level
// registrar — used to plug optional sub-packages (script, gen1, ...) without
// pulling them in unconditionally.
type InitExtra func(log logr.Logger, r types.MethodsRegistrar)

// Init wires up the registrar, rate limiter, and the always-present method
// handlers. Optional extras (e.g. script.Init) are invoked at the end so a
// caller can opt into the heavier sub-packages.
func Init(log logr.Logger, mc mqtt.Client, timeout time.Duration, rateLimitInterval time.Duration, extras ...InitExtra) {
	log.Info("Init", "package", reflect.TypeOf(empty{}).PkgPath(), "rateLimit", rateLimitInterval)
	registrar.Init(log)

	ratelimit.Init(rateLimitInterval)

	shelly.Init(log, &registrar, timeout)
	ble.Init(log, &registrar)
	ethernet.Init(log, &registrar)
	input.Init(log, &registrar)
	kvs.Init(log, &registrar)
	matter.Init(log, &registrar)
	mqtt.Init(log, &registrar, mc, timeout)
	schedule.Init(log, &registrar)
	shttp.Init(log, &registrar)
	sswitch.Init(log, &registrar)
	system.Init(log, &registrar)
	wifi.Init(log, &registrar)

	for _, extra := range extras {
		extra(log, &registrar)
	}
}

func (r *Registrar) CallE(ctx context.Context, d types.Device, via types.Channel, mh types.MethodHandler, params any) (any, error) {
	return r.channels[d.Channel(via)](ctx, d, mh, mh.Allocate(), params)
}
