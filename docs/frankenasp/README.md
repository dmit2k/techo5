# Franken ASP — experimental Athena factory audio processor

**Status: research preview. Not an official TECHO5 feature or a ready-to-install build.**

This branch experimentally routes four microphone channels and two channels of
physical playback loopback on **Echo Show 8 Gen 2 (Athena, MT8183)** to a separate
ARM32/Android Bionic process running an owner-supplied factory audio processor.
It does not replace the upstream speaker stereo implementation or kernel patches.

**No original Amazon binaries, images, coefficients, native DSP configurations,
voice recordings or complete factory runtimes are distributed here.**
Participants must extract components from their own legitimately acquired
device backups, keep them outside Git and review applicable licenses.

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
