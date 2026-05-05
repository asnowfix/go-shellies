// Command shelly is a standalone CLI for talking to Shelly devices over the
// local network (HTTP/mDNS or MQTT). It does not depend on any home-automation
// daemon — every subcommand resolves devices and dispatches RPCs directly.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/asnowfix/go-shellies/cmd/shelly/call"
	"github.com/asnowfix/go-shellies/cmd/shelly/components"
	"github.com/asnowfix/go-shellies/cmd/shelly/emulate"
	"github.com/asnowfix/go-shellies/cmd/shelly/jobs"
	"github.com/asnowfix/go-shellies/cmd/shelly/kvs"
	"github.com/asnowfix/go-shellies/cmd/shelly/mcp"
	"github.com/asnowfix/go-shellies/cmd/shelly/mqtt"
	"github.com/asnowfix/go-shellies/cmd/shelly/options"
	"github.com/asnowfix/go-shellies/cmd/shelly/reboot"
	"github.com/asnowfix/go-shellies/cmd/shelly/status"
	"github.com/asnowfix/go-shellies/cmd/shelly/sys"
	"github.com/asnowfix/go-shellies/cmd/shelly/wifi"
	"github.com/asnowfix/go-shellies/types"
)

var rootCmd = &cobra.Command{
	Use:   "shelly",
	Short: "Direct CLI for Shelly devices",
	Long: `shelly is a CLI for the github.com/asnowfix/go-shellies library.

It discovers devices via mDNS (default) or addresses them by MQTT topic id
(--via mqtt) and runs RPC calls, KVS operations, MQTT/WiFi configuration,
and a JS device emulator without any external daemon.`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		options.InitLog()
		switch cmd.Flag("via").Value.String() {
		case "mqtt":
			options.Via = types.ChannelMqtt
		case "http", "":
			options.Via = types.ChannelHttp
		default:
			return fmt.Errorf("invalid --via %q (want http or mqtt)", cmd.Flag("via").Value.String())
		}
		return nil
	},
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.BoolVarP(&options.Flags.Verbose, "verbose", "v", false, "verbose output")
	pf.BoolVar(&options.Flags.Debug, "debug", false, "debug output (implies verbose)")
	pf.BoolVarP(&options.Flags.Quiet, "quiet", "q", false, "quiet output (warnings and errors only)")
	pf.BoolVar(&options.Flags.Json, "json", false, "render results as JSON instead of YAML")
	pf.StringVar(&options.Flags.MqttBroker, "mqtt-broker", "", "MQTT broker URL (e.g. tcp://192.168.1.1:1883), required for --via mqtt and `mqtt config`")
	pf.DurationVarP(&options.Flags.MqttTimeout, "mqtt-timeout", "T", options.MqttDefaultTimeout, "MQTT call timeout")
	pf.DurationVarP(&options.Flags.MdnsTimeout, "mdns-timeout", "M", options.MdnsLookupDefaultTimeout, "mDNS browse timeout")
	pf.DurationVarP(&options.Flags.Wait, "command-timeout", "C", options.CommandDefaultTimeout, "overall command timeout (0 = no timeout)")
	pf.DurationVar(&options.Flags.ShellyRateLimit, "shelly-rate-limit", options.ShellyDefaultRateLimit, "min interval between consecutive RPCs to the same device")
	pf.String("via", "http", "channel: http (mDNS+HTTP) or mqtt (MQTT broker)")

	rootCmd.AddCommand(call.Cmd)
	rootCmd.AddCommand(components.Cmd)
	rootCmd.AddCommand(emulate.Cmd)
	rootCmd.AddCommand(jobs.Cmd)
	rootCmd.AddCommand(kvs.Cmd)
	rootCmd.AddCommand(mcp.Cmd)
	rootCmd.AddCommand(mqtt.Cmd)
	rootCmd.AddCommand(reboot.Cmd)
	rootCmd.AddCommand(status.Cmd)
	rootCmd.AddCommand(sys.Cmd)
	rootCmd.AddCommand(wifi.Cmd)
}

func main() {
	ctx := options.CommandLineContext(context.Background())
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
