# Franken ASP — comparing factory Echo audio with open TECHO5

## Purpose: an original-DSP reference for fair A/B comparisons

**The goal of Franken ASP is to let Echo owners objectively measure and
subjectively hear the differences between Amazon's original factory audio
processing and TECHO5's open audio-processing pipeline on the same hardware.**
Anyone considering or using TECHO5 should ultimately be able to evaluate the
trade-offs themselves, rather than relying on marketing claims, assumptions,
or anecdotes about which pipeline performs better.

The project investigates a carefully controlled, opt-in way to run the
device's *own* original factory audio processor as a **reference baseline**
alongside the existing open implementation. Comparisons can include:

- **Objective:** wake-word hits, misses and false activations under
  reproducible speech/background conditions; residual echo/AEC behavior,
  input clipping, signal levels, processing delay and recovery from faults.
- **Subjective:** intelligibility, word onset preservation, suppression of
  music/TV/background speech, and listening comparisons of recorded output.

Fair trials must control microphone position, source material, playback
volume, capture gain, test ordering and relevant DSP state. We intend to
publish repeatable test methods and results, **not proprietary factory
firmware or private voice recordings**. Evidence may reveal advantages,
weaknesses or trade-offs in *either* pipeline; superiority is a question
to test, not a predetermined conclusion. Findings should also inform
improvements to TECHO5's fully open audio path.

**Important limitation:** using the factory ASP library under Linux is
not the same as booting all of Fire OS or reproducing every stock control
path, timing relationship and acoustic state. A direct comparison against
original Fire OS needs separate validation. The project does **not**
currently provide one-click A/B switching for every TECHO5 user.

## Current experimental implementation and status

**Research preview, not an official TECHO5 feature or a ready-to-install build.**
The working proof of concept is limited to **Echo Show 8 Gen 2
(Athena, MT8183)** on a specific tested factory firmware build. It
routes four microphone channels and two physical playback-reference
channels to a separate ARM32/Android Bionic process running the
owner's original factory audio processor. The original TECHO5 audio
path stays the default. The experiment does not replace upstream
speaker stereo support or kernel patches.

**This repository distributes only independently authored source code,
tests and documentation.** No Amazon binaries, partition images,
beamformer coefficients, native DSP configurations, original factory
runtimes or voice recordings are included. Owners supply any required
factory components from their own lawfully obtained device backups,
keep them private and remain responsible for applicable license and
copyright requirements. The research/benchmarking purpose does not
itself grant permission to redistribute third-party materials.

## Contents

- `echod/internal/hardware/mic/factory_asp_athena.go`: guarded worker invocation,
  Q30 microphone calibration, six-channel packing and 16 kHz mono speech output.
- `athena.go` / `mic.go`: stereo AWB capture and delivery to MWW/STT without
  secondary TECHO5 AEC/denoising/leveling.
- `tools/frankenasp/asp_stream_worker.c`: clean research source code for ARM32
  Bionic worker; no native factory payload is embedded.
- `tools/frankenasp/extract_runtime.py` and `extract_configs.py`:
  read-only extraction from operator-selected factory partition backups.
- `tools/frankenasp/build_worker.sh`: optional Android NDK source build.
- `tools/frankenasp/status.sh`: distinguish requested marker from loaded ASP.
- `tools/frankenasp/audit_capture.py`: read-only, no-audio-playback integrity audit of previously saved 4-mic/2-AWB PCM (detects silent reference).
- `tools/frankenasp/preflight.py`: read-only inventory of preserved factory vendor, private Android runtime and optional original system backup (no extraction or install).
- `tools/frankenasp/tests`: synthetic PCM and fixed-point tests.

## Device and firmware scope

The **37** factory DSP configuration/model files observed under
`/vendor/etc/audio-algorithms/` are from **this Athena device's original
Fire OS**. Their ASR, VoIP, playback, Bluetooth and other settings are
**modes of the Athena hardware**, not factory profiles for every Echo.
Original `AFE.cfg` and `asp.cfg` identify the target board as `Athena`,
with four microphones and two speakers. Neither the Athena `libasp.so`
nor its beamformer coefficient bundle has been validated on any other Echo.
A reusable **multi-device Franken ASP framework** remains a possible
future design, but each board/firmware needs its own validated DSP/runtime,
calibration and synchronized-reference adapter.

See [multi-device compatibility and portability scope](multi-device-scope.md).

## Documentation

- [Architecture and open questions](architecture.md)
- [Multi-device scope: why 37 factory files are *not* universal Echo profiles](multi-device-scope.md)
- [Exact Athena factory paths, ELF dependencies and surviving partitions](partition-dependencies-20261023.md)
- [Local runtime assembly and limitations](runtime.md)
- [Controlled comparison protocol](testing.md)
- [Stage 3 preparation: input integrity, zero-AWB finding, factory preflight and acoustic test metadata](stage3-preparation-20261023.md)

## Safety and milestones

The research prototype has now completed limited live **Athena-only**
trials with time-limited rollback, controlled worker termination and a real
SIGSTOP-induced native IPC timeout. Detailed observations are in the
[trial reports](ipc-hang-trial-20261023.md).
**This is not a qualified release.** Long-duration ALSA/AFE stability,
acoustic wake-word and AEC performance, cross-firmware compatibility and
per-device installation remain unvalidated. Do not use the persistent
`/data/techo5-linux/fireos-asp.enabled` marker alone: the guarded trial
also requires the explicitly temporary `/run/techo5/frankenasp-armed`
flag and an independent watchdog.

No experiments should overwrite the official TECHO5 binary or modify
factory partitions. Keep the official release and its recovery method.
All AEC, beamforming and wake-word quality claims remain provisional
until measured with appropriately labeled recordings.
