# Building FrankenASP Dev Edition on Ubuntu 26.04 LTS

This records a **fresh-host, independent** build/verification on `independent build host` (23–24 October 2026). It is **not** an instruction to flash a TECHO5 image or to copy Amazon binaries. The deliverable built here is the open-source AArch64 experimental `echod` daemon. The ARM32 native DSP worker is a **separate Android NDK build** (below); the owner's factory `libasp.so` is never built or distributed.

## Tested host and dependencies

| Item | Independent build host |
| --- | --- |
| OS | Ubuntu 26.04 LTS (Resolute Raccoon), x86_64 |
| Kernel | 7.0.0-34-generic |
| Hardware | 8 logical CPUs, 31 GiB RAM |
| Git | 2.53.0 |
| Go | **1.26.0** linux/amd64 (official archive, downloaded independently) |
| Python | 3.14.4 (standard library sufficient for 31 installer/protocol tests) |
| GCC / Clang | GCC 15.2 / Clang 21 present, not needed for `CGO_ENABLED=0` ARM64 daemon |
| Auxiliary tools | Git, curl, CA certificates, file, binutils/readelf, e2fsprogs/debugfs, SSH and tar installed; **patchelf was absent on this build-only host** (required for the separate owner-backup installer/staging workflow) |

On a normal Ubuntu host, the prerequisite packages are:

```sh
sudo apt-get update
sudo apt-get install -y git curl ca-certificates python3 file binutils e2fsprogs patchelf openssh-client tar
```

`gcc` and `make` are useful for broader development but are not necessary for the pure-Go cross-build. Do not interpret this package list as permission to install FrankenASP on an unsupported device.

## Install a verified Go toolchain

The Go version is dictated by `echod/go.mod`: `go 1.26.0`. An ordinary `go` executable was **not** preinstalled on `independent build host`. For a user-local installation:

```sh
mkdir -p "$HOME/.local/opt"
cd "$HOME/.local/opt"
curl -fL --retry 3 -O https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
echo 'aac1b08a0fb0c4e0a7c1555beb7b59180b05dfc5a3d62e40e9de90cd42f88235  go1.26.0.linux-amd64.tar.gz' | sha256sum -c -
tar -xzf go1.26.0.linux-amd64.tar.gz
export PATH="$HOME/.local/opt/go/bin:$PATH"
go version
```

The SHA-256 above was obtained from the official `go.dev/dl/?mode=json&include=all` release metadata and verified against the downloaded 66,768,540-byte archive.

**Remote-executor caveat, not a tar or Ubuntu issue:** during the original environment setup, GNU tar 1.35 failed with `Function not implemented` **only when run through a restricted remote command-execution service bridge**. A follow-up A/B reproduction on **the same `independent build host` host, using the same tar archive**, succeeded under the regular Desktop Commander service. `strace` isolated the failing syscall to `openat2(..., RESOLVE_BENEATH) = -1 ENOSYS` in the restricted bridge (3 seccomp filters, `Seccomp: 2`), whereas it succeeded under the regular service (`Seccomp: 0`). This is a **sandbox/syscall-filter limitation of that particular remote executor**, not evidence of a GNU tar, archive, filesystem, or Ubuntu 26.04 defect. **Normal users should use the standard `tar -xzf` command above.** A Python `tarfile` extraction was used only as a temporary diagnostic workaround in the restricted bridge and is **not required** for FrankenASP builds.

## Clean source checkout and tests

```sh
git clone --single-branch --branch experimental/athena-factory-asp \
  https://github.com/dmit2k/techo5.git frankenasp
cd frankenasp
test -z "$(git status --porcelain)"
python3 -m unittest discover -s tools/frankenasp/tests -q

cd echod
go test -p 3 ./... -count=1
go test -p 3 -tags dot ./... -count=1
go test -p 3 -tags spot ./... -count=1
go test -p 3 -tags panel ./... -count=1
GOOS=windows go vet -p 3 ./cmd/techo5-deck
GOOS=darwin go vet -p 3 ./cmd/techo5-deck
cd ..
```

