package shelly

import (
	"context"
	"github.com/asnowfix/go-shellies/ble"
	"github.com/asnowfix/go-shellies/ethernet"
	"github.com/asnowfix/go-shellies/input"
	"github.com/asnowfix/go-shellies/kvs"
	"github.com/asnowfix/go-shellies/matter"
	"github.com/asnowfix/go-shellies/mqtt"
	"github.com/asnowfix/go-shellies/ratelimit"
	"github.com/asnowfix/go-shellies/script"
	"github.com/asnowfix/go-shellies/shelly"
	shttp "github.com/asnowfix/go-shellies/shttp"
	"github.com/asnowfix/go-shellies/sswitch"
	"github.com/asnowfix/go-shellies/system"
	"github.com/asnowfix/go-shellies/types"
	"github.com/asnowfix/go-shellies/wifi"
	"reflect"
	"github.com/asnowfix/go-shellies/schedule"
	scripts "github.com/asnowfix/home-automation/internal/shelly/scripts"
	"time"

	"github.com/go-logr/logr"
)

type empty struct{}

func Init(log logr.Logger, mc mqtt.Client, timeout time.Duration, rateLimitInterval time.Duration) {
	log.Info("Init", "package", reflect.TypeOf(empty{}).PkgPath(), "rateLimit", rateLimitInterval)
	registrar.Init(log)

	// Initialize rate limiter
	ratelimit.Init(rateLimitInterval)

	// Keep in lexical order
	// gen1.Init(log, &registrar)
	shelly.Init(log, &registrar, timeout)
	ble.Init(log, &registrar)
	ethernet.Init(log, &registrar)
	input.Init(log, &registrar)
	kvs.Init(log, &registrar)
	matter.Init(log, &registrar)
	mqtt.Init(log, &registrar, mc, timeout)
	schedule.Init(log, &registrar)
	script.Init(log, &registrar, scripts.GetFS())
	shttp.Init(log, &registrar)
	sswitch.Init(log, &registrar)
	system.Init(log, &registrar)
	// temperature.Init(log, &registrar)
	wifi.Init(log, &registrar)
}

func (r *Registrar) CallE(ctx context.Context, d types.Device, via types.Channel, mh types.MethodHandler, params any) (any, error) {
	return r.channels[d.Channel(via)](ctx, d, mh, mh.Allocate(), params)
}
