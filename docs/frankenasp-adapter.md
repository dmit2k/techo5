# FrankenASP adapter for TECHO5 (Athena)

This branch contains **only the Go adapter** for the independent [FrankenASP research and benchmarking toolkit](https://github.com/dmit2k/FrankenASP). FrankenASP's primary goal is to enable **repeatable objective measurements and subjective listening comparisons** of the open TECHO5 and original factory Amazon microphone-processing paths **on the same Athena hardware**. The worker and adapter are comparison-enabling tools, **not a proposal to replace TECHO5's audio frontend**. The standalone FrankenASP repository owns the open ARM32/Bionic stream worker, guarded installation and rollback scripts, tests, and research/acoustic documentation.

**This is a working Athena-specific research proof of concept, not yet a neutral, stable upstream plugin API.** A future vendor-neutral interface might benefit TECHO5 without requiring Amazon binaries; proposing that interface is optional and separate from comparing the two frontends today.

## Scope

The changes relative to the fork's `main` are restricted to:
- `echod/internal/hardware/mic/{athena,mic}.go`: capture of native four-mic PCM and physical stereo AWB reference, routing of factory-produced mono 16 kHz PCM into the existing wake-word and STT path without repeating TECHO5 AEC/denoise
- `echod/internal/hardware/mic/factory_asp_athena*.go` and `factory_pcm_probe_athena*.go`: guarded worker process, framing and timeouts; associated Go tests
- `echod/internal/update/{frankenasp_guard*,install}.go` and `echod/internal/feature/firmware/auto.go`: safety and update protection for opt-in experimental persistence

There are **no Amazon libraries, DSP configurations, factory runtime blobs, extracted backups, worker binaries or installer scripts** in this branch.

## Build together with FrankenASP

```sh
git clone --branch frankenasp-adapter https://github.com/dmit2k/techo5.git techo5-adapter
git clone https://github.com/dmit2k/FrankenASP.git frankenasp
cd frankenasp
python3 -m unittest discover -s tools/frankenasp/tests -q
export TECHO5_ADAPTER_ROOT="$(cd ../techo5-adapter && pwd)"
bash tools/frankenasp/build_experimental.sh v1.6.0 ./techo5-experimental
ANDROID_NDK_HOME=/path/to/android-ndk bash tools/frankenasp/build_worker.sh ./libasp_stream_worker.so
```

The TECHO5 release parameter must match the specific official base on the target Athena. The installation requires the owner's original firmware backup and factory `libasp.so` resident on the device; refer to FrankenASP's preflight and rollback guide.

## Current process/PCM contract

- Hardware input: four-microphone S32LE 96 kHz, dual-channel physical AWB reference S16LE 96 kHz
- Owner-local runtime: `/data/techo5-linux/amazon-arm32/` with native Bionic shell and independently built `libasp_stream_worker.so`
- Start handshake `ASPREADY`, fixed PCM packet framing via stdin/stdout, deadlines/fault handling
- Mono PCM16 16 kHz output to the standard TECHO5 voice chain
- An explicit owner-selected flag plus ephemeral arming marker is required; persistence and update fallback remain guarded

The interface and hardware behavior are **unchanged** from the [combined Dev Edition 1 snapshot](https://github.com/dmit2k/techo5/tree/snapshot/frankenasp-v1.0.0-dev.1). That snapshot remains the historical acoustic-test provenance; the split has not yet been acoustically retested on the device. The preliminary advantage was measured **only for wake-word activation while the Echo played music on its own speakers**. Here mono/stereo refers to **that speaker music playback**, not microphone array operation. Actual effective MWW thresholds differed (stock 0.75, FrankenASP 0.85); false-positive trade-offs, post-wake STT, and independent subjective speech-listening comparisons were **not qualified**.

## Verify

```sh
cd echod
go test ./...
go test -tags dot ./...
go test -tags spot ./...
go test -tags panel ./...
GOOS=windows go vet ./cmd/techo5-deck
GOOS=darwin go vet ./cmd/techo5-deck
```

A future upstream-neutral audio-frontend contract would abstract PCM formats, mic/ref streams, timestamps, lifecycle, health/fallback and provider selection. **It has not been implemented yet.**
