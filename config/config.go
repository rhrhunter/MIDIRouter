package config

import (
	"MIDIRouter/filter"
	"MIDIRouter/filteraftertouch"
	"MIDIRouter/filterchannelpressure"
	"MIDIRouter/filtercontrolchange"
	"MIDIRouter/filternoteoff"
	"MIDIRouter/filternoteon"
	"MIDIRouter/filterpitchwheel"
	"MIDIRouter/filterprogramchange"

	"MIDIRouter/genaftertouch"
	"MIDIRouter/genchannelpressure"
	"MIDIRouter/gencontrolchange"
	"MIDIRouter/gennoteoff"
	"MIDIRouter/gennoteon"
	"MIDIRouter/genpitchwheel"
	"MIDIRouter/genprogramchange"
	"MIDIRouter/gensysex"

	"MIDIRouter/router"
	"MIDIRouter/rule"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"
)

type RouterConfig struct {
	SourceDevice       string
	DestinationDevice  string
	DefaultPassthrough bool
	SendLimitMs        int
	Verbose            bool
	Rules              []RuleConfig
}

type RuleConfig struct {
	Name      string
	Filter    FilterConfig
	Transform TransformConfig
	Generator GeneratorConfig
}

// Example: "program change 52" => 0xC0 0x34 => [0xC=PgmChange | 0x0 : Channel 0 | 0x34 : 52]
type FilterConfig struct {
	MsgType string //Note On, Note Off, Aftertouch, Control Change..
	Channel string // 4bits or '*'

	Settings json.RawMessage
}

// Update the TransformConfig struct to include noise settings
type TransformConfig struct {
	FromMin       int
	FromMax       int
	ToMin         int
	ToMax         int
	Mode          string
	NoiseSettings NoiseSettingsConfig `json:"NoiseSettings,omitempty"`
	// No additional settings needed for PreventRunningStatus
}

// Add a new struct for noise settings in config
type NoiseSettingsConfig struct {
	MsgType    string `json:"MsgType"`
	Channel    string `json:"Channel"`
	MinValue   int    `json:"MinValue"`
	MaxValue   int    `json:"MaxValue"`
	DelayMsMin int    `json:"DelayMsMin"`
	DelayMsMax int    `json:"DelayMsMax"`
}

type GeneratorConfig struct {
	Name string

	MsgType                 string //Note On, Note Off, Aftertouch, Control Change..
	Channel                 string // 4bits or '*'
	DropDuplicates          bool
	DropDuplicatesTimeoutMs int
	Settings                json.RawMessage
}