`-p 3` bounds package-build concurrency; it is not required for functional correctness. Keep diagnostics/logs and downloaded toolchains **outside the source checkout** so that the release builder can stamp a clean commit instead of `.dirty`.

## Reproducible AArch64 experimental daemon build

```sh
# Use the official TECHO5 release already installed on your Athena as the base.
bash tools/frankenasp/build_experimental.sh v1.6.0 ./techo5-experimental
file ./techo5-experimental
readelf -h ./techo5-experimental | grep -E 'Class:|Machine:'
sha256sum ./techo5-experimental
git status --porcelain
```

Expected architecture: **ELF64 AArch64**. The daemon's software version retains a TECHO5-compatible base identifier `v1.6.0_frankenasp.<source-commit>`. The standalone research release label **FrankenASP 1.0 Dev Edition** is *not* the upstream TECHO5 version. The Git state must be clean and the version must **not** contain `.dirty`. The recorded hash of any compiled binary applies to that exact build; timestamps embedded in Go linker flags can change binary hashes between otherwise identical builds.

## ARM32/Bionic worker and device-side components

The open worker source `tools/frankenasp/asp_stream_worker.c` is compiled independently using the **Android NDK ARMv7 API 28** toolchain, not host glibc:

```sh
export ANDROID_NDK_HOME=/absolute/path/to/android-ndk
bash tools/frankenasp/build_worker.sh ./libasp_stream_worker.so
readelf -h ./libasp_stream_worker.so | grep -E 'Class:|Machine:'
```

The ARM32 worker requires an independently installed official Android NDK; it is **not** produced by the Go-only build above. Private factory binaries, original Bionic runtime files, DSP configurations, factory beamformer coefficients and user recordings are **not** included in this repository or a public release. Use the owner-specific read-only preflight, explicit `--install`, guarded activation and rollback procedures in [installation.md](installation.md).

### Independent native-worker build actually performed on `independent build host`

The additional **ARM32/Bionic native worker** was successfully built using **Android NDK r27d (27.3.13750724)**, downloaded from Google's official Android repository. Google's release checksums list the `android-ndk-r27d-linux.zip` size as **663,956,036 bytes**, SHA-1 **`22105e410cf29afcf163760cc95522b9fb981121`**, which was verified after the download. No proprietary Amazon files were needed.

The output was verified as `ELF32 ARM EABI5`, API 28, built by NDK r27d, and the worker's glibc-reference guard passed:

```text
libasp_stream_worker.so SHA-256:
7017258766e7496b088a8b11b06b7b5ed11ae647c55831c3ee58aacab47dd649
```

For normal Android NDK setup, extract the official ZIP with `unzip`, which preserves executable permissions and symbolic links. During this independent validation, a Python `zipfile` extraction was used instead; because that API does not automatically restore Unix executable bits and symlinks, those attributes were subsequently restored from the ZIP metadata before running the NDK compiler. **This is not an Android NDK or FrankenASP problem, nor is Python ZIP extraction a recommended installation method.**

**Build execution resources:** the independent host has 31 GiB RAM, but its private private infrastructure execution bridges imposed **768 MiB service cgroup memory limits**, which caused an unrelated Go compiler process to be killed by the environment. The final Go CI therefore runs through the already connected Desktop Commander service on the **same `independent build host` host** without that artificial 768 MiB limit. This is an execution-adapter limitation, not a functional or hardware compatibility claim.

## Scope and safety

Supported research prototype: **Echo Show 8 Gen 2 (Athena)** with the specifically validated original factory DSP revision and TECHO5 AArch64 baseline. Short guarded runs, native PCM continuity, a persistent startup probation/stock fallback, and controlled **wake-word-only** acoustic comparisons have been demonstrated. Published mono/stereo modes describe the Echo's **own background-music speaker output**, not its microphone channel count. Post-wake STT, longitudinal stability, false-positive qualification, alternate factory firmware and other Echo models remain unverified.

See [initial acoustic results](acoustic-results-20261023.md) and [release notes](release-v1.0.0-dev.1.md).
