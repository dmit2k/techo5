# Stage07 Retake — финальная программная приёмка TECHO5 / FrankenASP V3
> **COORDINATOR STATUS, 2026-10-31:** retake software execution PASS; delegated independent technical acceptance ACCEPTED; code integration unchanged (Stage07 verification-only); all S01–S10 accepted for **software only**. Original executor statuses PENDING below are historical until coordinator §9. Hardware/vendor runtime/acoustics NOT RUN.

**EXECUTOR HANDOFF для документационного координатора. Отдельный Stage07 retake, не опубликованный handoff и не решение владельца.**

## 1. Идентичность и статусы

- Дата 2026-10-31 Europe/Riga; выполняла на anfisa-trainer через Desktop Commander. TASK: `stages/07_SOFTWARE_ACCEPTANCE_TASK.md`.
- Execution / final software gate: **PASS**. Owner acceptance Stage07: **PENDING**; retake docs publication: **PENDING**.
- Принятая входная Stage06-S07 correction: independent delegated technical ACCEPTED + integration COMPLETE, проверено в опубликованном correction handoff §9.
- Docs remote HEAD: `202a6d0c09b6358bcf0e466b526360eb85344f3f`; действующий V2 SPEC Git blob: `f3e1b00d1f5d184e345c0d2c79da8394435ba3a4`.
- Hardware / acoustics / vendor runtime on Echo / privileged launcher namespace: **НЕ ВЫПОЛНЕНЫ** (отдельное поручение владельца). Ни запуска Echo, ни vendor libasp, ни деплоя.
- Code integration action Stage07: **не выполнялась**. Input = output, ни новых stage refs/commits, ни изменения main/dev-main/docs.

## 2. Go/C immutable pair и worktrees

