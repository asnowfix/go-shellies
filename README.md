# go-shellies

Go library and CLI for [Shelly](https://www.shelly.com/) Gen2+ devices.

Originally extracted from [home-automation](https://github.com/asnowfix/home-automation)
via `git filter-repo` — full per-file history is preserved (`git log --follow` works).

## Layout

Three Go modules:

- `github.com/asnowfix/go-shellies` — the library: device, registrar, mDNS,
  per-component packages (`ble`, `blu`, `kvs`, `mqtt`, `schedule`, `sswitch`,
  `system`, `wifi`, ...) and the `shelly` CLI binary under `cmd/shelly/`.
- `github.com/asnowfix/go-shellies/script` — goja-based JS runtime for
  emulating Shelly device-side scripts (heavy `goja`+`minify` deps; isolated).
- `github.com/asnowfix/go-shellies/gen1` — gen1 device support
  (`gorilla/schema` dep; isolated).

A `go.work` file ties the three together for local development.

## CLI

```sh
go install github.com/asnowfix/go-shellies/cmd/shelly@latest
shelly --help
```

The `shelly` binary discovers devices on the local network via mDNS by default,
or addresses them by MQTT topic id with `--via mqtt`. It does **not** depend on
any home-automation daemon.

Subcommands:

| Subcommand | What it does |
|---|---|
| `call <device> <method> [params-json]` | Direct RPC call |
| `kvs get / set / delete` | Key-Value Store ops |
| `wifi config / status / scan / list-ap-clients` | WiFi config & status |
| `mqtt config / status` | MQTT config & status (`--mqtt-broker URL` for `config`) |
| `sys config / reboot` | System config & reboot |
| `status` | Component statuses |
| `reboot` | Reboot a device |
| `components` | List components |
| `jobs show / cancel / schedule` | Schedule jobs |
| `mcp` | Run an MCP stdio server exposing `shelly_list` and `shelly_call` |
| `emulate <script.js>` | Run a Shelly JS script in the goja runtime locally |

### Examples

```sh
# Discover and list status for every shellyplus device
shelly status 'shellyplus*'

# Call an RPC method directly
shelly call shellyplusht-d4b9f4 Switch.Set '{"id":0,"on":true}'

# Run a JS script locally with an initial KVS snapshot
shelly emulate ./pool-pump.js --device-state ./pool-state.json --duration 30s
```

## Library use

```go
import (
    shelly "github.com/asnowfix/go-shellies"
    "github.com/asnowfix/go-shellies/mqtt"
    "github.com/asnowfix/go-shellies/script"
)

// Optional extras: pull in the script package, embed your scripts FS.
script.SetFS(myEmbed.GetFS())
shelly.Init(log, mqttClient, timeout, rateLimit, script.Init)
```

## Build

```sh
go build ./...
go test ./...
```

(or, equivalently, from each sub-module directory).

## License

GPL-3.0 — see `LICENSE`.
