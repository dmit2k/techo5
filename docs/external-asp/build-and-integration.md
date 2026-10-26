# External ASP — build, integration and acceptance guide

> **Draft / not a deployment guide.** The generic ASP host is being developed and tested in isolated experimental branches. The optional Go build-tag boundary is **not defined or implemented as a stable release contract**. This file deliberately does not provide fabricated `go build -tags ...` commands.

## Intended packaging

The objective is to make External ASP an **opt-in, reviewable generic TECHO5 feature** that can be built without access to a vendor DSP library, calibration data or factory firmware.

A future upstream-ready split should contain:

| Component | Owner | Distribution expectation |
|---|---|---|
| Native-capture descriptor and host transport | TECHO5 | Generic code with explicit board compatibility |
| ASP2 codec, status, guarded lifetime and fallback | TECHO5 | Generic code and synthetic conformance tests |
| HA Select + diagnostic TextSensors + saved preference | TECHO5 | Only when the feature is supported/enabled |
| Synthetic helper and sample mapping examples | Open test project | No proprietary code, no real calibration |
| Device-specific channel map, geometry and DSP profile | External helper | Profile-owned; not host HA/config |
| Factory DSP runtime, its binaries and calibration | Vendor/authorized device owner | **Not** bundled by generic TECHO5 |
| Build-tag and packaging instructions | Future reviewed patch | **Not yet finalized** |

A generic wire format does not make every TECHO5 device's native capture layout available. Physical 7/8-channel devices still need truthful capture providers; the existing Athena native tap currently supplies only 4 MIC + 2 REF.

## Open build-tag decision

Before writing reproducible build instructions, the implementation review must explicitly decide:

- Whether a build tag gates **the entire External ASP host**, only native high-rate capture hooks, or both.
- Whether stock/default binaries omit all ASP entities or show a disabled informational status.
- Which boards may expose the feature and when a missing true hardware playback reference is an unsupported configuration.
- How disabled builds retain stock-parity and unchanged audio/privacy behavior.
- How the optional helper is packaged or discovered without embedding proprietary assets.
- How Go tag combinations interact with existing board tags, architecture targets and tests.
- Whether the public generic API is considered stable enough to promise third-party binary compatibility.

**No particular tag name is reserved by this document.** Do not infer that `external_asp`, `asp` or another example is already implemented. Actual command lines will be added only after a tested build-tag patch exists.

## External helper deployment contract (conceptual)

When implemented and enabled, the host preferentially looks for an executable named `external-asp` at:

`/data/techo5-linux/bin/external-asp`

Other safe lookup mechanisms, if retained, are host-defined and do not permit arbitrary HA-supplied command paths. Binary and parent-directory ownership/permission checks are required. Discovery is cached for the daemon lifetime; installing the helper after startup may require a daemon restart.

The helper must be built for the actual **device architecture and ABI**. A separate factory processor may require a compatibility runtime; a successful host cross-compile does not prove factory DSP loading or real-time execution.

An ASP2 helper must write its startup HELLO to stdout, accept one CFG2 descriptor, perform one exact request/reply exchange per successful capture frame, and use stderr for logs. A helper should validate its declared MIC/REF topology and DSP geometry **before** accepting input processing. It must refuse unsupported reference layouts and calibration rather than silently substituting channels.

The helper's own mapping/profile configuration has not been assigned a universal file location or schema beyond the illustrative `micChannelMap` / `refChannelMap` concept. That deployment decision belongs to the device-specific helper integration.

## How users will operate it

The feature, if available, uses three entities on the existing microphone device:

| Display name | Kind | Purpose |
|---|---|---|
| ASP frontend | Select | User requests Built-in or External |
| Active ASP | Diagnostic TextSensor | Shows which PCM actually reaches listeners |
| ASP debug | Diagnostic TextSensor | Startup, warning or fallback state |

Only the second entity proves that externally processed PCM is in use. Choosing External does not by itself establish a working helper.

