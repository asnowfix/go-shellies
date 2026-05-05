// Package emulate exposes the script package's goja-based Shelly device
// emulator as a CLI subcommand. Useful for running and debugging Shelly JS
// scripts on a workstation without flashing them to a real device.
package emulate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/asnowfix/go-shellies/cmd/shelly/options"
	"github.com/asnowfix/go-shellies/mqtt"
	"github.com/asnowfix/go-shellies/script"
)

var (
	deviceStateFile string
	saveState       bool
	duration        time.Duration
)

var Cmd = &cobra.Command{
	Use:   "emulate <script.js>",
	Short: "Run a Shelly JS script locally against an emulated device",
	Long: `Loads <script.js> into a goja runtime configured with Shelly globals
(KVS, Storage, Schedule, Switch, Input, Shelly.call, Timer, MQTT, ...) and
runs it for --duration (or until Ctrl-C if --duration is 0).

If --device-state points at a JSON file, the file is loaded as the initial
KVS / Storage / ComponentStatus snapshot. Pass --save to write the post-run
state back to the same file.`,
	Args: cobra.ExactArgs(1),
	RunE: run,
}

func init() {
	f := Cmd.Flags()
	f.StringVarP(&deviceStateFile, "device-state", "s", "", "JSON file with initial device state (KVS, Storage, ComponentStatus)")
	f.BoolVar(&saveState, "save", false, "write final device state back to --device-state on exit")
	f.DurationVarP(&duration, "duration", "d", 0, "stop after this duration (0 = run until Ctrl-C)")
}

func run(cmd *cobra.Command, args []string) error {
	scriptPath := args[0]
	src, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("read script %s: %w", scriptPath, err)
	}

	state := &script.DeviceState{
		KVS:     map[string]interface{}{},
		Storage: map[string]interface{}{},
	}
	if deviceStateFile != "" {
		loaded, err := script.LoadDeviceState(options.Log, deviceStateFile)
		if err != nil {
			return fmt.Errorf("load device-state: %w", err)
		}
		state = loaded
	}

	mqtt.SetClient(mqtt.NewMockClient())
	defer mqtt.ResetClient()

	ctx, cancel := context.WithCancel(logr.NewContext(cmd.Context(), options.Log))
	defer cancel()

	if duration > 0 {
		t := time.AfterFunc(duration, cancel)
		defer t.Stop()
	}

	runErr := script.RunWithDeviceState(ctx, scriptPath, src, false, state)
	if runErr != nil && !isContextCancellation(runErr, ctx) {
		return runErr
	}

	if saveState && deviceStateFile != "" {
		if err := script.SaveDeviceState(options.Log, deviceStateFile, state); err != nil {
			return fmt.Errorf("save device-state: %w", err)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(state)
}

func isContextCancellation(err error, ctx context.Context) bool {
	if err == nil {
		return false
	}
	return ctx.Err() != nil
}
