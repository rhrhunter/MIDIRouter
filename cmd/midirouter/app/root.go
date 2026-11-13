package app

import (
	"fmt"
	"os"

	"github.com/youpy/go-coremidi"

	"github.com/spf13/cobra"
)

const version = "1.2"

var rootCmd = &cobra.Command{
	Use:   "midirouter",
	Short: "MIDI routing engine on macOS",
	Long: `MIDIRouter bridges CoreMIDI source devices to destination devices
using JSON configuration files.

Usage:
  midirouter [command]

Available Commands:
  list        List available MIDI input and output ports
  route       Start MIDI route(s) from config file(s)
  version     Print the version number

Flags:
  -h, --help   Help for midirouter`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("MIDIRouter v%s\n", version)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available MIDI ports",
	Run: func(_ *cobra.Command, _ []string) {
		listPorts()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(routeCmd)
	rootCmd.AddCommand(listCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func listPorts() {
	sources, err := coremidi.AllSources()
	if err != nil {
		fmt.Println("Error listing sources:", err)
		return
	}
	fmt.Println("MIDI inputs:")
	for _, source := range sources {
		fmt.Printf("  %s/%s/%s\n",
			source.Entity().Device().Name(),
			source.Manufacturer(),
			source.Name())
	}

	destinations, err := coremidi.AllDestinations()
	if err != nil {
		fmt.Println("Error listing destinations:", err)
		return
	}
	fmt.Println("MIDI outputs:")
	for _, destination := range destinations {
		fmt.Printf("  %s/%s\n",
			destination.Manufacturer(),
			destination.Name())
	}
}
