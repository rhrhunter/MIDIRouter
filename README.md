# MIDIRouter

MIDIRouter is a MIDI routing engine for macOS that bridges input devices to output devices using JSON configuration files. It processes MIDI messages through a pipeline of filters, transforms, and generators — enabling channel conversion, value remapping, message type transformation, and selective forwarding.

## Features

- Bridge any CoreMIDI source to any destination
- Filter messages by type, channel, and value
- Remap values linearly (e.g., 7-bit to 14-bit)
- Transform message types (Note On → Program Change, etc.)
- Route multiple independent configurations simultaneously
- Clean shutdown with All Notes Off + controller reset on all 16 channels

## Quick Start

### Install

Build from source:

```bash
make build
sudo cp bin/midirouter /usr/local/bin/
```

### List available devices

```bash
midirouter list
```

This outputs all connected MIDI input and output devices. Use the displayed device names in your configuration.

### Create a route

Write a JSON config file:

**Passthrough route** (bridge two devices unchanged):

```json
{
    "SourceDevice": "Interface/Port 1",
    "DestinationDevice": "Synth/Channel 1",
    "DefaultPassthrough": true
}
```

**Rule-based route** (filter and transform messages):

```json
{
    "SourceDevice": "Controller/Port 1",
    "DestinationDevice": "Synth/Channel 1",
    "DefaultPassthrough": false,
    "Rules": [
        {
            "Name": "CC 74 remap: 0–127 to 0–16383",
            "Filter": {
                "MsgType": "Control Change",
                "Channel": "1",
                "Settings": {
                    "ControllerNumber": "74",
                    "Value": "*"
                }
            },
            "Transform": {
                "Mode": "Linear",
                "FromMin": 0,
                "FromMax": 127,
                "ToMin": 0,
                "ToMax": 16383
            },
            "Generator": {
                "MsgType": "Control Change",
                "Channel": "1",
                "Settings": {
                    "ControllerNumber": "74",
                    "Value": "$"
                }
            }
        }
    ]
}
```

### Start a route

```bash
# Single route
midirouter route config.json

# Multiple simultaneous routes
midirouter route route-a.json route-b.json
```

Press `Ctrl+C` to stop all routes cleanly.

## Configuration Reference

### Top-level fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `SourceDevice` | string | **required** | Source port name (e.g., `"Interface/Port 1"`) |
| `DestinationDevice` | string | **required** | Destination port name (e.g., `"Synth/Channel 1"`) |
| `DefaultPassthrough` | bool | `false` | Forward unmatched messages unchanged |
| `Verbose` | bool | `false` | Log every routed packet to stdout |
| `SendLimitMs` | int | — | Minimum interval between output messages (milliseconds) |
| `Rules` | array | — | Filter → Transform → Generator pipeline rules |

### Rule structure

Each rule processes a single message type. Rules are evaluated in order; the first match wins.

A rule has three components:

1. **Filter** — matches incoming MIDI messages by type, channel, and value fields
2. **Transform** (optional) — remaps matched values using Linear or LinearDrop mode
3. **Generator** — creates the outgoing message, optionally reusing original (`*`) or transformed (`$`) values

### Message types

Note On, Note Off, Control Change, Program Change, Pitch Wheel, Channel Pressure, Aftertouch, SysEx.

### Transform modes

| Mode | Behavior |
|------|----------|
| `None` | Pass filtered value unchanged (default) |
| `Linear` | Remap `[FromMin, FromMax]` → `[ToMin, ToMax]` via interpolation |
| `LinearDrop` | Same remap, drop messages whose output falls outside `[ToMin, ToMax]` |
| `Noise` | Linear transform plus a secondary packet with randomized delay |
| `PreventRunningStatus` | Force full status bytes on every outgoing message |

## Licensing

MIDIRouter is free for personal use. Commercial and professional use requires a license. Contact: midirouter [at] radix-studio.fr
