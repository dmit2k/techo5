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

## Documentation

- [Architecture and open questions](architecture.md)
- [Local runtime assembly and limitations](runtime.md)
- [Controlled comparison protocol](testing.md)

## Safety and milestones

The prior prototype processed 1250/1250 native 8-ms blocks. The port of its
Go speech path to current TECHO5 source has passed initial local tests.
**This is not yet a qualified on-device release.** An installation mechanism,
Bionic runtime activation, IPC supervision, atomic rollback and device
long-duration testing remain outstanding. Creating the
`/data/techo5-linux/fireos-asp.enabled` marker without this work is unsafe.

No experiments should overwrite the official TECHO5 binary or modify
factory partitions. Keep the official release and its recovery method.
All AEC, beamforming and wake-word quality claims remain provisional
until measured with appropriately labeled recordings.