The first External activation is validated on actual native PCM and persisted only after its real output is committed. Already saved External is retained across boot/runtime fallback; the next boot may attempt the helper again. Manual Built-in selection reverts and persists Built-in. There is no silent auto-retry loop after process failure.

External mode does not disable stock calculations: microphone mixing/AEC/denoise/leveler settings continue to affect the built-in fallback and metrics, while not post-processing external PCM. The device's analog microphone gain may still affect both paths.

## Verification matrix required before installation

### Offline engineering gates

- Correct descriptor and immutable MIC/REF captured slot ordering.
- Variable channel counts and every supported sample format, scale/alignment combination.
- Explicit DSP logical channel maps and rejection of invalid/unverified geometry.
- No REF fabrication; capability ACK for deliberate reference-free processing.
- Valid DATA and NOT_READY, strict sample-origin checks, bounded 8→20 FIFO and finite resource use.
- Three real DATA preflight outputs prior to switching.
- Fallback on startup failure, IPC corruption, timeout, fatal helper status and missing real reference.
- Exactly three HA entities with correct persistence and asynchronous cancellation.
- Bit-identical stock capture/delivery when no external helper is present.
- No regressions to existing WebRTC, mute behavior and other board builds.
- Documented latency boundaries and transition discontinuities.

### Hardware gates (not satisfied by the experimental offline checks)

- Verify capture hardware nodes, sampling rates, channel-slot ordering and true left/right hardware reference.
- Prove the physical microphone-to-captured-slot-to-DSP geometry mapping with real signals.
- Verify miccal/calibration ownership and whether calibration is applied by the vendor DSP itself.
- Quantify mic/ref clock synchronization and any native-to-DSP alignment offset.
- Verify HELLO, descriptor ACK and first exchange fit real device budgets without stalling capture.
- Measure peak/p95/p99 processing delay, sustained CPU/RAM and ALSA ring overruns with speaker playback and multitasking.
- Test wake-word false-positive/false-negative impact, voice latency and speech-recognition quality.
- Test helper exit, hard stall, malformed output and manual override under real load.
- Validate the interaction with hardware mute **and any software-only mute** before the helper can read microphone content.
- Ensure installation and uninstall preserve the original system and allow restoration.

Only after both groups pass should a user-facing device deployment recipe claim hardware support.

## Privacy and diagnostics review

The generic interface supplies native array data and the playback reference to a separate executable. It has no privilege isolation equivalent to a security sandbox. Treat the helper and its configuration as trusted code.

Do not publish personal recordings, live calibration values, binary factory libraries, private host/user paths, internal IP addresses, hardware unique IDs or model coefficients in docs, fixtures, commits or log examples.

Host-facing status supports `000`/`2xx`/`3xx`/`4xx`/`5xx` classifications and a bounded readable message. Helpers must keep secrets out of status, even when the host filters unsafe characters. Detailed diagnostics belong in standard logging, not repeated per-frame HA state changes.

## Future instructions to insert here

After an explicit build-tag implementation and an independently reviewed build matrix, replace this section with **verified**, platform-specific instructions covering:

1. Exact upstream/fork commit and tag name.
2. Requirements (Go toolchain, supported board/ABI, dependencies).
3. Default build (feature disabled) and opt-in tagged build (feature enabled).
4. Tests with and without the tag for each supported board.
5. Helper placement, executable ownership and permission checks.
6. No-helper first boot, helper discovery/restart and HA entity appearance.
7. Configuration of helper-local channel maps **only when validated**.
8. A deterministic synthetic end-to-end smoke test and a separate hardware qualification procedure.
9. Recovery: select Built-in, preserve diagnostic reason, remove helper when appropriate, restore stock binary.
10. Version compatibility and migration between experimental ASP1 and ASP2 helpers.

Do not treat this checklist as instructions that have already been executed or demonstrated.

See [Architecture](architecture.md) and [Protocol](protocol.md) for the signal and wire-level contracts.
