package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/youpy/go-coremidi"

	"MIDIRouter/config"
	"MIDIRouter/router"
)

const (
	version = "1.2"
)

var routers []*router.MIDIRouter

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("MIDIRouter v%s\n", version)
		fmt.Println("Usage:", os.Args[0], "<config file 1> [config file 2] ...")
		fmt.Println("MIDI inputs:")
		sources, err := coremidi.AllSources()
		if err != nil {
			panic(err)
		}
		for _, source := range sources {
			fmt.Println("  " + source.Entity().Device().Name() + "/" + source.Manufacturer() + "/" + source.Name())
		}

		fmt.Println("MIDI outputs:")
		destinations, err := coremidi.AllDestinations()
		if err != nil {
			panic(err)
		}
		for _, destination := range destinations {
			fmt.Println("  " + destination.Manufacturer() + "/" + destination.Name())
		}

		return
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for _, configFile := range os.Args[1:] {
			go startRouter(configFile)
		}
	}()

	<-sigchan
	for _, router := range routers {
		router.Cleanup()
	}
}

func startRouter(file string) {
	router, err := config.LoadConfig(file)
	if err != nil {
		fmt.Printf("Error loading config %s: %v\n", file, err)
		return
	}
	routers = append(routers, router)
	router.Start()
}
