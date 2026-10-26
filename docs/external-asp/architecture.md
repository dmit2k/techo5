# External ASP Frontend — architecture and design trade-offs

> **Status:** experimental design, not an upstream feature or a hardware-qualified release. This document describes the intended generic interface and the offline-tested P3-2/P3-3A implementation. A later low-latency handoff revision (P3-3A.1), factory integration, build-tag packaging and device acceptance remain separate review gates.

## Goals and non-goals

The External Advanced Sound Processing (ASP) Frontend allows TECHO5 to deliver 16-kHz mono microphone PCM produced by a separate local processor instead of its built-in frontend. The external processor can implement complete array processing (beamforming, AEC, denoising, gain control, etc.); it is **not** merely a selectable echo canceller.

The existing TECHO5 process remains the **sole ALSA microphone/reference capture owner**. The external helper receives a bounded copy/exchange of the native captured samples through local binary pipes and returns PCM. It never opens a second ALSA capture stream through this interface.

The principal objectives are opt-in behavior, no required proprietary dependencies in the host, preservation of the stock audio path and diagnostics, honest Home Assistant controls, deterministic bounded failure handling and a reusable channel-count-neutral protocol.

Non-goals include a generic DSP sandbox, a public stable plugin ABI today, speaker-output replacement, arbitrary per-frame topology changes, exposing factory geometry in the host, or making the vendor DSP itself redistributable.

## Processing graph

```text
                               One physical capture owner
                         +-----------------------------+
      physical mics ---->| native microphone capture   |
   playback L/R refs --->| and clock-aligned reference |
                         +-------------+---------------+
                                       |
                 +---------------------+----------------------+
                 |                                            |
        Native capture view                          Built-in conversion
      original mic/ref slots                         into TECHO5 input
                 |                                            |
      ASP2 StreamConfig/INP2                                  v
                 |                               mixer / optional AEC
                 v                               denoise / leveler and
     external-asp local process                     room-related analysis
      DSP-owned channel maps                                 |
      geometry / calibration                                  |
      8 ms or other DSP quanta                                |
                 |                                            |
      RSP2: 16-kHz mono PCM                             stock mono PCM
                 |                                            |
                 +---------------------+----------------------+
                                       |
                              final PCM selection
                                       |
                        +--------------+---------------+
                        |              |               |
                   history/pre-roll  spectrum   all mono Listen()
                                                      subscribers
                                          (wake word, STT, talkback ...)
```

The external selection happens **after** the built-in processors have updated their internal states and diagnostics, but **before** the common mono PCM history/spectrum/listener publication. Exactly one selected mono output is published at that boundary. TECHO5's `ListenRaw()` remains a separate stock-converted multichannel feed.

## Which signal drives each feature?

The processing source differs by API. This distinction is part of the interface contract, not an accidental implementation detail.

| Feature or observable | When External is active | Why |
|---|---|---|
| Wake-word inference input | Selected external PCM | Uses ordinary mono `Listen()` |
| Speech recognition and voice PCM consumers | Selected external PCM | Same mono subscriber stream |
| Talkback and other `Listen()` consumers | Selected external PCM | Existing subscriptions remain unchanged |
| Recent history / wake pre-roll / `ListenHistory()` | Selected external PCM | Stored after the selector |
| Microphone spectrum | Selected external PCM | Pushed after the selector |
| `ListenRaw()`, raw capture consumers | Built-in converted multichannel raw | Separate subscriber path, not ASP output |
| Room-noise measurement | Built-in captured/processed signal | Computed before selection |
| Standout / nearest-room score | Built-in signal | Updated before selection |
| Facing / direction estimate | Built-in beam analysis | No external beam metadata currently supplied |
| Built-in level, floor, clipped-sample and gain statistics | Built-in leveler | Leveler executes before selection |
| Built-in AEC / ERLE statistics | Built-in cancellation path | External helper does not replace telemetry |
| Analog capture gain / hardware setup | Shared physical capture | Affects both paths at the source |

Accordingly, the existing **mixing, built-in echo cancellation, denoising and leveling controls do not post-process external PCM**. They continue configuring the always-running stock path and associated measurements. A user must not assume that a built-in diagnostic such as `Noise` or `Facing` describes the externally beamformed output.

This design keeps stock DSP state warm and allows an immediate stock-frame fallback. Costs include extra CPU/RAM consumption, possible disagreement between built-in room diagnostics and the sound actually heard by wake recognition, and no external direction/beam/level telemetry unless a future separately designed interface supplies it.

## Device capture vs DSP geometry

These are intentionally distinct:

1. **Physical microphone** — an actual sensor, not necessarily one independently captured PCM slot.
2. **Captured MIC/REF slot** — a channel in the native capture buffers, in the native provider's declared stable order.
3. **DSP logical input** — an input expected by a particular algorithm and geometry/calibration profile.

The TECHO5 host declares the captured shape in an immutable ASP2 stream descriptor. It does **not** reorder channels into a vendor-defined beam geometry. The helper owns validation and mapping from captured source slots to its DSP logical destinations; see [Protocol and channel mapping](protocol.md).

A new device may require native ALSA/capture support before it can supply the descriptor. A channel-count-generic wire protocol does not automatically make that device's driver, reference or array geometry compatible.

## Startup, activation and configuration

The optional executable is called `external-asp`. The established preferred location is `/data/techo5-linux/bin/external-asp`; additional safe sibling/PATH lookup follows host policy. It is a trusted local process, not a user-supplied command string. Stdout is exclusively framed ASP protocol; stderr is for diagnostics.

