package router

import (
	"MIDIRouter/rule"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/youpy/go-coremidi"
)

type MIDIRouter struct {
	sourceDevice      string
	destinationDevice string

	midiClient coremidi.Client
	srcPort    coremidi.InputPort

	destPort    coremidi.OutputPort
	destination coremidi.Destination

	defaultPassThrough bool
	lastMIDIMsg        time.Time
	sendLimit          time.Duration
	rules              []*rule.Rule

	verbose bool
}

func New(sourceDevice string, destinationDevice string) (*MIDIRouter, error) {
	var relay MIDIRouter
	var err error

	relay.sourceDevice = sourceDevice
	relay.destinationDevice = destinationDevice
	relay.defaultPassThrough = false

	relay.midiClient, err = coremidi.NewClient("MIDIRouter")
	if err != nil {
		return nil, err
	}
	err = relay.setupSource()
	if err != nil {
		return nil, err
	}

	err = relay.setupDestination()
	if err != nil {
		return nil, err
	}
	return &relay, nil
}

func (relay *MIDIRouter) SetVerbose(verb bool) {
	relay.verbose = verb
}

func (relay *MIDIRouter) SetPassthrough(pass bool) {
	relay.defaultPassThrough = pass
}

func (relay *MIDIRouter) SetSendLimit(delay time.Duration) {
	relay.sendLimit = delay
}

func (relay *MIDIRouter) Start() {
	for {
		time.Sleep(5 * time.Second)
	}
}

func (relay *MIDIRouter) Cleanup() {
	relay.sendAllNotesOffAndResetControllers()
}

func (relay *MIDIRouter) AddRule(rule *rule.Rule) {
	relay.rules = append(relay.rules, rule)
	fmt.Println(rule)
}

// Method to schedule and send noise packets
func (relay *MIDIRouter) scheduleNoisePacket(packet coremidi.Packet, delayMs time.Duration) {
	// For zero or negative delay, send immediately without a goroutine
	if delayMs <= 0 {
		// Check if we're within the send limit
		if time.Since(relay.lastMIDIMsg) <= relay.sendLimit {
			if relay.verbose {
				fmt.Println("Ignoring noise MIDI message (send limit)")
			}
			return
		}

		if relay.verbose {
			fmt.Printf("Sending noise packet immediately after original message: %v\n",
				hex.EncodeToString(packet.Data))
		}

		// Send the noise packet directly
		packet.Send(&relay.destPort, &relay.destination)
		relay.lastMIDIMsg = time.Now()
		return
	}

	// For positive delays, use a goroutine
	go func() {
		// Use the specified delay
		time.Sleep(delayMs)

		// Check if we're within the send limit
		if time.Since(relay.lastMIDIMsg) <= relay.sendLimit {
			if relay.verbose {
				fmt.Println("Ignoring noise MIDI message (send limit)")
			}
			return
		}

		if relay.verbose {
			fmt.Printf("Sending noise packet after %v delay: %v\n",
				delayMs,
				hex.EncodeToString(packet.Data))
		}

		// Send the noise packet
		packet.Send(&relay.destPort, &relay.destination)
		relay.lastMIDIMsg = time.Now()
	}()
}

func (relay *MIDIRouter) onPacket(source coremidi.Source, packet coremidi.Packet) {
	if relay.verbose {
		fmt.Printf(
			"device: %v, manufacturer: %v, source: %v, data: %v\n",
			source.Entity().Device().Name(),
			source.Manufacturer(),
			source.Name(),
			hex.EncodeToString(packet.Data),
		)
	}

	// if it's a SysEx message, handle it directly without splitting
	if len(packet.Data) > 0 && packet.Data[0] == 0xF0 {
		relay.handleSinglePacket(packet)
		return
	}

	// For non-SysEx messages, proceed with splitting if needed
	if len(packet.Data) > 3 {
		// Only split if packet is longer than max standard MIDI message
		messages := splitMIDIData(packet.Data)
		for _, msg := range messages {
			relay.handleSinglePacket(coremidi.Packet{Data: msg})
		}
	} else {
		// Single short message - process directly
		relay.handleSinglePacket(packet)
	}
}

func (relay *MIDIRouter) handleSinglePacket(packet coremidi.Packet) {
	if relay.defaultPassThrough == true {
		if time.Since(relay.lastMIDIMsg) <= relay.sendLimit {
			fmt.Println("Ignoring midi message (send limit)")
			return
		}
		packet.Send(&relay.destPort, &relay.destination)

		if len(packet.Data) > 0 && packet.Data[0] == 0xFC { // Stop message
			relay.sendAllNotesOffAndResetControllers()
		}

		relay.lastMIDIMsg = time.Now()
		return
	}

	ruleMatched := false
	for _, r := range relay.rules {
		if len(packet.Data) == 0 {
			continue
		}

		// Get match result from rule
		matchResult := r.Match(packet, relay.verbose)

		if matchResult.Result == rule.RuleMatchResultMatchInject {
			if relay.verbose {
				fmt.Println("-> Sending generated packet :")
				fmt.Println(hex.Dump(matchResult.MainPacket.Data))
			}

			if time.Since(relay.lastMIDIMsg) <= relay.sendLimit {
				fmt.Println("Ignoring midi message (send limit)")
				return
			}

			// Send the main packet
			matchResult.MainPacket.Send(&relay.destPort, &relay.destination)
			relay.lastMIDIMsg = time.Now()

			// Handle noise packet if present
			if matchResult.NoisePacket != nil {
				// Schedule/send noise packet after the main packet is sent
				relay.scheduleNoisePacket(*matchResult.NoisePacket, matchResult.NoiseDelayMs)
			}

			ruleMatched = true
			break
		} else if matchResult.Result == rule.RuleMatchResultMatchNoInject {
			ruleMatched = true
			break
		}
	}

	if (ruleMatched == false) && (relay.verbose == true) {
		fmt.Println("-> No match")
	}
}

func splitMIDIData(data []byte) [][]byte {
	var messages [][]byte
	for i := 0; i < len(data); {
		status := data[i]
		length := midiMessageLength(status)
		if i+length > len(data) {
			break
		}
		messages = append(messages, data[i:i+length])
		i += length
	}
	return messages
}

func midiMessageLength(status byte) int {
	switch status & 0xF0 {
	case 0x80, 0x90, 0xA0, 0xB0, 0xE0:
		return 3
	case 0xC0, 0xD0:
		return 2
	case 0xF0:
		switch status {
		case 0xF1, 0xF3:
			return 2
		case 0xF2:
			return 3
		default:
			return 1
		}
	default:
		return 1
	}
}

func (relay *MIDIRouter) sendAllNotesOffAndResetControllers() {
	for ch := 0; ch < 16; ch++ {
		// All notes off
		packet := coremidi.Packet{Data: []byte{0xB0 | byte(ch), 123, 0}}
		packet.Send(&relay.destPort, &relay.destination)

		// Reset all controllers
		packet = coremidi.Packet{Data: []byte{0xB0 | byte(ch), 121, 0}}
		packet.Send(&relay.destPort, &relay.destination)
	}
}
