# Contributing to MIDIRouter

## Development setup

```bash
go mod tidy        # fetch dependencies
make build         # compile
./bin/midirouter   # run
```

Requires Go 1.23+. macOS only (uses `github.com/youpy/go-coremidi`).

## Make targets

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `bin/midirouter` |
| `make test` | Run tests |
| `make fmt` | Format code (`go fmt`) |
| `make vet` | Static analysis (`go vet`) |
| `make check` | Run `fmt` + `vet` + `tidy` |
| `make clean` | Remove build artifacts |
| `make install` | Install binary to `~/go/bin` |
| `make help` | List all targets |

## CLI structure

The CLI uses Cobra. Each subcommand lives in its own file under `cmd/midirouter/app/`:

```
cmd/midirouter/
├── main.go       # Entry point — calls app.Execute()
└── app/
    ├── root.go   # Root command, shared helpers, flag setup
    ├── route.go  # "route" subcommand
    └── list.go   # "list" subcommand
```

### Adding a subcommand

1. Create `cmd/midirouter/app/<name>.go`.
2. Define a package-level cobra command:

```go
package app

import "github.com/spf13/cobra"

var myCmd = &cobra.Command{
    Use:   "my",
    Short: "Do something",
    RunE: func(_ *cobra.Command, args []string) error {
        // ...
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

3. Register it in `init()` via `rootCmd.AddCommand(<name>Cmd)`.

Keep `main.go` to a single line (`app.Execute()`). All command logic belongs in the `app/` package.

## Rule pipeline architecture

The core routing pipeline is: **Source Device → Rules (filter → transform → generate) → Destination Device**.

Extending the engine means adding new filter/generator packages in the top-level directory structure:

```
filternoteon/     # Note On matching logic
gennoteon/        # Note On generation logic
gencontrolchange/ # Control Change generation logic
# ... etc
```

Each must implement its respective interface (`filterinterface.Filter`, `generatorinterface.Generator`). Register the new type in `config/config.go` under the relevant switch statements.

## Code style

- Run `make check` before committing (`fmt` + `vet` + `tidy`).
- Follow standard Go conventions: godoc comments on exported symbols, one variable per line, explicit error handling.
- Keep functions focused. Each rule component (filter/generator) is its own package with a single public constructor (`New`).

## Testing

```bash
make test
make test-coverage
make coverage        # HTML report in coverage.out.html
```
