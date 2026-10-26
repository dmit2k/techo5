# Experimental External ASP Frontend

Design and review documents for an **optional generic external audio-processing frontend** in TECHO5. The proposal keeps TECHO5 as the sole microphone/reference capture owner and selects between continuously available built-in PCM and a trusted local helper's 16-kHz mono output.

This directory is **documentation only**. It does **not** mean External ASP has been merged upstream, that the candidate host code is present on this documentation branch, that factory DSP support is shipped, or that a Go build tag and deployment procedure already exist.

## Read in this order

1. **[Architecture and trade-offs](architecture.md)** — full signal graph, which consumers use external PCM versus continuing built-in diagnostics, fallback behavior, 8-ms/20-ms latency choice, privacy and acceptance boundaries.
2. **[ASP2 protocol and channel mapping](protocol.md)** — exact negotiated MIC/REF counts, captured slot order, independent logical DSP destination maps, scaling, ACK/REJECT, binary framing, statuses and timing.
3. **[P3-3A.1 offline acceptance](validation-p3-3a1.md)** — low-latency-first selection policy, composite 33/33 PASS receipts, the preserved initial Spot timeout and qualification limits.
4. **[Build and integration plan](build-and-integration.md)** — current status, future Go build-tag decisions, conformance/hardware gates and the deployment instruction checklist.

## Critical distinctions

- **ASP1 (P3-2)**: fixed 4 MIC + 2 hardware REF input, comparison-only and not a general multichannel plugin interface.
- **ASP2 (P3-3A)**: a generic negotiated channel-count-aware wire format with up to 16 captured MIC and 8 REF channels (bounded by payload size), validated in offline synthetic tests.
- **Channel mapping**: the host preserves captured source-slot order. The *helper* maps source indices onto destination DSP logical channel indices. The current Athena capture tap still uses 4 MIC + 2 true stereo hardware REF; no generic capture support for other boards is implied by the protocol.
- **P3-3A.1**: low-latency-first handoff has **OFFLINE GO** with 33/33 successful canonical receipts (25 previously passed + 8 rerun, not a second uninterrupted full matrix). Built-in current-tick fallback and uninterrupted publication replace the earlier sticky delayed shadow. A preserved initial Spot deadline failure remains a host scheduling risk; no hardware realtime claim.
- **Factory adapter and hardware qualification**: future independent work, including calibration/geometry verification, true DSP integration, device realtime tests and mute/security review.

No proprietary audio, calibration data or runtime binaries belong in this documentation or a generic host PR.

**Design maturity:** offline prototype and public review draft, not a stable SDK/ABI or end-user installation guide.