| Field | TECHO5 Go | FrankenASP V3 C |
|---|---|---|
| Repo | dmit2k/techo5 | dmit2k/frankenasp-v3 |
| Integration ref | experimental/external-asp-v3-dev-main | experimental/v3-dev-main |
| Input / tested / unchanged output | `71788409df4d9eacef21816a8f82fc545a6f4ff1` | `f6266907edb90e7c9c2f414e8e46e6f215a21b7b` |
| Tree SHA | `0b55a3d840fd2f21238ee11f82313be1c3489965` | `d79887b4cd9944759748dc83378204d34a7c8bf2` |
| Detached retake worktree | `/home/gptlab/dev/techo5-v3-stage07-retake` | `/home/gptlab/dev/frankenasp-v3-stage07-retake` |
| Stage07 branches / commits | none | none |
| Frozen main unchanged | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` | `e986ae18643e08e95f29a38aae1dd74762f6362e` |

- Prior accepted correction changed **only** Go test `echod/internal/hardware/mic/asp_control_test.go` (+18/-2) to align full physical Channels with MIC/REF and assert both REF L/R; production untouched.
- C mock tests were built from a 52-file git-archive source copy of exact C commit outside tracked C worktree; full SHA256 check `receipts/c-archive-source-validation.log` / `receipts/c-checkout-hashes.txt`.
- Input/postflight Git and unchanged old FAIL doc blob: `receipts/immutable-input-receipt.txt` / `receipts/FINAL-POSTFLIGHT.txt`.

## 3. Joint contract and toolchain identities

- Descriptor PCM-only ASP1 16-byte header + 28-byte physical buffer record + 8-byte channel records. Exactly one readiness byte 0xA5; 640-byte mono S16LE/16kHz/320-sample output; common 5s PCM exchange and one 10s startup, no extra metadata/IDME.
- Shared independent Go/C descriptor fixture SHA256: **9d26d6cca1e29b10f1abaee583da12ee918ea4252f16f70c3eb2c01531e1d2f4** (each); readiness.bin SHA256 **6922e93e3827642ce4b883c756b31abf80036649d3614bf5fcb3adda43b8ea32**.
- Go session is production OS-pipe process implementation; C runtime compiled with test-only mock_vendor / ASP_TEST_ONLY, compiled independent decoder and io_faults.so; no physical vendor ABI qualification inferred.
- Go 1.26.9 linux/amd64, host GCC 13.3.0, ARM32 cross GCC 13.3.0; NDK r28c / 28.2.13676358, Android API 26 Clang 19.0.1, sourced `/home/gptlab/.config/techo5/toolchains/android-ndk.env.sh`. Toolchain provenance `receipts/toolchains.txt` and `ANDROID_NDK_TOOLCHAIN.md`.
- Both real Android ABI builds on this **corrected** exact Go SHA; GOOS=android CGO_ENABLED=1 -tags panel -trimpath -ldflags matching current upstream workflow; versions and commands below. Old NDK installation smoke-builds not counted as S08.

## 4. Changed files / LOC / negative scope

- Stage07 retake code changes **zero**. Go production vs T01: {'added': 1193, 'removed': 33, 'files': 11, 'net': 1160}; textual tests: {'added': 2852, 'removed': 0, 'files': 19, 'net': 2852}. Binary fixtures excluded from LOC.
- C production vs C01: {'added': 1032, 'removed': 0, 'files': 16, 'net': 1032}; test/fixtures: {'added': 1076, 'removed': 0, 'files': 20, 'net': 1076}; build: {'added': 100, 'removed': 0, 'files': 3, 'net': 100}.
- Actual file name inventory and changes: `receipts/LOC-FILE-INVENTORY.json`, `receipts/git-source-inventory.txt`. No full old subsystems deleted from clean baseline.
- Source review: `receipts/S09-source-review.txt`. V3 no CFG2/TLV, wire statuses/feedback, NotReady, ASP Debug, preflight, platform whitelist, inode cache, synthetic production DSP, async worker/queue/retry, generic managers or compatibility layer. Stock service retry/status and stock raw/call/DSP/Panel remain untouched.
- Only C launcher owns trusted vendor mount/aliases/FD3+FD4, not generic Go host; no installer/OTA/vendor blobs. Reuse provenance in LAUNCH_PROVENANCE.md and THIRD_PARTY_NOTICES.md (MIT reference historical FrankenASP, no merge).
- Historical minor gofmt-only warning in Go descriptor_test.go retained from Stage02-R, not part of Stage07 retake production. No hidden code fix.

## 5. Software S01–S10 matrix

| Gate | Verdict | Evidence summary |
|---|---|---|
| S01 | **PASS** | C01–C07; S16/S24/S32; >2 buffers; no REF; REF L/R; strict 16/28/8 and metadata-free; C ASan/UBSan; Go/C independent fixtures |
| S02 | **PASS** | W01–W10 production Go session + compiled C test-only runtime, real pipes/FD/stdout, startup >10s and common 5s deadline |
| S03 | **PASS** | N01–N08; SoftwareCut MIC and REF/stock, borrowed ownership and retained published frames |
| S04 | **PASS** | L03–L06: eight cancellation cases, stale response, close/Run/Rewire/XRUN, preWait shutdown, sole reaper |
| S05 | **PASS** | L01/L02/L09 Selected/Active, immediate persistence/HA/manual relaunch/logging |
| S06 | **PASS** | H01–H08 supported DSP mock/admission vs unsupported, 64 output-only initial padding, 8ms quantum and provenance |
| S07 | **PASS** | Five **full** Go suites default/dot/dot,donut/spot/panel, panel-debug, tagged TestL race and full relevant package race |
| S08 | **PASS** | Five vet, five Linux amd64 + ARM32 builds, production ARM32 C child and launcher, two **real** Android NDK API26 Panel ABI builds |
| S09 | **PASS** | Exact focused diff, minimalism, no legacy/forbidden subsystems, stock preservation and source provenance |
| S10 | **PASS** | SHA pair, commands/exits/log/binary SHA256, ABI/toolchain, 99 D/F/V/A traceability, LOC, remote & clean postflight |

### Exact executed commands and per-log SHA256

All commands run on exact pair stated above; commands/cwd/environment in runner `run-stage07-retake.sh` and ledger `results.tsv`. Log SHA256 manifests `receipts/LOGS-SHA256SUMS`.

| ID | CWD | Command | Exit | Log SHA256 |
|---|---|---|---:|---|
| `S01-C-conformance` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-src` | `make test` | 0 | `fbfae44fcffb0ef60cfca86722e132315d6ed28d28dafe2cac5365c880307906` |
| `S01-C-ASAN` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-src` | `make asan` | 0 | `c49f039ce755eec16d2e6693353252792dff5ab5205c94c36358be8bd21abe24` |
| `S01-C-UBSAN` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-src` | `make ubsan` | 0 | `e14f22abe7f6e1d5a80310d6b78c8cc8208f8b00e71f147d4180194d8013f68d` |
| `S01-fixture-parity` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-src/testdata` | `cmp SHA256SUMS '/home/gptlab/dev/techo5-v3-stage07-retake/echod/internal/asp/testdata/SHA256SUMS' && sha256sum -c SHA256SUMS` | 0 | `b2c29c181614052cfa88bd2d79444a70c0f6257fc19c8e4eccac4f9a13f76737` |
| `S01-Go-codec` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 ./internal/asp` | 0 | `5b86ca69a788e35081807b9758a4acf0f94bc31dee5a5d8d077a31d00c08dde6` |
| `S02-compiled-joint` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 -v ./internal/asp` | 0 | `8cd61112efdeb0e9d86d9e7ebd9d4bdae27a25f8b862d2bd77e5f71e07b737a3` |
| `S03-native-privacy` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 ./internal/hardware/mic ./internal/hardware/privacy ./internal/layout ./internal/lib/alsa` | 0 | `39defe51e57d76c390036e14ade63200d3dbeb6226f2e28747a93734dd55023e` |
| `S04-lifecycle` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1 -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 | `b2d571adc570a1e619e5f4f3fe4c3cd5c95a583f2aa3f92d0515e75e5347ea01` |
| `S05-selected-active` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 ./internal/feature/microphone ./internal/config ./internal/boot ./internal/service ./internal/component/all` | 0 | `61b7a13c9ee896c22013eaad8d976ae05675d24aad5a193d765f04e1114fb5d9` |
| `S07-full-default` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1  ./...` | 0 | `a0273edee5d7a5cc5ef2ec352cb77d8c4928f671f28bc1950c461c54f592fcfe` |
| `S07-full-dot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 -tags dot ./...` | 0 | `f0d1413cb871b9ddcde1e81484339dc23a49d5a79c783ff443e74a27cce3c285` |
| `S07-full-dot_donut` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 -tags dot,donut ./...` | 0 | `aa1c2d765516da22db9102100eb81337ab47286fc20b6f38e6d4ebb0900c5350` |
| `S07-full-spot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 -tags spot ./...` | 0 | `d9784a6cefac859fc77711ced23dfa8aa4400e9765b93c0e2bd75ce198380ddf` |
| `S07-full-panel` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 -tags panel ./...` | 0 | `5f2b669d4fdd02725766f8ea296379af244b713e9bf130193fc7e165bca8f118` |
| `S07-panel-debug` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -count=1 -parallel 1 -tags panel,panel_debug ./internal/lib/loopback` | 0 | `72abd94ba04ecc7fc0b7ef8a6453f9a85e685aafd1c283d6ce9075fee7bd23be` |
| `S07-race-L-default` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1  -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 | `fc371020d7558b9b16b9cca906a9d45aa98492708ddfd420a2e0a750102fb9a5` |
| `S07-race-L-dot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1 -tags dot -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 | `a68654ea852a7070bc449bc860c502385e65e8eb090d1b561b98e335ccabaa5e` |
| `S07-race-L-dot_donut` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1 -tags dot,donut -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 | `f16e42a2cc88b2cb2d4b2ff9593b64b24a28c1eb9820f0d70bc94fab68124f7c` |
| `S07-race-L-spot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1 -tags spot -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 | `2a1e3a34d122090976c1723a1a59bf6f8ef49df62eb711b01080a1a0c661a8fe` |
| `S07-race-L-panel` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1 -tags panel -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 | `d1fafd60fbfa20b521f96e5de02c48467f1c7b7290e3639b94346f6e02d6ec7d` |
| `S07-race-relevant-default` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1  ./internal/hardware/mic ./internal/feature/microphone ./internal/boot ./internal/service ./internal/asp ./internal/config ./internal/component/all ./internal/feature/phone ./internal/hardware/privacy ./internal/layout ./internal/lib/alsa` | 0 | `205fae41c710008c499ac0298c14474e07521f0f890bef3cf6d6760db9ce7260` |
| `S07-race-relevant-dot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go test -race -count=1 -parallel 1 -tags dot ./internal/hardware/mic ./internal/feature/microphone ./internal/boot ./internal/service ./internal/asp ./internal/config ./internal/component/all ./internal/feature/phone ./internal/hardware/privacy ./internal/layout ./internal/lib/alsa` | 0 | `57d4d8521bce7ca9d719ead3746d86aad3207f069aeb109557523de4413c690b` |
| `S08-vet-default` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go vet  ./...` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-amd64-default` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-amd64-default' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-arm32-default` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build  -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-arm32-default' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-vet-dot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go vet -tags dot ./...` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-amd64-dot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags dot -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-amd64-dot' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-arm32-dot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags dot -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-arm32-dot' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-vet-dot_donut` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go vet -tags dot,donut ./...` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-amd64-dot_donut` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags dot,donut -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-amd64-dot_donut' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-arm32-dot_donut` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags dot,donut -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-arm32-dot_donut' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-vet-spot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go vet -tags spot ./...` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-amd64-spot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags spot -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-amd64-spot' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-arm32-spot` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags spot -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-arm32-spot' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-vet-panel` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `go vet -tags panel ./...` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-amd64-panel` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags panel -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-amd64-panel' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-arm32-panel` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags panel -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/echod-linux-arm32-panel' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-C-bionic` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-src` | `LD_LIBRARY_PATH='/home/gptlab/dev/techo5-stock-audio-audit-20261021/arm32-tools/binutils-unpacked/usr/lib/x86_64-linux-gnu' BIONIC_CC='/home/gptlab/dev/techo5-stock-audio-audit-20261021/arm32-tools/cross/usr/bin/arm-linux-gnueabihf-gcc-13' BIONIC_LD='/home/gptlab/dev/techo5-stock-audio-audit-20261021/arm32-tools/binutils-unpacked/usr/bin/arm-linux-gnueabihf-ld' BIONIC_ROOT='/home/gptlab/dev/techo5-stock-audio-audit-20261021/factory-arm32-root' OUT_DIR='/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-bionic' bash build_bionic.sh` | 0 | `78a1c42007830259c488723692819e1ad46a8195cf30af6f2543b5cf1f052ae0` |
| `S08-C-launcher` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-src` | `LD_LIBRARY_PATH='/home/gptlab/dev/techo5-stock-audio-audit-20261021/arm32-tools/binutils-unpacked/usr/lib/x86_64-linux-gnu' ARM32_CC='/home/gptlab/dev/techo5-stock-audio-audit-20261021/arm32-tools/cross/usr/bin/arm-linux-gnueabihf-gcc-13' ARM32_BINUTILS='/home/gptlab/dev/techo5-stock-audio-audit-20261021/arm32-tools/binutils-unpacked/usr/arm-linux-gnueabihf/bin' ARM32_SYSROOT='/home/gptlab/dev/techo5-asp2-f2b-minimal-20261028/private/toolchain/sysroot' OUT_DIR='/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/c-launcher' bash build_launcher.sh` | 0 | `50c7c6f3a09d917838617f3680bf710bf7373cf208a622dcfd585b0f325e54a7` |
| `S08-android-arm64-v8a` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC='/home/gptlab/Android/Sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android26-clang' go build -tags panel -trimpath -ldflags '-s -w -X github.com/HuskerMinion/techo5/echod/internal/layout.Version=v0.0.0-dev' -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/libechod-arm64-v8a.so' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-android-armeabi-v7a` | `/home/gptlab/dev/techo5-v3-stage07-retake/echod` | `CGO_ENABLED=1 GOOS=android GOARCH=arm GOARM=7 CC='/home/gptlab/Android/Sdk/ndk/28.2.13676358/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi26-clang' go build -tags panel -trimpath -ldflags '-s -w -X github.com/HuskerMinion/techo5/echod/internal/layout.Version=v0.0.0-dev' -o '/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/go-builds/libechod-armeabi-v7a.so' ./cmd/echod` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `S08-binary-abi` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031` | `find go-builds c-bionic c-launcher -type f -exec file {} \\; -exec sha256sum {} \\;` | 127 | `af56c5050966d2e8ba253ad113b341ebddb2400602fe70fc3ef01efed9908a65` |
| `S08-binary-abi-corrected` | `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031` | `find go-builds c-bionic c-launcher -type f -exec file {} \; -exec sha256sum {} \;` | 0 | `254ddb845a34e31b9982f49674f08a57dcc659b6aa3e074ea64e680b18d76811` |

