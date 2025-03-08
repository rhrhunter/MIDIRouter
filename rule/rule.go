package rule

import (
	"MIDIRouter/filter"
	"MIDIRouter/filterinterface"
	"MIDIRouter/generatorinterface"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/youpy/go-coremidi"
)

type RuleMatchResult int

const (
	RuleMatchResultNoMatch       = iota
	RuleMatchResultMatchInject   = iota
	RuleMatchResultMatchNoInject = iota
)

type TransformMode uint8

const (
	TransformModeNone       = iota
	TransformModeLinear     = iota
	TransformModeLinearDrop = iota
	TransformModeNoise      = iota // New transform mode for noise
	/*
		Transform7bitsTo100   = iota // Convert a 7bits [0-127] value to a displayed [0-100] value
		Transform14bitsTo100  = iota // Convert a 14bits [0-16383] value to a displayed [0-100] value
		Transform7bitsTo1000  = iota // Convert a 7bits [0-127] value to a displayed [0-1000] value
		Transform14bitsTo1000 = iota // Convert a 14bits [0-16383] value to a displayed [0-1000] value
		Transform7bitsTo127   = iota // Convert a 7bits [0-127] value to a [0-127] value
		Transform14bitsTo127  = iota // Convert a 14bits [0-16383] value to a [0-127] value*/
)

// Define a new NoiseSettings struct
type NoiseSettings struct {
	MsgType    filter.FilterMsgType // MIDI message type for noise
	Channel    filter.FilterChannel // MIDI channel for noise
	MinValue   uint8                // Minimum noise value
	MaxValue   uint8                // Maximum noise value
	DelayMsMin uint16               // Minimum delay in milliseconds
	DelayMsMax uint16               // Maximum delay in milliseconds
}

type Transform struct {
	mode          TransformMode
	fromMin       uint32
	fromMax       uint32
	toMin         uint32
	toMax         uint32
	noiseSettings NoiseSettings // New field for noise settings
}

type Rule struct {
	name                  string
	filter                filterinterface.FilterInterface
	transform             Transform
	dropDuplicates        bool
	dropDuplicatesTimeout time.Duration

	generator generatorinterface.GeneratorInterface

	lastValue   uint16
	lastValueTs time.Time
}

func New(ruleName string) (*Rule, error) {
	var r Rule

	r.name = ruleName
	r.lastValue = 0xFFFF
	r.transform.mode = TransformModeNone
	return &r, nil
}

func (r *Rule) SetTransform(mode TransformMode, fromMin uint32, fromMax uint32, toMin uint32, toMax uint32) {
	r.transform = Transform{
		mode:    mode,
		fromMin: fromMin,
		fromMax: fromMax,
		toMin:   toMin,
		toMax:   toMax,
	}
}

// Add a method to set noise settings
func (r *Rule) SetNoiseSettings(noiseSettings NoiseSettings) {
	r.transform.noiseSettings = noiseSettings
}

func (r *Rule) SetFilter(f filterinterface.FilterInterface) error {
	if r.filter != nil {
		return errors.New("Filter already set")
	}
	r.filter = f
	return nil
}

func (r *Rule) EnableDropDuplicates(enable bool, timeout time.Duration) {
	r.dropDuplicates = enable
	r.dropDuplicatesTimeout = timeout
}

func (r *Rule) SetGenerator(g generatorinterface.GeneratorInterface) error {
	if r.generator != nil {
		return errors.New("Generator already set")
	}
	r.generator = g
	return nil
}

// Add new function to generate a noise packet
func (r *Rule) generateNoisePacket(packet coremidi.Packet, value uint16) coremidi.Packet {
	// Get random values for noise
	ns := r.transform.noiseSettings

	// Generate random value between MinValue and MaxValue
	randVal := ns.MinValue
	if ns.MaxValue > ns.MinValue {
		randVal = ns.MinValue + uint8(rand.Intn(int(ns.MaxValue-ns.MinValue+1)))
	}

	// Create status byte based on message type and channel
	// Fix: Properly convert to byte with appropriate casting
	msgType := byte(ns.MsgType)
	channel := byte(ns.Channel)
	statusByte := byte((msgType << 4) | channel)

	// Create MIDI message based on message type
	var data []byte
	switch ns.MsgType {
	case filter.FilterMsgTypeNoteOn, filter.FilterMsgTypeNoteOff, filter.FilterMsgTypeAftertouch,
		filter.FilterMsgTypeControlChange, filter.FilterMsgTypePitchWheel:
		// Two data bytes (e.g., note/control number and velocity/value)
		data = []byte{statusByte, byte(value & 0x7F), randVal}
	case filter.FilterMsgTypeProgramChange, filter.FilterMsgTypeChannelPressure:
		// One data byte (e.g., program number or pressure value)
		data = []byte{statusByte, randVal}
	default:
		// Default to a simple message with just the random value
		data = []byte{statusByte, randVal}
	}

	return coremidi.NewPacket(data, packet.TimeStamp)
}

