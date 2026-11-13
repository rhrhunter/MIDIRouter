---
name: midiroute-rulebased
description: Create a rule-based MIDI route with per-message-type filtering, transforms, and value remapping.
---

# Skill: /midiroute-rulebased

Create a rule-based route that filters and transforms MIDI messages per message type.

## When to use

- Remapping CC values (e.g., 7-bit → 14-bit, custom range → different device range).
- Changing channels (source Ch 3 → destination Ch 1).
- Per-channel routing for multi-timbral devices (each voice on a different channel).
- Forwarding only specific message types while dropping others.

## Steps

1. Run `midirouter list` to list ports/channels.
2. For each message type you want to route, create a rule in order:
   - First matching rule wins — break on match.
3. Use generator value semantics: `"*"` (reuse), `"$"` (transformed), or literal number.

## Value remapping with Linear transform

```json
{
    "SourceDevice": "Interface/Port 1",
    "DestinationDevice": "Synth/Channel 1",
    "DefaultPassthrough": false,
    "Rules": [
        {
            "Name": "CC remap: 0–127 → 0–16383 (CC 74)",
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

## Multi-channel per-channel routing (e.g., drum machine)

One rule per channel per message type:

```json
{
    "SourceDevice": "Interface/Port 1",
    "DestinationDevice": "DrumSynth/Channel 1",
    "DefaultPassthrough": false,
    "Rules": [
        {
            "Name": "Kick - Ch1 → Dest Ch1",
            "Filter": {"MsgType": "Note On", "Channel": "1"},
            "Transform": {"Mode": "None"},
            "Generator": {"MsgType": "Note On", "Channel": "1"}
        },
        {
            "Name": "Snare - Ch2 → Dest Ch2",
            "Filter": {"MsgType": "Note On", "Channel": "2"},
            "Transform": {"Mode": "None"},
            "Generator": {"MsgType": "Note On", "Channel": "2"}
        }
    ]
}
```

## Noise transform for analog drift

Add randomized variation to a CC value after Linear remap:

```json
{
    "Transform": {
        "Mode": "Noise",
        "FromMin": 0,
        "FromMax": 127,
        "ToMin": 0,
        "ToMax": 16383,
        "NoiseSettings": {
            "Amount": 5,
            "DelayMsMin": 10,
            "DelayMsMax": 200
        }
    }
}
```

Emits a secondary packet with a randomly perturbed value after a random delay.

## PreventRunningStatus for reliable USB/classic MIDI bridging

Some devices drop running-status messages. Use this transform to force full status bytes:

```json
{
    "Transform": {"Mode": "PreventRunningStatus"}
}
```
