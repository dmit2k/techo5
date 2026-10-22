# Franken ASP — comparing factory and TECHO5 microphone capture and speech processing

## Purpose: a factory voice-input reference for objective and subjective A/B tests

**Franken ASP is a research project for comparing the quality of microphone
capture and speech-signal processing on Amazon Echo hardware: Amazon's
original factory voice front end versus TECHO5's open implementation.**
Our goal is to let Echo owners measure and listen to the differences in
what the voice assistant actually *hears*, using the same physical device
and reproducible conditions. Users should be able to evaluate both
approaches themselves instead of relying on assumptions or anecdotal claims.

**This is not a comparison of speaker sound quality, music playback,
amplifier tuning, or how the Echo sounds when playing audio.** Music,
TV and the Echo's own speaker output matter here only as acoustic
interference and/or playback-reference signals for echo cancellation
(AEC). The output under evaluation is the **captured and processed
microphone signal sent toward wake-word detection and speech recognition**.

The project investigates a controlled, opt-in way to run the device's
*own* original factory microphone/voice DSP as a **reference baseline**
alongside TECHO5's existing open voice-capture and processing path.
Comparisons can include:

- **Objective voice-input measurements:** wake-word detection and false
  activations; missed speech and clipped word onsets; residual playback
  echo/AEC behavior, background-speech/noise suppression, input signal
  levels and clipping, processing latency and recovery from faults.
- **Subjective evaluation of captured speech:** blind listening to
  recorded **microphone/processed-voice PCM**, focusing on speech
  intelligibility, naturalness, onset preservation and suppression of
  TV, music or other competing sounds. These recordings are not
  evaluations of the Echo's loudspeakers.

Fair comparisons must control speaker/background level, microphone
position, source material, input gain, test ordering and relevant DSP
state. We intend to publish repeatable test methods and results,
**not proprietary factory firmware or private voice recordings**.
Either voice-processing pipeline may have strengths, weaknesses and
trade-offs; which performs better is a question to measure, not a
predetermined conclusion. Findings should help improve TECHO5's
fully open **microphone and speech-processing path**.

**Important limitation:** running Amazon's original ASP library under
TECHO5 Linux does not recreate the complete Fire OS microphone pipeline,
its control state or its capture/reference timing. Claims of equivalence
to voice capture on stock Fire OS require separate validation. This
project does **not** currently provide one-click A/B switching for every
TECHO5 user.

## Current experimental implementation and status

**Research preview, not an official TECHO5 feature or a ready-to-install build.**
The working proof of concept is limited to **Echo Show 8 Gen 2
(Athena, MT8183)** on a specific tested factory firmware build. It
routes four microphone channels and two physical playback-reference
channels to a separate ARM32/Android Bionic process running the
owner's original factory **speech-input** processor. TECHO5's open
microphone/speech-processing pipeline stays the default. This experiment
does not alter or benchmark the speaker playback path, amplifier tuning,
upstream stereo support or kernel patches.

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
- `tools/frankenasp/install.py`: explicit `--backup-partitions DIR` (no discovery), read-only verification by default; optional guarded staging of owner's private ARM32 system runtime and self-built experimental binaries with `--install`, **without activation**.
- `tools/frankenasp/tests`: synthetic PCM and fixed-point tests.

## Device and firmware scope

The **37** factory DSP configuration/model files observed under
`/vendor/etc/audio-algorithms/` are from **this Athena device's original
Fire OS**. The directory also contains ASR, VoIP, playback, Bluetooth
and other settings — **modes of Athena hardware, not factory profiles
for every Echo**. Their presence does not expand Franken ASP's scope:
this research compares **microphone capture and voice processing only**,
not the sound produced by the speaker.
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
- [Owner setup: guarded automatic installer and step-by-step manual instructions](installation.md)
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

No experiments should overwrite the official TECHO5 binary, modify
factory partitions or change speaker playback settings as part of
microphone-quality comparisons. Keep the official release and its recovery method.
All AEC, beamforming and wake-word quality claims remain provisional
until measured with appropriately labeled recordings.
