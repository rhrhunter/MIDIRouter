package main

import (
	"MIDIRouter/config"
	"fmt"
	"os"
	"sync"

	"github.com/youpy/go-coremidi"
)

const (
	version = "1.2"
)

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

	var wg sync.WaitGroup

	for _, configFile := range os.Args[1:] {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			router, err := config.LoadConfig(file)
			if err != nil {
				fmt.Printf("Error loading config %s: %v\n", file, err)
				return
			}
			router.Start()
		}(configFile)
	}

	wg.Wait()
}
