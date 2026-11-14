# TODO

## Critical

- [ ] #1 Div-by-zero in Linear/LinearDrop/Noise transforms (`rule/rule.go:196,206,217`)
  - Guard `fromMax == fromMin` — return error or log warning before division.

- [ ] #2 Race on `relay.lastMIDIMsg` (`router/router.go:100,125 vs 197`)
  - Use `sync.Mutex` or `atomic` to protect concurrent reads/writes of `lastMIDIMsg`.

- [ ] #3 Race on `routers` slice (`cmd/midirouter/app/route.go:39`)
  - Add `sync.Mutex` around concurrent `routers = append(routers, r)`.

- [ ] #4 Unseeded `math/rand` (`rule/rule.go:132,232`)
  - Call `rand.Seed(time.Now().UnixNano())` at package init or startup.

- [ ] #5 Generator error returns original packet as MatchInject (`rule/rule.go:253`)
  - Return `MatchNoInject` (drop) on generator error instead of `MatchInject`.

## High

- [ ] #6 Stop message (0xFC) triggers All Notes Off mid-stream in pass-through (`router/router.go:167`)
  - Remove transport-stop check from pass-through path. Move cleanup to signal handler only.

- [ ] #7 CCAh generator always errors (`gencontrolchange/gencontrolchange.go:148`)
  - Implement CCAh two-message protocol or remove the unimplemented stub.

- [ ] #8 Negative FromMin casts to huge uint32 (`config/config.go:195`)
  - Validate `FromMin >= 0` and `FromMax > FromMin` in config loader.

- [ ] #9 Pitch Wheel noise packet truncated to 7 bits (`rule/rule.go:146`)
  - Use both low and high bytes for PitchWheel values (>127) in `generateNoisePacket`.

- [ ] #10 Config load failures silently swallowed (`cmd/midirouter/app/route.go:34-37`)
  - Add `sync.WaitGroup` + error channel. Return combined error to caller.

## Medium

- [ ] #11 PreventRunningStatus is a no-op (`rule/rule.go:270-285`)
  - Implement actual running status prevention (force full status byte) or remove dead code.

- [ ] #12 Negative SendLimitMs blocks all messages (`config/config.go:117`)
  - Validate `SendLimitMs >= 0` in config loader; reject negative values.

- [ ] #13 Noise MinValue/Delay unvalidated (`config/config.go:224-227`)
  - Add `MinValue >= 0`, `MinValue <= MaxValue`, `DelayMsMin <= DelayMsMax` checks.

- [ ] #14 toMax > 65535 silently wraps in uint16 (`rule/rule.go:198,219`)
  - Validate `toMax <= 65535` for uint16-based transforms; reject or clamp in config.

- [ ] #15 Noise mode no bounds check on transformed output (`rule/rule.go:219`)
  - Clamp `transformedValue` to `[ToMin, ToMax]` in Noise mode like LinearDrop does.

## Low

- [ ] #16 Debug `fmt.Println` in production (`router/router.go:78`)
  - Remove `fmt.Println(rule)` from `AddRule()`. Keep only if verbose flag gating.

- [ ] #17 Deprecated `ioutil.ReadFile` (`config/config.go:91`)
  - Replace `ioutil.ReadFile` with `os.ReadFile`.

- [ ] #18 Unused debug print on error path (`config/config.go:246-247`)
  - Remove `fmt.Println(generateMsgType)` from error condition block.

- [ ] #19 0xF7 (SysEx end) misparse (`router/router.go:141`)
  - Handle 0xF7 as SysEx continuation or validate full SysEx packet integrity.

---

**Note:** No unit tests exist anywhere in the codebase. All changes must be verified by running against real or simulated MIDI hardware.
