# FrankenASP 1.0 — Dev Edition 1

**Proposed release identifier:** `frankenasp-v1.0.0-dev.1`  
**Status:** experimental, developer-oriented research preview (not a stable release)  
**Supported device:** Amazon Echo Show 8 Gen 2, **Athena / MT8183 only**, with the validated original factory ASP revision  
**TECHO5 baseline:** v1.6.0, AArch64 mainline-Linux edition

## What works

FrankenASP demonstrates an alternative **microphone-capture and speech-input frontend** using the Athena owner's original Amazon factory ASP within a separate native ARM32/Bionic worker. TECHO5 continues to handle music, Home Assistant integration, wake-word detection and the rest of the voice-satellite workflow. The processor receives four physical microphone channels and two physical playback-reference channels, and returns **16 kHz mono PCM** to the existing wake-word/STT input.

The implemented research features include:

- Source-built, owner-controlled AArch64 experimental `echod` and separate ARM32 DSP worker, with no proprietary factory payload included.
- Factory identity and private-runtime integrity preflight; explicit backup-directory selection; read-only verification by default.
- Guarded, opt-in staging and trial activation with rollback; optional slot-local persistent activation with 120-second startup verification and **fallback to stock TECHO5**.
- TECHO5 update protection while the experimental persistent process may be active.
- Real native-ASP output checks, synthetic PCM/protocol tests, worker stall/failure recovery and validated cold boot into the experiment.
- Completed controlled first-pass **wake-word-only** acoustic tests during **background music played through the Echo's own speakers**.

## Initial acoustic findings

At stimulus level 35 (the study's “normal voice” level), 60 scheduled wake-word stimuli yielded:

| Voice frontend | Echo music mono | Echo music stereo |
| --- | ---: | ---: |
| Stock TECHO5 | 45/60 (75.0%) | 22/60 (36.7%) |
| FrankenASP | **59/60 (98.3%)** | **54/60 (90.0%)** |

The study's mono/stereo designations refer to the **Echo's own speaker playback of music**, not microphone-array modes or multiple ASP output channels. All conditions used four physical microphones. The stock TECHO5 effective wake threshold was observed at **0.75** during music; FrankenASP remained at **0.85**. This is an end-to-end engineering comparison at the observed settings, **not** a matched-threshold causal attribution to any one ASP algorithm.

More results, including quiet/loud stimuli, limitations and study boundaries: [Initial acoustic results](acoustic-results-20261023.md).

## Verification and reproducibility

- The experimentally deployed ARM32 worker and Amazon `libasp.so` were verified loaded together on real Athena hardware; continuous output was observed after reboot.
- Short live trials, watchdog/fallback and native PCM signal-integrity tests were performed; long-duration stability qualification remains open.
- The open staging/protocol suite contains **31 Python tests**, passed on an independent Ubuntu 26.04 host.
- Key Go microphone, firmware and update-package tests have passed. Full multi-build Go CI is a release gate and must be reported honestly as passed **only after** a successful clean-checkout run.
- A separate, clean Ubuntu 26.04 x86_64 host and an official checksum-verified Go 1.26.0 toolchain are documented in [independent build instructions](build-on-ubuntu-26.04.md).
- Original factory libraries, Bionic runtime, device-private coefficients, backups, microphone recordings, credentials and unredacted logs are **not distributed**.

## Explicitly not qualified

- Production or unattended long-duration stability; continuous ALSA/AFE and thermal resilience.
- Post-wake spoken-command quality and STT transcription accuracy, spatial speaker tracking, false-positive rate and algorithm-specific AEC/beamforming claims.
- Other Echo models, other Fire OS factory revisions, altered or missing vendor partitions.
- One-click consumer installation, automatic updates while FrankenASP runs, and stock Fire OS equivalence.
- Redistribution of any proprietary Amazon DSP library or native runtime.

**Installation is deliberately owner-controlled and potentially risky.** Keep a known-working stock TECHO5 build and device recovery method; review [Athena setup, activation and rollback](installation.md) before attempting anything. Do not interpret “Dev Edition 1.0” as an official TECHO5 release or endorsement.

## Next development tracks

Stability and fault-injection soak tests; full documentation of room/test geometry and the acoustic harness; post-wake command/STT tests with directional interference; spatial beam selection/tracking research; and architectural discussion of a neutral, versioned external audio-frontend interface for TECHO5.
