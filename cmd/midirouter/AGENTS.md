# MIDIRouter CLI — cmd structure

```
cmd/midirouter/
├── main.go       # Entry point (calls app.Execute)
└── app/
    ├── root.go   # Root cobra command, shared helpers, subcommand registration
    ├── route.go  # "route" subcommand
    └── list.go   # "list" subcommand
```

## Adding a new subcommand

1. Create `cmd/midirouter/app/<name>.go`.
2. Define a package-level `*cobra.Command` variable with `Use`, `Short`, `Run`.
3. Call `rootCmd.AddCommand(<name>Cmd)` in `init()`.

Example:

```go
// cmd/midirouter/app/foo.go
package app

import (
    "github.com/spf13/cobra"
)

var fooCmd = &cobra.Command{
    Use:   "foo",
    Short: "Do something cool",
    Run: func(_ *cobra.Command, args []string) {
        // ...
    },
}

func init() {
    rootCmd.AddCommand(fooCmd)
}
```

## Patterns to follow

- Keep `main.go` minimal — one line calling `app.Execute()`. No logic there.
- Shared code (flag parsing, port listing, helpers) lives in `root.go`.
- Each subcommand gets its own file. One command variable per file.
- Use package-level variables for commands (`var <name>Cmd`). Cobra expects this pattern.
- If a subcommand needs flags, define them with `pflag` or `cobra.Command.Flags()`.

## Testing

```bash
go build ./cmd/midirouter/
./midirouter list    # list MIDI ports
./midirouter foo     # test new command
```
