---
name: midiroute-passthrough
description: Create a simple passthrough MIDI route between any source and destination device.
---

# Skill: /midiroute-passthrough

Create a simple passthrough route — bridges any source port to a destination with no filtering or remapping.

## When to use

- You just need to forward all MIDI transparently (e.g., hardware synth → DAW, sequencer → multi-synth interface).
- No value remapping or channel changes needed.

## Steps

1. Run `midirouter list` to list available ports/channels.
2. Pick source and destination from the output.
3. Create a config:

```json
{
    "SourceDevice": "<source-port-name>",
    "DestinationDevice": "<dest-device-name>",
    "DefaultPassthrough": true,
    "Verbose": false
}
```

4. Run it: `midirouter route /path/to/config.json`

## Example source/dest names (from macOS CoreMIDI device listing)

- `"Focusrite USB ASIO/Port 1"`
- `"Novation /Launchkey Mini SL mkIII/Port 0"`
- `"Ableton Push 3/Port 1"`
- `"Arturia Beatstep Pro/Port 1"`
- `"IAC Driver/IrDA Port"` (for looping back to another app)
