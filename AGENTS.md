# MIDIRouter — MIDI routing engine on macOS

Go utility that bridges any CoreMIDI source to destination via JSON configs.

## Architecture

```
Source Device ──► MIDIRouter (CoreMIDI)
                    │
                    ├─ Rule pipeline: Filter → Transform → Generator
                    │   Per-rule: filternoteon/, filtercontrolchange/, etc.
                    │         gennoteon/, gencontrolchange/, etc.
                    │
                    └─ Output ──► Destination Device
```

- Loads one or more JSON configs at startup via CLI args. Each config = one route.
- CoreMIDI source → rule pipeline → CoreMIDI destination.
- First matching rule wins (break on match). If no match and `DefaultPassthrough: true`, passthrough unchanged.
- On exit (`SIGINT`/`SIGTERM`): All Notes Off + Reset Controllers on all 16 channels.

## Running

```bash
# Single route
midirouter route config/passthrough.json

# Multiple simultaneous routes (each in its own goroutine)
midirouter route config/device-a.json config/device-b.json

# List available MIDI ports/channels, then exit
midirouter list
```

## Config format reference

### Top-level fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `SourceDevice` | string | required | e.g. `"Interface/Port 1"` — any CoreMIDI source port name |
| `DestinationDevice` | string | required | e.g. `"Synth Name/Channel 1"` — any CoreMIDI destination |
| `DefaultPassthrough` | bool | false | Pass unmatched messages through unchanged |
| `Verbose` | bool | false | Log every routed packet |
| `Rules` | Rule[] | none | Per-message-type routing rules (ignored if passthrough) |

### Rule: Filter settings (per MsgType)

**Note On / Note Off:** `"Note": "*"|"N"`, `"Velocity": "*"|"V"`
**Control Change:** `"ControllerNumber": "*"|"CC#"`, `"Value": "*"|"val"`, `"Mode": "Standard"\|"CCAh"`
**Program Change:** `"ProgramNumber": "*"|"prog"`
**Pitch Wheel:** `"Pitch": "*"|"pitch"` (14-bit)
**Channel Pressure / Aftertouch:** `"Pressure": "*"|"press"`

`"*"` = match any. `channels: "1"`–`"16"` or `"*"` on every filter/gen block.

### Transform modes

| Mode | Behavior |
|------|----------|
| `None` (default) | Pass filtered value unchanged |
| `Linear` | Remap `[FromMin, FromMax]` → `[ToMin, ToMax]` via interpolation |
| `LinearDrop` | Same remap but drop messages outside `[ToMin, ToMax]` after transform |
| `Noise` | Linear + emit secondary random packet after random delay (`DelayMsMin`–`DelayMsMax`) |
| `PreventRunningStatus` | Force full status byte on every outgoing message |

### Generator value semantics

In Generator Settings: `"*"` = reuse original, `"$"` = use transformed value, `<number>` = literal.

7-bit MIDI: 0–127. Pitch wheel: up to 0xFFFF.

### Message types

`Note On`, `Note Off`, `Control Change`, `Program Change`, `Pitch Wheel`, `Channel Pressure`, `Aftertouch`, `SysEx`.

## Writing configs

**Passthrough** (bridge two ports, no remapping):
```json
{
    "SourceDevice": "MyInterface/Port 1",
    "DestinationDevice": "TargetSynth/Channel 1",
    "DefaultPassthrough": true
}
```

**Rule-based** — one rule per message type per source channel. Each config file defines a single route with its own rules.

## Skills

Route creation helpers in `skills/`. Use `/midiroute-passthrough` or `/midiroute-rulebased` to spin up a route on the fly.
