# Contributing to MIDIRouter

## Setup

```bash
go mod tidy        # fetch dependencies
make build         # compile
./bin/midirouter   # run
```

## Make targets

| Command | Description |
|---------|-------------|
| `make build` | Build the binary to `bin/midirouter` |
| `make test` | Run all tests |
| `make fmt` | Format code with gofmt |
| `make vet` | Run go vet for static analysis |
| `make check` | Run fmt + vet + tidy |
| `make clean` | Remove build artifacts |
| `make install` | Install binary to ~/go/bin |
| `make help` | List all make targets |

## Adding a new subcommand

1. Create `cmd/midirouter/app/<name>.go`.
2. Define `var <name>Cmd = &cobra.Command{...}` with `Use`, `Short`, and `Run`.
3. Call `rootCmd.AddCommand(<name>Cmd)` in `init()`.

```go
// cmd/midirouter/app/foo.go
package app

import "github.com/spf13/cobra"

var fooCmd = &cobra.Command{
    Use:   "foo",
    Short: "Do something",
    Run: func(_ *cobra.Command, args []string) {
        // ...
    },
}

func init() {
    rootCmd.AddCommand(fooCmd)
}
```

Keep `main.go` minimal — one line calling `app.Execute()`. No logic there.

## Building for macOS

```bash
make build        # default (arm64 on Apple Silicon)
make build-macos  # arm64 (default)
```

## Code style

Run `make fmt vet check` before committing. Go conventions apply: `gofmt`, godoc comments on exported symbols, one variable per line.