The helper starts without blocking Home Assistant restoration. A successful ASP2 greeting followed by an accepted `CFG2` descriptor does not by itself prove that usable PCM exists. **Three consecutive valid real-capture DATA outputs** are discarded while normal stock PCM continues to be published. Only the next valid applicable frame may activate External. `NOT_READY`, handshake, statuses, silence in an invalid packet or a claimed process-ready flag cannot replace the live PCM requirement. Protocol-valid all-zero PCM is legitimate silence and can count as data.

There are exactly three HA entities, on the existing microphone device:

| Entity | Type | Responsibility |
|---|---|---|
| ASP frontend (object `microphone_asp`) | Config Select | Requested Built-in / External |
| Active ASP (object `active_asp`) | Diagnostic TextSensor | Actual currently selected PCM source |
| ASP debug (object `asp_debug`) | Diagnostic TextSensor | Operational state, helper status, failure reason |

On a **first-time** External request, the new choice is saved only upon real PCM promotion. If preflight fails, the UI returns to Built-in and the stored Built-in remains. In contrast, if External was **already saved**, a boot/runtime failure keeps that saved choice and displays Built-in as active with the relevant fallback. An explicit Built-in command saves Built-in. Configuration export/restore uses the normal TECHO5 Store policy.

A failed process is terminal for that instance; there is **no implicit restart every frame**. A fresh attempt needs a deliberate Built-in→External operation or a subsequent daemon start that restores a saved External request. Concurrent commands are resolved last-command-wins; stale child completion cannot reselect PCM or overwrite configuration.

## 8-ms DSP vs 20-ms delivery

TECHO5 delivers 320 mono S16 samples per 20-ms tick at 16 kHz. A representative factory DSP requires 768 native samples at 96 kHz per 8-ms quantum and emits 128 mono samples at 16 kHz. In the first 20-ms input window only two full 8-ms calls are possible (256 output samples); the remaining 4 ms must carry into the next tick. The common boundary is 40 ms (five DSP quanta, two TECHO5 frames).

The synthetic DSP-shaped adapter implements a bounded sample-preserving FIFO and proves the 2/3/2/3 quantum pattern without zero padding or arbitrary duplication. It can return `NOT_READY` on the first request and carry an explicit source-origin index in subsequent replies.

Two handoff policies have been studied:

- **Full cross-processor sample continuity (P3-3A baseline):** a one-tick publication pause and persistent delayed stock shadow. This avoids an origin discontinuity but introduced an approximately 40-ms publication interval and permanently delayed future Built-in playback to the voice pipeline.
- **Low-latency-first (P3-3A.1 revision under qualification):** publish current stock on *every* capture tick, including preflight and failure. Activate a valid delayed external frame without an intentional publication pause. Permit a **single bounded overlap of at most 320 input-time samples at Built-in→External**, or a **single skip at External→Built-in**, but never corrupt the continuous sequence within one healthy external run. The stock path returns immediately to current-tick PCM, with no sticky delay.

An overlap or skip of *source-time identity at a processor boundary* is distinct from duplicated byte blocks within the external DSP stream. This is an explicit streaming-latency trade-off, not lossless file conversion. The low-latency revision must pass its final regression before being called stable.

Neither policy removes the DSP's own internal group delay. `pcmStartSample` identifies a position in the input timeline, not a measured acoustic end-to-end latency.

## Fallback and realtime limits

TECHO5 computes the stock frame regardless of the selected mode. Malformed or stale responses, missing genuine reference, invalid configuration, child exit, timeout and fatal helper status immediately select stock for the failing tick and later ticks. The child is closed/killed/reaped through a bounded lifecycle; it is not silently restarted.

The provisional host I/O deadline is **25 ms total per exchange (request write + reply read)**, not per syscall. Startup HELLO uses 500 ms; on first capture, configuration negotiation and first exchange share the per-tick budget. Host synthetic timing tests do **not** prove that the scheduler, ARM CPU, ALSA buffers or real factory worker meet this deadline.

Athena's currently documented physical microphone/reference rings are approximately 30/40 ms deep. A 25-ms stall can threaten a 20-ms capture cadence, despite eventually falling back. Measure overruns, frame drops, reference synchronization, CPU, memory, deadline percentiles and wake-word/STT performance on separately authorized hardware before deploying this as a product.

A transition between independently processed streams can create a brief audio-characteristic or source-time discontinuity. The design intentionally **does not** clear wake-model state, flush history, reconstruct subscriptions or force a second audio event when switching.

## Privacy, trust and redistribution

Native microphone PCM and hardware playback references are sensitive: the helper must be **trusted local code** with carefully controlled filesystem ownership and permissions. The implementation has no arbitrary remote helper endpoint or user-supplied executable path.

Hardware mute and software-only mute are not interchangeable. The initial processing call in the experimental capture pipeline occurs before `SoftwareCut()`, while the selected output is committed later. That boundary needs explicit privacy tests and, where relevant, protection **before raw native samples are handed to the helper**. A platform must not be enabled merely because it compiles.

The generic host design and synthetic conformance fixtures are separable from any proprietary DSP binaries, coefficients, models, mic calibration or vendor runtime. Such assets are not to be committed with the generic upstream proposal.

## Upstream boundary and outstanding gates

The prospective upstream contribution is the **generic, opt-in TECHO5 ASP host interface**: capture descriptor, process lifecycle, bounded ASP2 protocol, status/control, fallback and sample publication. Vendor-specific geometry, maps, calibration and runtime belong to independently supplied helpers.

Outstanding acceptance gates include the final low-latency handoff regression, real factory bridge (a distinct follow-on stage), acoustic and realtime device qualification, privacy/mute review, protocol stabilization and a decision on optional Go build tags. No tag is assumed to exist; see [Build and integration](build-and-integration.md).

The standalone protocol documentation is a **design contract for review**, not a guarantee of plugin ABI compatibility across unreleased revisions.
