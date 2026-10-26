# P3-3A.1 — low-latency handoff offline acceptance

> **Result:** OFFLINE GO for the synthetic ASP2 host/handoff implementation, **not** for a hardware-qualified device feature or an upstream-stable external-helper ABI.

The production changes are available only in the private experimental development branch; this documentation-only GitHub branch does **not** contain or release the ASP host implementation.

## Exact implementation checkpoint

| Item | Revision |
|---|---|
| Workstream | P3-3A.1: Low-Latency-First ASP Handoff |
| Experimental code branch | `experimental/asp-low-latency-handoff` (not published here) |
| Code HEAD | `831995bab26d81afeac003f2e70cb296b9a7e472` |
| Code tree | `b57e91d11d7d09a8961b15b1238e96fdee046e54` |
| Previous P3-3A basis | `e95848c35de0addaffd98441e8a38d3550d71eee` |
| Frozen P3-2 / ASP1 comparison | `7f5714242eba0977c7c82e4415596a68afa57b8b` |
| Offline verdict | GO, subject to the explicit limitations below |

The isolated implementation worktree was clean at acceptance. The full-history Git bundle was independently cloned, its HEAD and tree compared with this checkpoint, and the restored clone passed `git fsck --full`. No factory DSP assets or live device audio are included.

## Why the handoff policy changed

The P3-3A baseline initially prioritized source-sample continuity *across different processors*: it intentionally withheld one output publication (~40-ms wall-clock interval) and then kept a delayed stock shadow after External failed or was switched off. That behavior preserved one global input-time sequence at the cost of a permanent additional stock delay.

For a streaming wake-word assistant, the revised policy prioritizes **current-tick Built-in latency and one delivery per successful capture tick**. It no longer postpones stock audio to preserve an exact global source-time sequence across different processing algorithms.

P3-3A.1:

- Publishes **exactly one current 320-sample mono frame per successful 20-ms capture tick** in the supported test model; no intentional 40-ms gap.
- Uses current-tick Built-in PCM during HELLO, CFG2, NOT_READY and the three validated DATA preflight replies.
- Promotes External only on the next validated DATA output (the fourth valid DATA after a first NOT_READY in the factory-shaped scenario).
- Returns to current-tick Built-in PCM immediately on manual Built-in selection or any fatal/invalid/timeout response; no permanently delayed stock shadow.
- Allows at most one **320-sample input-time overlap** on Built-in→External or one **320-sample input-time skip** on External→Built-in at the processor boundary, due to the bounded one-frame DSP latency.
- Enforces continuous, strictly increasing 320-sample origins **within a healthy external session**. No fabricated PCM, duplicated or omitted DSP intervals during steady operation, and no accumulated lag across repeated activations.
- Does not clear MWW/STT state, duplicate publications, catch up by sending extra frames, or synchronously drain the final DSP tail at shutdown.

These are deliberately different guarantees: continuity of PCM **delivery cadence** is prioritized over lossless source-time concatenation **at a processor transition**. The static ASP2 request/response format, status codes, configurable MIC/REF descriptor and helper-local channel mapping are not changed by P3-3A.1.

## Offline test evidence

The final acceptance inventory comprises **33 successful canonical job receipts on the same experimental source revision**.

**Provenance matters:** this was a *composite acceptance*, not a second uninterrupted run of all 33 jobs. Twenty-five original successful jobs were preserved and verified, and eight canonical jobs were rerun successfully, including complete Show and Spot suites, race/integration, 100k-frame processing and the frozen ASP1 comparison. The different reference trees used in baseline-golden/ASP1 checks were individually verified rather than mislabelled as the candidate tree.

Coverage includes five board-tag full suites (Show, Spot, Dot, Donut, Panel), affected-package race detection, actual synthetic child pipes→Source→HA Store save/reload, long-running 100k-frame External continuity, no-helper bit-exact stock golden checks, unchanged WebRTC tests, helper failure/timeout tests, Go vet, six Linux ARM/ARM64 cross-builds, and paced logging stress.

The final verification checked each recorded exit code, relevant tested HEAD/tree, clean tracked status and SHA-256 of the corresponding test log. A separate restore from the sealed Git bundle passed integrity checks.

### Original scheduling-sensitive Spot failure

The initial full Spot suite had one failing test:

`TestASP2QuantumSchedulingAndHardDeadline` — the synthetic 2/3-quantum adapter exceeded the provisional 25-ms write+read watchdog on the first exchange under that full-suite execution.

No timeout threshold was increased, no test was disabled and no ASP2 implementation code was changed to conceal the failure. The identical test passed **5/5 isolated reruns**. A subsequent complete Spot run, including `internal/hardware/mic`, also passed. The original FAIL receipt and log remain retained in private evidence alongside the successful rerun.

Successful host reruns do **not** prove worst-case real-time scheduling under contention. Measurements around the 25-ms deadline remain an explicit integration risk. The 25-ms bound is provisional and the real hardware ring/ALSA budget still requires measurement.

## Channel mapping and scope

The already implemented **ASP2** descriptor announces captured MIC and REF channel counts, formats, rates and scales. The host leaves captured source-slot order intact; DSP logical destination mapping is owned by the helper. Generic synthetic tests cover variable input counts and permutations. The existing Athena native tap supplies **4 captured MIC + 2 independent 96-kHz hardware REF channels**, and a synthetic factory-shaped profile accepts only that topology.

These protocol tests do not identify the real physical microphone-to-ADC geometry, establish calibration ownership, or prove that any vendor-specific DSP binary runs correctly. That work belongs to a separately authorized factory bridge and on-device qualification stage.

## Gates not satisfied by this offline GO

- Actual Amazon factory ASP/Bionic worker, its calibration, beam/status decoder, runtime dependencies and acoustic processing quality.
- Real Athena ARM CPU latency, scheduling tails, reference synchronization, ALSA ring occupancy, playback AEC quality and wake-word/STT performance.
- Hardware and software mute trust-boundary qualification, including any platform that can feed native samples to the helper before a software PCM cut.
- Final upstream-friendly opt-in Go build tag, build instructions and a stable public plugin ABI.
- Any actual device install, live Home Assistant device modification or upstream submission.

See [Architecture](architecture.md), [ASP2 framing and mapping](protocol.md) and [Build and integration](build-and-integration.md).