// Update the Match method to handle noise
func (r *Rule) Match(packet coremidi.Packet, verbose bool, router interface{}) (RuleMatchResult, coremidi.Packet) {
	msgType := filter.FilterMsgType((packet.Data[0] & 0xF0) >> 4)
	channel := filter.FilterChannel(packet.Data[0] & 0x0F)
	if r.filter.QuickMatch(msgType, channel) == false {
		return RuleMatchResultNoMatch, packet
	}

	result, value := r.filter.Match(packet)
	if result == filterinterface.FilterMatchResult_NoMatch {
		return RuleMatchResultNoMatch, packet
	}

	if result == filterinterface.FilterMatchResult_MatchNoValue {
		if verbose {
			fmt.Println("Filter match (no value)")
		}
		return RuleMatchResultMatchNoInject, packet
	}
	if result != filterinterface.FilterMatchResult_Match {
		return RuleMatchResultNoMatch, packet
	}
	if verbose {
		fmt.Println("Filter", r.String(), "matched. Extracted value:", value)
		fmt.Println("-> Extracted value:", value)
	}

	// Transform the value based on transform mode
	transformedValue := value
	switch r.transform.mode {
	case TransformModeLinear:
		// Linear transformation logic
		a := float64(r.transform.toMax-r.transform.toMin) / float64(r.transform.fromMax-r.transform.fromMin)
		b := float64(r.transform.toMin) - a*float64(r.transform.fromMin)
		transformedValue = uint16(a*float64(value) + float64(b))
	case TransformModeLinearDrop:
		// Check bounds and apply linear transformation
		if (uint32(value) > r.transform.fromMax) || (uint32(value) < r.transform.fromMin) {
			fmt.Println("-> Transform dropped out of bounds input value")
			return RuleMatchResultNoMatch, packet
		}
		a := float64(r.transform.toMax-r.transform.toMin) / float64(r.transform.fromMax-r.transform.fromMin)
		b := float64(r.transform.toMin) - a*float64(r.transform.fromMin)
		v := uint16(a*float64(value) + float64(b))
		if (uint32(v) > r.transform.toMax) || (uint32(v) < r.transform.toMin) {
			fmt.Println("-> Transform dropped out of bounds output value")
			return RuleMatchResultNoMatch, packet
		}
		transformedValue = v
	case TransformModeNoise:
		// Apply normal linear transformation first
		a := float64(r.transform.toMax-r.transform.toMin) / float64(r.transform.fromMax-r.transform.fromMin)
		b := float64(r.transform.toMin) - a*float64(r.transform.fromMin)
		transformedValue = uint16(a*float64(value) + float64(b))

		// Generate and schedule noise packet to be sent after the main packet
		if verbose {
			fmt.Println("-> Generating noise packet")
		}
		noisePacket := r.generateNoisePacket(packet, value)

		// Schedule noise packet to be sent after a random delay
		ns := r.transform.noiseSettings
		delayMs := ns.DelayMsMin
		if ns.DelayMsMax > ns.DelayMsMin {
			delayMs = ns.DelayMsMin + uint16(rand.Intn(int(ns.DelayMsMax-ns.DelayMsMin+1)))
		}

		// Use the router to send the noise packet with delay
		if midiRouter, ok := router.(MIDIRouter); ok {
			midiRouter.SendNoisePacket(noisePacket, time.Duration(delayMs)*time.Millisecond)
			if verbose {
				fmt.Println("-> Scheduled noise packet with delay", delayMs, "ms")
			}
		} else if verbose {
			fmt.Println("-> Cannot send noise packet: router not available")
		}
	}

	if verbose {
		fmt.Println("-> Transformed value:", transformedValue)
	}

	// Apply duplicate check
	if r.dropDuplicates && (r.lastValue == transformedValue) && (time.Since(r.lastValueTs) < r.dropDuplicatesTimeout) {
		fmt.Println("-> Ignored duplicate")
		return RuleMatchResultMatchNoInject, packet
	}
	r.lastValue = transformedValue
	r.lastValueTs = time.Now()

	// Generate output
	newPacket, err := r.output(packet, transformedValue)
	if err != nil {
		fmt.Println(err)
		return RuleMatchResultMatchInject, packet
	} else {
		return RuleMatchResultMatchInject, newPacket
	}
}

func (r *Rule) output(packet coremidi.Packet, value uint16) (newPacket coremidi.Packet, err error) {
	newPacket, err = r.generator.Generate(packet, value)
	if err != nil {
		return packet, err
	}

	return newPacket, nil
}

func (r Rule) String() string {
	var str string
	str += "***** Rule '" + r.name + "' *****\n"
	str += "  Match    : " + r.filter.String() + "\n"
	str += "  Transform: " + r.transform.String() + "\n"
	str += "  Output   : " + r.generator.String()

	return str
}

func (t Transform) String() string {
	switch t.mode {
	case TransformModeNone:
		return "None"
	case TransformModeLinear:
		return fmt.Sprintf("Linear from [%d, %d] to [%d, %d]", t.fromMin, t.fromMax, t.toMin, t.toMax)
	case TransformModeLinearDrop:
		return fmt.Sprintf("Linear from [%d, %d] to [%d, %d] (drop out of range values)", t.fromMin, t.fromMax, t.toMin, t.toMax)
	case TransformModeNoise:
		return fmt.Sprintf("Noise from [%d, %d] to [%d, %d] with noise (channel %s, msgType %s, value range [%d, %d], delay [%d, %d]ms)",
			t.fromMin, t.fromMax, t.toMin, t.toMax,
			t.noiseSettings.Channel.String(), t.noiseSettings.MsgType.String(),
			t.noiseSettings.MinValue, t.noiseSettings.MaxValue,
			t.noiseSettings.DelayMsMin, t.noiseSettings.DelayMsMax)
	}
	return "?"
}

// Define MIDIRouter interface for use in Match
type MIDIRouter interface {
	SendNoisePacket(packet coremidi.Packet, delayMs time.Duration)
}