### Actual built artifacts (new retake binaries)

| Binary | SHA256 | ABI identified by file |
|---|---|---|
| `c-bionic/asp1-child` | `86d4f4d7827129e944c76187dc4575ef7bc5240f2b99f048e7860ed71057fa05` | ELF 32-bit LSB pie executable, ARM, EABI5 version 1 (SYSV), dynamically linked, interpreter /system/bin/linker, not stripped |
| `c-bionic/asp1-child.elf.txt` | `437e04305109e3248c900c4b140bb886632f748a0e934aa102084e849c21881d` | ASCII text |
| `c-bionic/objects/bionic_entry.o` | `1d590342445978fea6b4996d6cf2fce4bb56c6314372753d1bb7e5708b7fd1b7` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-bionic/objects/child.o` | `a26eb288d1ddea9bb5e7120d61fef2171948a27080e3449963a8dba9a78ee8ed` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-bionic/objects/descriptor.o` | `181645f7f6733a6a2b93356a3fb202e28bff0fdd4effa8360c0601a838d99ed2` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-bionic/objects/factory_adapter.o` | `229e429d4c1ff9daaf7caf2c7ea24cc552f2950bb84077c61c193faad7c47822` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-bionic/objects/io.o` | `4571c74121657e287f7760771fa17ef0828ee5559ebe9ce2f8624138bec2d448` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-bionic/objects/runtime.o` | `0c73fa0648e54497fe6cd9f57668c7e812958f0e5a445c35472330f87edf3a19` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-bionic/objects/start32.o` | `f7b6127e70f57099294bba8156ef7d36ead00e21c0930d9c9d33e16c95e8760c` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-bionic/objects/vendor_factory.o` | `fbbe48cd23cee7959f9133dca0e06ff273a3449674c4036b283a2762c0e31c63` | ELF 32-bit LSB relocatable, ARM, EABI5 version 1 (SYSV), not stripped |
| `c-launcher/asp1-launcher-arm32` | `fce1f3e1abc427d7e5402da5c56dd530d5bf2acd7686b4f3aba98d1445b24ab7` | ELF 32-bit LSB executable, ARM, EABI5 version 1 (GNU/Linux), statically linked, BuildID[sha1]=22c6ff96a024c37032154340819c55da52f7908d, for GNU/Linux 3.2.0, not stripped |
| `c-launcher/asp1-launcher.elf.txt` | `cd34f04059fbb16ef18321ff56219857551f94889e0906b24f56b5396d7bb853` | ASCII text |
| `c-src/build/codec_probe` | `d38cd6ae23d213595c98c5a2c24f932d6577a6b94cbc29fc45a216c71482da38` | ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, BuildID[xxHash]=0990363cc659c84d, not stripped |
| `c-src/build/codec_probe_asan` | `5616eabfd26852e680250e75b57433afef8c01856588473106d9b9454395b7da` | ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, BuildID[xxHash]=b6121e7a11cec4f1, with debug_info, not stripped |
| `c-src/build/codec_probe_ubsan` | `acb1846b0b2001ed0e4407083d2fc3f875c488cc258d354683553bde1be79348` | ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, BuildID[xxHash]=e873f7bccadb5c81, with debug_info, not stripped |
| `c-src/build/io_faults.so` | `cce44e3ee3e2355f941b1763e15a43036a1ce2e84ec6ac7c1665adc237f70559` | ELF 64-bit LSB shared object, x86-64, version 1 (SYSV), dynamically linked, BuildID[xxHash]=10fcf2bd3ea1702a, not stripped |
| `c-src/build/runtime_mock` | `0fdcc3d16713d7396dcfd687cb6c780d3581833aad38278a2c584db45172c8c7` | ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, BuildID[xxHash]=f1ef3217dd2a028e, not stripped |
| `c-src/build/runtime_mock_asan` | `bf620d1c915c443924796dab850d2ad332aca1f8c3ba00570e8f149536b0ce83` | ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, BuildID[xxHash]=1dc1183f009f0ef4, with debug_info, not stripped |
| `c-src/build/runtime_mock_ubsan` | `ee9bf7db1a1af09a904512a9b4d714f918c57822e1ff594df3dc93444d2b0fe2` | ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, BuildID[xxHash]=bbcb4d306e9e2b93, with debug_info, not stripped |
| `c-src/build/test_alias_policy` | `afde45504b46981175e072d8b280287e453a6d3b0caed1b5dfe5894a49cb4f68` | ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, BuildID[xxHash]=aaa94fedec6afb0b, not stripped |
| `go-builds/echod-linux-amd64-default` | `48fb4b1193608c62d476e92ec877a0eab15e1cd56297cb55a3507770ed6ac926` | ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=3K5WeBJosu3uyWs1PRgS/anuR6RT1yQe2sPyXTtI3/9JQFzuHFlKjBlp9ta6Mv/ob-KABKiqCuva6SYys5Z, BuildID[sha1]=962a7c2530bee1d92edb39708197ab6c591c5402, with debug_info, not stripped |
| `go-builds/echod-linux-amd64-dot` | `346dfcc94090bf5d4020ce3a1218b90fed08dbe7c4c33c2e7ebc496bec860b98` | ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=iRswHmKRiR6a0xwHYY6r/2kW0gBuUakXTmwlvBLbY/tGL7ydXXqqaCqE4pXWIB/oS5v6ApcTVW9dTHFBzY_, BuildID[sha1]=34f0125fe452f43f06dbf3c447226eea40bc93bd, with debug_info, not stripped |
| `go-builds/echod-linux-amd64-dot_donut` | `848d0a7b2dc286cfa4aebf39d071f9b65f9384f26e7f78a80af2b3d8b8ddc6ec` | ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=mqsFVpnLhzGMfmxQTaol/iK5UvZCS4mJijWU4fqFt/NEPlbb5OqHfFxPIbNK3g/tX3DwCjgSlcWkZDZ-VP3, BuildID[sha1]=9d23a078300fd88ab34691e21be787dd5ad9a212, with debug_info, not stripped |
| `go-builds/echod-linux-amd64-panel` | `e8616deb563493ba97bc8857eed716ff5936b99f4fd815c2254e6eba87adaf3b` | ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=-78gFpfGXHDmbjO9yep5/VjKVsnv28KZbNl8RZYol/ppaKrJMVkMJv_4t4ixD2/2NSp2fMqhU0V4eKrvqQL, BuildID[sha1]=6bcb4e9ce1f23af2d12df8aa6edb14bed27d7736, with debug_info, not stripped |
| `go-builds/echod-linux-amd64-spot` | `59f9919d21302a9d7da6381087264b7f26ed3ee1b0e6ae04840483553ad6d3cf` | ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=j4RoWcd0stkZRs2KEol2/doH_eKU11HNp9lu9_CNh/vGnHf-7XYONWdry6Epzk/9pKwWxxxNt3s2tWmGDt-, BuildID[sha1]=c54d9395f5ad2f7eb2c5015a8bafa335378b947f, with debug_info, not stripped |
| `go-builds/echod-linux-arm32-default` | `2452b734a42c0bda075c0b60ba5bb45ed3215ccace65bfd25ccc344c5672c1a1` | ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, Go BuildID=8VFc3ulsKbRxb7TouJhN/0v-1LTW_MREiPe9v9mBO/9lFMjDLFzvt5H_apexzD/qWu8ALXFxYQFYbyHxH5n, BuildID[sha1]=5b241d36f02892815245020a65e9941870152230, with debug_info, not stripped |
| `go-builds/echod-linux-arm32-dot` | `3f0faf917b5fd8d90297ea798ce4461ee78a446dc9b9f7cc81d014147969338d` | ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, Go BuildID=iUY7eW1zmkgyXhpsf8CO/YEoTV_QTHx4DUCLcUiY9/zG6Q6ARJevp_FqWAi6Gl/fMj9-R8OZf7BO90zg1Ic, BuildID[sha1]=87548d335ec127e7309ec871f57167968adcc96b, with debug_info, not stripped |
| `go-builds/echod-linux-arm32-dot_donut` | `65a1767c038ce019cdf02f550c1cc9a0fb160b66552fdda8da646bc78b91766d` | ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, Go BuildID=lqaSfiIgDXGfYGCsloYz/ZX-QL3VBFYiHSEgNxBOU/FpPtluLufj8Exe27T1AT/Gj8LyS36PMiqfc8GAmLK, BuildID[sha1]=3fedebf70714f9fe34f3be9cedb8b61dd39872a1, with debug_info, not stripped |
| `go-builds/echod-linux-arm32-panel` | `391cc1919ef9035eb91b3d0099120d714bf222e9d15f51659f730d96cb800c42` | ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, Go BuildID=mA5UzuVI795ry5qZDLNK/aghjKAy1sIYc-CZ-S3gt/SLBoF0w2EENgPQS5MCY4/dAaRax61p7jt0eKrR2fx, BuildID[sha1]=3951da9eb4311a8feeaa6d44077ae12c73c56627, with debug_info, not stripped |
| `go-builds/echod-linux-arm32-spot` | `ae464b0e1bd4ed83c755432c2e961d4eff6ecfa6d23ece22c567e9e137a842a2` | ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked, Go BuildID=VofO0ErtXRYtgxvRf-hA/ZT1ZwGvoepcCoCf33UpT/KWhI7OJWRKK3kx8XyYA6/InHqVI6IIw68hsJi9L7e, BuildID[sha1]=85173897649ca6a722fcdd3ae2748be8a3e296bc, with debug_info, not stripped |
| `go-builds/libechod-arm64-v8a.so` | `ea42091b7b1c342d3a1720a5f0237b29d129166e4b0c12494f9dea0d156a5937` | ELF 64-bit LSB pie executable, ARM aarch64, version 1 (SYSV), dynamically linked, interpreter /system/bin/linker64, Go BuildID=uIWuTSrakz4byD7MXNas/wK-5pYNaXl_pib56WIso/I0ESZ8I0Umg-rj-mpa_k/6iFhdm1r6YoV3rDUM3vg, BuildID[sha1]=5df20ea0a9493cc1072c0af1a83f3db043cb1fed, stripped |
| `go-builds/libechod-armeabi-v7a.so` | `3e211441de2b844a5faa53259305129835ae8693ce452b2a132295e699c64bd3` | ELF 32-bit LSB pie executable, ARM, EABI5 version 1 (SYSV), dynamically linked, interpreter /system/bin/linker, Go BuildID=754qBLgay8Uuyy-AqtZC/xWKGzqiC24wJSSdKRwt4/cxusang4e2jIz6a_JHl9/0-kCW9nHqsBTJFYLGQff, BuildID[sha1]=2fe69d87966494950c8317a5a72cd03204835ed1, stripped |

## 6. Risks / unexecuted items / STOP

- Execution PASS; required check IDs absent: []; nonzero command IDs: []. No mandatory software item may be marked PASS without fresh execution.
- Hardware/acoustic/vendor live tests were **NOT RUN by design**; compiled mock, ARM32 crossbuild and Android ABI binaries do not qualify proprietary DSP, factory calibration, privileged target runtime or acoustics.
- Legacy old FAIL handoff stays unchanged; software PASS does not constitute owner acceptance. No unapproved code modifications, new stage commits or automatic deployment.
- One optional ELF/file inventory helper syntax error (if present in ledger) is separately documented in original logs; corrected command must return exit0 and its log is preserved before accepting ABI inventory: ['S08-binary-abi'].

## 7. Next authorization

- Owner/coordinator may conduct independent technical acceptance and decide later PR/software review or separately assigned hardware checks. No automatic Stage08/PR/release/deploy/hardware.
- Code output integration remains unchanged: `71788409df4d9eacef21816a8f82fc545a6f4ff1` / `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`.

## 8. Coordinator publication boundary

- Executor report: `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/07_RETAKE_HANDOFF_EXECUTOR.md`. Evidence manifests and 99 requirement mapping: `receipts/BINARIES-SHA256SUMS`, `receipts/LOGS-SHA256SUMS`, `receipts/TRACEABILITY_99.md`, `EVIDENCE-SHA256SUMS`.
- Publish only after independent coordinator verification to new `stages/07_RETAKE_HANDOFF.md` on docs branch; original `stages/07_HANDOFF.md` left unchanged. Docs publication SHA cannot be predicted. Owner acceptance remains PENDING.

## 9. Coordinator independent software acceptance and immutable-pair receipt (2026-10-31)

**Source-of-truth preserved:** executor original `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/07_RETAKE_HANDOFF_EXECUTOR.md` SHA256 **`a928640776f1cda06d53aef20a74850e65d991ba3477efe8233e7bbe354a8a10`** remains untouched. Entire 146-entry evidence index `/home/gptlab/dev/techo5-v3-stage07-retake-evidence-20261031/EVIDENCE-SHA256SUMS`, SHA256 **`183d1d6c9f3a84b296cd905d4c29a0b51391e261e5796d11d8a9038d73bdbfaa`**, independently verified against its referenced files. Original executor acceptance/documentation PENDING rows are historical, prior to this coordinator review.

**Independent technical verdict: ACCEPTED (Stage07 Retake software S01–S10 only).** This is a **delegated independent technical acceptance** after reviewing the source/traceability and repeating critical tests; it does **not** claim that the owner personally declared acceptance. Stage07 is verification-only; there is **no new code commit, stage branch, merge or FF**. The tested pair and pre-/post-check GitHub integration refs agree exactly:
- TECHO5 `experimental/external-asp-v3-dev-main`: **`71788409df4d9eacef21816a8f82fc545a6f4ff1`**, tree `0b55a3d840fd2f21238ee11f82313be1c3489965`.
- FrankenASP V3 `experimental/v3-dev-main`: **`f6266907edb90e7c9c2f414e8e46e6f215a21b7b`**, tree `d79887b4cd9944759748dc83378204d34a7c8bf2`.
- Frozen protected main refs still Go `62603a7854283a3a07473a17c831c6ef0fd6f8d4`, C `e986ae18643e08e95f29a38aae1dd74762f6362e`; the separate Stage06-S07 accepted correction receipt and old Stage07 FAIL handoff remain immutable. All checked code/docs verification worktrees clean.

**Independent test/evidence checks:** verified exact 57 Go/C source SHA256 entries across both working trees (source index paths are relative to their respective Go/C roots), both shared descriptor fixture manifests, the 43-row execution ledger and 43 per-command logs via the complete evidence manifest, all 32 binary manifest entries and original C archive source identity. The traceability table contains exactly **99 unique IDs**: **58 D, 5 F, 27 V, 9 A**, mapped to S gates. This table is traceability evidence, **not** an invented independent PASS per requirement. The S01–S10 PASS verdict is supported by executed tests, source review, actual build/artifact hashes and source/commit provenance. Original tests include all five complete stock Go suites, Panel debug, five relevant L-series tagged race suites, default/dot relevant race, full tagged vet/amd64/ARM32, real Go↔C compiled IPC and C mock ABI plus ASan/UBSan, H01–H08, privacy/native/reference/channel cases, and Source/HA/lifecycle/reaper requirements.

**Coordinator independently re-executed on the exact frozen pair:** compiled C `make test`, `make asan`, `make ubsan` **PASS**; compiled Go `go test -count=1 -parallel 1 -v ./internal/asp` against C runtime mock/io faults/decoder **PASS**; Go dot L-series MIC/HA/service race + targeted `go vet` **PASS**; full dot `go test -count=1 -parallel 1 -tags dot ./...` **PASS**. Coordinator log files are stored without changing tracked worktrees under `.../techo5-v3-stage07-retake-evidence-20261031/coordinator-review/`. Remaining four full tagged suites and comprehensive vet/race/build matrix were independently SHA256-validated from the executor receipts rather than individually re-executed again.

**Real Android ABI proof:** independent fresh `GOOS=android CGO_ENABLED=1 -tags panel` builds with installed official NDK r28c revision `28.2.13676358`, API26 Clang, current Go 1.26.9 on exact accepted Go SHA: both `arm64-v8a` and `armeabi-v7a` **PASS**. Independently built ELF outputs are **byte-identical** to Stage07 executor retake binaries. AArch64 interpreter `/system/bin/linker64`, 32-bit ARM `/system/bin/linker`, digest SHA256 **`ea42091b7b1c342d3a1720a5f0237b29d129166e4b0c12494f9dea0d156a5937`** and **`3e211441de2b844a5faa53259305129835ae8693ce452b2a132295e699c64bd3`** respectively. C production ARM32 Bionic child and static launcher build receipts/provenance/ELFs independently checked, production binary hashes **`86d4f4d7827129e944c76187dc4575ef7bc5240f2b99f048e7860ed71057fa05`** / **`fce1f3e1abc427d7e5402da5c56dd530d5bf2acd7686b4f3aba98d1445b24ab7`**. Build/ABI is **not actual Echo runtime operation**.

**Nonblocking ledger deviation disclosed:** executor's first `S08-binary-abi` inventory helper invocation exited **127** due to malformed shell `find -exec ... \;` escaping; exact corrected invocation `S08-binary-abi-corrected` exited **0**, produced the complete ELF inventory and verified artifacts. This is one failed auxiliary inventory command, not a failed required build gate, and is included in the signed evidence manifest; 42 other ledger entries exit0. No source or production code patch was made to compensate. There are no unexecuted mandatory software gate items.

**Final status:** **Stage07 Retake software execution PASS / independent delegated technical acceptance ACCEPTED**. **Code integration: UNCHANGED** (verification-only stage, no FF needed); clean refs and tested SHA pair confirmed. Historic initial `stages/07_HANDOFF.md` remains a valid record of the earlier S07/S08 FAIL/STOP; only the separate retake on the accepted Stage06-S07 corrected pair is PASS. **Echo hardware, privileged launcher/device runtime, proprietary `libasp.so`/Calibrator and acoustics: NOT RUN / NOT QUALIFIED.** This receipt is not authorization for deployment, protected-main merge, production PR, device access, software release or hardware tests; each needs its own explicit scope/decision.