func LoadConfig(configPath string) (*router.MIDIRouter, error) {
	var config RouterConfig
	var relay *router.MIDIRouter

	config.Verbose = false
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.New("Failed parsing config file: " + err.Error())
	}
	if len(config.SourceDevice) == 0 {
		return nil, errors.New("MIDI source cannot be empty")
	}
	if len(config.DestinationDevice) == 0 {
		return nil, errors.New("MIDI destination cannot be empty")
	}
	if config.SourceDevice == config.DestinationDevice {
		return nil, errors.New("MIDI source and destination cannot identical")
	}

	relay, err = router.New(config.SourceDevice, config.DestinationDevice)
	if err != nil {
		return nil, err
	}

	relay.SetVerbose(config.Verbose)
	relay.SetPassthrough(config.DefaultPassthrough)
	relay.SetSendLimit(time.Duration(config.SendLimitMs) * time.Millisecond)

	for _, r := range config.Rules {
		newRule, _ := rule.New(r.Name)

		//Load input filter from config
		filterMsgType, err := stringToMsgType(r.Filter.MsgType)
		if err != nil {
			return nil, err
		}
		ruleChannel, err := stringToFilterChannel(r.Filter.Channel)
		if err != nil {
			return nil, errors.New("Invalid channel " + err.Error())
		}
		fmt.Println("Loading rule '" + r.Name + "'...")

		switch filterMsgType {
		case filter.FilterMsgTypeNoteOn:
			f, err := filternoteon.New(ruleChannel, r.Filter.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetFilter(f)
			break
		case filter.FilterMsgTypeNoteOff:
			f, err := filternoteoff.New(ruleChannel, r.Filter.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetFilter(f)
			break
		case filter.FilterMsgTypeAftertouch:
			f, err := filteraftertouch.New(ruleChannel, r.Filter.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetFilter(f)
			break
		case filter.FilterMsgTypeControlChange:
			f, err := filtercontrolchange.New(ruleChannel, r.Filter.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetFilter(f)
			break
		case filter.FilterMsgTypeProgramChange:
			f, err := filterprogramchange.New(ruleChannel, r.Filter.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetFilter(f)
			break
		case filter.FilterMsgTypeChannelPressure:
			f, err := filterchannelpressure.New(ruleChannel, r.Filter.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetFilter(f)
			break
		case filter.FilterMsgTypePitchWheel:
			f, err := filterpitchwheel.New(ruleChannel, r.Filter.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetFilter(f)
			break
		default:
			return nil, errors.New("Failed to add rule, invalid filter type: " + r.Filter.MsgType)
		}

		//Load Transform
		transformMode, err := stringToTransformMode(r.Transform.Mode)
		if err != nil {
			return nil, err
		}
		if transformMode != rule.TransformModeNone {
			newRule.SetTransform(
				transformMode,
				uint32(r.Transform.FromMin),
				uint32(r.Transform.FromMax),
				uint32(r.Transform.ToMin),
				uint32(r.Transform.ToMax),
			)

			// Handle noise settings if mode is Noise
			if transformMode == rule.TransformModeNoise {
				// Parse MsgType
				noiseMsgType, err := stringToMsgType(r.Transform.NoiseSettings.MsgType)
				if err != nil {
					return nil, errors.New("Invalid noise message type: " + err.Error())
				}

				// Parse Channel
				noiseChannel, err := stringToFilterChannel(r.Transform.NoiseSettings.Channel)
				if err != nil {
					return nil, errors.New("Invalid noise channel: " + err.Error())
				}

				// Validate value ranges
				if r.Transform.NoiseSettings.MaxValue > 127 {
					return nil, errors.New("Noise MaxValue exceeds MIDI limit of 127")
				}

				// Create NoiseSettings struct
				noiseSettings := rule.NoiseSettings{
					MsgType:    noiseMsgType,
					Channel:    noiseChannel,
					MinValue:   uint8(r.Transform.NoiseSettings.MinValue),
					MaxValue:   uint8(r.Transform.NoiseSettings.MaxValue),
					DelayMsMin: uint16(r.Transform.NoiseSettings.DelayMsMin),
					DelayMsMax: uint16(r.Transform.NoiseSettings.DelayMsMax),
				}

				// Set noise settings on the rule
				newRule.SetNoiseSettings(noiseSettings)
			}
			// PreventRunningStatus doesn't need additional settings
		}

		//Drop consecutive identical values?
		newRule.EnableDropDuplicates(r.Generator.DropDuplicates, time.Duration(time.Duration(r.Generator.DropDuplicatesTimeoutMs)*time.Millisecond))

		//Load Generator
		generateMsgType, err := stringToMsgType(r.Generator.MsgType)
		if err != nil {
			return nil, err
		}
		generatorChannel, err := stringToFilterChannel(r.Generator.Channel)
		if (err != nil) && (generateMsgType != filter.FilterMsgTypeSysEx) {
			fmt.Println(generateMsgType)
			return nil, errors.New("Invalid channel " + err.Error())
		}

		switch generateMsgType {
		case filter.FilterMsgTypeNoteOn:
			g, err := gennoteon.New(generatorChannel, r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		case filter.FilterMsgTypeNoteOff:
			g, err := gennoteoff.New(generatorChannel, r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		case filter.FilterMsgTypeAftertouch:
			g, err := genaftertouch.New(generatorChannel, r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		case filter.FilterMsgTypeChannelPressure:
			g, err := genchannelpressure.New(generatorChannel, r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		case filter.FilterMsgTypeControlChange:
			g, err := gencontrolchange.New(generatorChannel, r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		case filter.FilterMsgTypeProgramChange:
			g, err := genprogramchange.New(generatorChannel, r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		case filter.FilterMsgTypePitchWheel:
			g, err := genpitchwheel.New(generatorChannel, r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		case filter.FilterMsgTypeSysEx:
			g, err := gensysex.New(r.Generator.Settings)
			if err != nil {
				return nil, err
			}
			newRule.SetGenerator(g)
		default:
			return nil, errors.New("Failed to add rule, invalid generate type.")
		}

		relay.AddRule(newRule)
	}

	return relay, nil
}

// Update the stringToTransformMode function to handle the new mode
func stringToTransformMode(str string) (rule.TransformMode, error) {
	switch str {
	case "":
		return rule.TransformModeNone, nil
	case "None":
		return rule.TransformModeNone, nil
	case "Linear":
		return rule.TransformModeLinear, nil
	case "LinearDrop":
		return rule.TransformModeLinearDrop, nil
	case "Noise":
		return rule.TransformModeNoise, nil
	case "PreventRunningStatus":
		return rule.TransformModePreventRunStatus, nil
	default:
		return rule.TransformModeNone, errors.New("Invalid transform mode: " + str)
	}
}

func stringToFilterChannel(str string) (filter.FilterChannel, error) {
	switch str {
	case "1":
		return filter.FilterChannel1, nil
	case "2":
		return filter.FilterChannel2, nil
	case "3":
		return filter.FilterChannel3, nil
	case "4":
		return filter.FilterChannel4, nil
	case "5":
		return filter.FilterChannel5, nil
	case "6":
		return filter.FilterChannel6, nil
	case "7":
		return filter.FilterChannel7, nil
	case "8":
		return filter.FilterChannel8, nil
	case "9":
		return filter.FilterChannel9, nil
	case "10":
		return filter.FilterChannel10, nil
	case "11":
		return filter.FilterChannel11, nil
	case "12":
		return filter.FilterChannel12, nil
	case "13":
		return filter.FilterChannel13, nil
	case "14":
		return filter.FilterChannel14, nil
	case "15":
		return filter.FilterChannel15, nil
	case "16":
		return filter.FilterChannel16, nil
	case "*":
		return filter.FilterChannelAny, nil
	}
	return filter.FilterChannelAny, errors.New("Invalid MIDI channel value: '" + str + "'")
}

func stringToMsgType(str string) (filter.FilterMsgType, error) {
	switch str {
	case "Note On":
		return filter.FilterMsgTypeNoteOn, nil
	case "Note Off":
		return filter.FilterMsgTypeNoteOff, nil
	case "Aftertouch":
		return filter.FilterMsgTypeAftertouch, nil
	case "Control Change":
		return filter.FilterMsgTypeControlChange, nil
	case "Program Change":
		return filter.FilterMsgTypeProgramChange, nil
	case "Channel Pressure":
		return filter.FilterMsgTypeChannelPressure, nil
	case "Pitch Wheel":
		return filter.FilterMsgTypePitchWheel, nil
	case "SysEx":
		return filter.FilterMsgTypeSysEx, nil
	case "*":
		return filter.FilterMsgTypeAny, nil
	default:
		return filter.FilterMsgTypeUnknown, errors.New("Invalid message type: " + str)
	}
}
