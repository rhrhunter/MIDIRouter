package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"MIDIRouter/config"
	"MIDIRouter/router"

	"github.com/spf13/cobra"
)

var routeCmd = &cobra.Command{
	Use:   "route [config files...]",
	Short: "Start MIDI route(s) from config file(s)",
	Long: `Start one or more MIDI routes from JSON configuration files.
Each config file defines a single route (source → destination).
Multiple configs are started concurrently, one per goroutine.

Examples:
  midirouter route config/passthrough.json
  midirouter route config-a.json config-b.json`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		var routers []*router.MIDIRouter

		for _, configFile := range args {
			go func(cfg string) {
				r, err := config.LoadConfig(cfg)
				if err != nil {
					fmt.Printf("Error loading config %s: %v\n", cfg, err)
					return
				}
				routers = append(routers, r)
				r.Start()
			}(configFile)
		}

		<-sigchan
		for _, r := range routers {
			r.Cleanup()
		}
	},
}
