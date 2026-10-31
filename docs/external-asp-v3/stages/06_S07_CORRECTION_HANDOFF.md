# Stage06-S07 — Go-only correction of test PCM geometry — executor handoff

> **COORDINATOR FINAL STATUS (2026-10-31): Execution PASS / independent delegated technical acceptance ACCEPTED / integration COMPLETE.** The PENDING entries below are historical executor evidence. Final decision and remote SHA receipt are in §9.

## 1. Identity and acceptance boundaries

- Stage/TASK: **Stage06-S07**, `stages/06_S07_CORRECTION_PROMPT.md`; performed **2026-10-31T20:04:55+02:00** on `anfisa-trainer` via Desktop Commander.
- **Execution PASS:** original dot FAIL reproduced, test fixture corrected, **40/40** post-change validation commands exit **0**.
- **Owner / delegated independent acceptance: PENDING. Integration: PENDING.** This is the executor receipt, not a coordinator decision. No Stage07 retake or Stage07 software acceptance claimed.
- Docs verified HEAD `00886925a2620bc8850e2596707eaab01529be6a` on `dmit2k/techo5:experimental/v3-design-staging`. Applicable normative files read: EXECUTION_BRANCH_STRATEGY, MASTER_CONTRACT, PCM_ONLY_ARCHITECTURE_DECISION, SPEC_V2_APPROVED, Stage06 TASK and accepted Stage06 handoff §9, Stage07 FAIL handoff including §9, Stage07 TASK, HANDOFF_TEMPLATE. No applicable AGENTS.md in target ancestry.
- Actual SPEC V2 blob `f3e1b00d1f5d184e345c0d2c79da8394435ba3a4`, both `git rev-parse HEAD:path` and `git hash-object` equal. Normative docs, C, frozen mains and integration branches not edited.
- Software gate: **Stage06-S07 correction regression only**, not S01–S10 Stage07 final. Hardware / acoustic / Echo vendor runtime: **NOT RUN**.

## 2. Exact revisions / stage branch

| Field | Go: dmit2k/techo5 | C: dmit2k/frankenasp-v3 |
|---|---|---|
| Integration ref | `experimental/external-asp-v3-dev-main` | `experimental/v3-dev-main` |
| **Actual accepted input** | `cce2f220aeab9835e10a7f7ab659f5ca390011e8` | `f6266907edb90e7c9c2f414e8e46e6f215a21b7b` |
| Stage branch | `experimental/external-asp-v3-dev-stage06-s07fix` | **none** |
| Worktree | `/home/gptlab/dev/techo5-v3-stage06-s07fix` | `/home/gptlab/dev/frankenasp-v3-stage07` (detached, unchanged) |
| Commit tested | `71788409df4d9eacef21816a8f82fc545a6f4ff1` | `f6266907edb90e7c9c2f414e8e46e6f215a21b7b` unchanged |
| Tree tested | `0b55a3d840fd2f21238ee11f82313be1c3489965` | `d79887b4cd9944759748dc83378204d34a7c8bf2` unchanged |
| Accepted candidate | **PENDING** | **PENDING; C unchanged** |
| Output integration | **PENDING** (still `cce2f220aeab9835e10a7f7ab659f5ca390011e8`) | unchanged `f6266907edb90e7c9c2f414e8e46e6f215a21b7b` |
| Clean worktree | `git status --porcelain` empty | empty |

- Commit parent = exact **published integrated** Stage06 Go SHA `cce2f220aeab9835e10a7f7ab659f5ca390011e8`; original local Stage06 executor commit `cc725420...` **not** a parent. Commit tree is `0b55a3d840fd2f21238ee11f82313be1c3489965`.
- New Go stage only, no C stage branch and no C commit. Protected mains: Go `62603a7854283a3a07473a17c831c6ef0fd6f8d4`, C `e986ae18643e08e95f29a38aae1dd74762f6362e`, unchanged.
- Published correction Go stage SHA **`71788409df4d9eacef21816a8f82fc545a6f4ff1`**, tree **`0b55a3d840fd2f21238ee11f82313be1c3489965`**, checked by independent post-push `git ls-remote`, fresh `git fetch origin refs/heads/experimental/external-asp-v3-dev-stage06-s07fix`, remote `FETCH_HEAD^{tree}` equality and `git diff --exit-code HEAD FETCH_HEAD --` exit0. Full receipt `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/receipts/POSTFLIGHT.txt` SHA256 `1e09c80905ddf27c75a2be266436aa3550ba4d8267687cf0ccf6cafe05b8cd86`. Stage ref publication does **not** move dev-main or imply acceptance.

## 3. Exact PCM-only Go/C pair and source/fixture parity

- **Candidate test pair:** T `71788409df4d9eacef21816a8f82fc545a6f4ff1` / C `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`. C remains exact accepted Stage04/Stage05 C04 implementation; local C dev-main HEAD and detached verification worktree both `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`, tree `d79887b4cd9944759748dc83378204d34a7c8bf2`.
- The Go fixture manifest and C fixture manifest are byte-for-byte identical: `9d26d6cca1e29b10f1abaee583da12ee918ea4252f16f70c3eb2c01531e1d2f4` (SHA256 each). `cmp` and `sha256sum -c SHA256SUMS` in C `testdata/` PASS.
- Go production `internal/asp` compiled session tests executed against C04 `build/runtime_mock`, `build/io_faults.so`, `build/codec_probe`; session and decoder exercised with real compiled C, **test-only ABI vendor mock** (not proprietary on-device DSP).
- C `make test`, `make asan`, `make ubsan` exit0; immutable source ref/tree and independent fixture parity recorded. C codec/H/W and launcher policy regressions therefore run on unchanged C04. Independent C receipt: `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/receipts/C04_REGRESSION_MANIFEST.md`, SHA256 `067d1ca327cb1d4b50f784ea425b2eba8f34fbdd6f5c96ad72b287f7a3617a71`.
- C04/mock hashes and all new Go Linux build hashes are in `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/receipts/BINARIES-SHA256SUMS`; full entries follow.

- `echod-linux-arm32-dot_donut`: `289861736d959c810f88d09f1fb2cf121af8861cf262448233f42845df3e4fea`
- `echod-linux-arm32-panel`: `27f8840b366585bc8425b95446a3ed9bf38baf1c4423c6d56ab0c33c2fc863fb`
- `echod-linux-amd64-spot`: `ea551b9c8b8b4cc606ab403d2e049822080acf4150d5b1824d30165ef7baeaac`
- `echod-linux-amd64-panel`: `959838080808b5d1043bbf9983dd0e219f3273093dceac28280dca888abe0e87`
- `echod-linux-amd64-dot_donut`: `7e134fb2a2009d93574613f6765b308f13480210f0ae003bafeee932f11701c7`
- `echod-linux-arm32-dot`: `78af7f8a68abb85517fe8218af9fb6afd3a60e3d157fdc62688427c55238367f`
- `echod-linux-arm32-default`: `3e0b733bbfa014debacb607a8a5ae0cf72fd6c8c8b1eb74e97e7f263504a7292`
- `echod-linux-amd64-dot`: `a020917e91e2750074332928b9b42df98129406e1d8a5dd7a9b7b4c820162f39`
- `echod-linux-arm32-spot`: `d962e597eecffa3509b998ef23e803f3a68e797609f5570996332f24ecfb16a0`
- `echod-linux-amd64-default`: `9e90602b0e2dd62cf8f8c37702e51efb39446c1fe43eba408791c8002b1ebc55`
- `runtime_mock`: `5c19a4ee07d25c6821aea43964a3f46ac966310e6ba5e461f97a3e344fdd244d`
- `io_faults.so`: `4724d98c4cf89fa15454a39ff0161c81b0f55b310a98d49a4bda17363bc9d4d3`
- `codec_probe`: `423084acb8c414d8178b6ebfab39d3fcc40f61eab549f4fafebb585aa249cf3b`
- `test_alias_policy`: `775481c1fb0e4752cf2a4b9f0e0409e9fd819fdb61a70ed20ff61edf6fae4a39`

## 4. Change inventory / privacy guarantee

**Changed exactly one file:** `echod/internal/hardware/mic/asp_control_test.go` in the Go stage worktree. `TestL07NativeSoftwareCutAndStockRaw` uses the full physical ALSA channel count:
`cfg := alsaConfig(); cfg.Channels = Channels; fake := &fakePlainCapture{cfg: cfg, bytes: raw}`.

This matches `stage06Stock()`, whose input is `FrameSamples*Channels*3`. The test additionally verifies native descriptor physical roles/indices/byte offsets for every MIC and REF, notably **both REF L/R** where the stock platform carries two reference channels, all native sample bytes are muted, and stock raw is muted. MIC-only `alsaConfig()` is retained for tests that actually model MIC-only geometry.

- Production: **+0/-0/net0**; Go tests **+18/-2/net+16**; C production/tests **+0/-0**, docs **+0/-0**. No file deletions.
- `plainNative`, native validation and fail-closed behavior remain byte-identical; no IPC/descriptor/calibrator/lifecycle/stock-driver/mount/IDME changes. Preexisting `descriptor_test.go` formatting finding outside scope remains unchanged.
- No legacy CFG2/TLV/status/feedback/preflight/async/retry/manager additions; only test fixture/observability changed.

## 5. Reproduction and complete tests/builds

**Pre-change FAIL required by coordinator:** CWD `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod`, command `go test -count=1 -parallel 1 -tags dot -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic` on original Go accepted base HEAD, **exit1**: `native geometry unavailable: ... bytes=8640 stride=21`. Actual baseline log SHA256 `2c3d0d003e9e4649fdcf26940869f2c62079e258abc71e459d7414cf65cd2b96`.

**Post-change:** 5/5 focused for default/dot/dot,donut/spot/panel; 5/5 **full** (not compile-only) `go test -count=1 -parallel 1 [tags] ./...`, Panel debug loopback, 5 tagged L01–L09 race groups, full relevant race for default and dot including Source, HA, boot, service, ASP and phone, five `go vet ./...` variants, 5 amd64 and 5 Linux ARM32 builds; compiled joint Go/C session, C04 test/ASan/UBSan, fixture checks, diff/gofmt checks. All return exit0.

**Environment for all 40 post-change commands:** Go `PATH=/home/gptlab/dev/techo5-asp-p3-4a-arm32-build-20261027/toolchains/go1.26.9/bin:$PATH`; `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOFLAGS=-p=2 GOMAXPROCS=2`; `ASP_STAGE05_C_MOCK=/home/gptlab/dev/frankenasp-v3-stage07/build/runtime_mock`; `ASP_STAGE05_C_FAULTS=/home/gptlab/dev/frankenasp-v3-stage07/build/io_faults.so`; `ASP_C_DECODER=/home/gptlab/dev/frankenasp-v3-stage07/build/codec_probe`. Go compiler `go1.26.9 linux/amd64`; Linux ARM32 `GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0`; C tests use C04 Makefile/native GCC. Runner: `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/run-s07-correction.sh`, SHA256 `48139bfd999aff0f9d7ba135bc9dfdccd836262f20e96fd5f619c848657e427a`.

**Exact command ledger** `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/results.tsv` SHA256 `fccf087087b2602ba287f4b2bd82e25cf3a0366515e00ac8dce52602a7b5a18d`, all 40 per-log SHA256 checked independently against the ledger:

| ID | CWD | Exact command | Exit | Log SHA256 |
|---|---|---|---|---|
| c-test | `/home/gptlab/dev/frankenasp-v3-stage07` | `make test` | 0 PASS | `ff887bed27fa08d8896b4fd9c60667a1ce13c9c36fe6b01fc2bb41c411b75c82` |
| c-asan | `/home/gptlab/dev/frankenasp-v3-stage07` | `make asan` | 0 PASS | `c49f039ce755eec16d2e6693353252792dff5ab5205c94c36358be8bd21abe24` |
| c-ubsan | `/home/gptlab/dev/frankenasp-v3-stage07` | `make ubsan` | 0 PASS | `e14f22abe7f6e1d5a80310d6b78c8cc8208f8b00e71f147d4180194d8013f68d` |
| fixtures | `/home/gptlab/dev/frankenasp-v3-stage07/testdata` | `cmp SHA256SUMS '/home/gptlab/dev/techo5-v3-stage06-s07fix/echod/internal/asp/testdata/SHA256SUMS' && sha256sum -c SHA256SUMS` | 0 PASS | `b2c29c181614052cfa88bd2d79444a70c0f6257fc19c8e4eccac4f9a13f76737` |
| compiled-go-c-session | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -v ./internal/asp` | 0 PASS | `615846161ba54b82ed54aed289aa1ba2a13f104e703ef877348236745a272b8d` |
| focused-default | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1  -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic` | 0 PASS | `d82ff019a6fcff208d761c463566a12d95bf2f7fcd4fae4dfa39b49eb65f82e7` |
| focused-dot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags dot -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic` | 0 PASS | `49c32e4701afd10f5a68d21f12c18361d8f4e5d0fbf396b8e6a685e6d37155fc` |
| focused-dot_donut | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags dot,donut -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic` | 0 PASS | `49c32e4701afd10f5a68d21f12c18361d8f4e5d0fbf396b8e6a685e6d37155fc` |
| focused-spot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags spot -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic` | 0 PASS | `d82ff019a6fcff208d761c463566a12d95bf2f7fcd4fae4dfa39b49eb65f82e7` |
| focused-panel | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags panel -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic` | 0 PASS | `d82ff019a6fcff208d761c463566a12d95bf2f7fcd4fae4dfa39b49eb65f82e7` |
| stock-full-default | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1  ./...` | 0 PASS | `fe953f6dd3a4c29c8065fb2117b941afa5d1d9d6acaa3e2f1ffcb77062fa8306` |
| stock-full-dot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags dot ./...` | 0 PASS | `c9438083039db54c63139b74867090adb6df9911a68772ae69c7e8a118f35ed5` |
| stock-full-dot_donut | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags dot,donut ./...` | 0 PASS | `100192192a701b32fd759f9ae8a39e5726c750c0839d3ea81d2ea2c3f4321b4f` |
| stock-full-spot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags spot ./...` | 0 PASS | `018677217ed6412cb132827cf62caf33a7c19af5767d20745a6d5f5b5cbbed12` |
| stock-full-panel | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags panel ./...` | 0 PASS | `13241f03efa227bf383f64151e5cedb583cf10bcd9bd6df3f9b909cc55591047` |
| panel-debug-loopback | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -count=1 -parallel 1 -tags panel,panel_debug ./internal/lib/loopback` | 0 PASS | `72abd94ba04ecc7fc0b7ef8a6453f9a85e685aafd1c283d6ce9075fee7bd23be` |
| race-L-default | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -race -count=1 -parallel 1  -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 PASS | `7d9dae7660ed111ac82ef07b7e9cc8bbd456855c05b8f2da648d837ac4366664` |
| race-L-dot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -race -count=1 -parallel 1 -tags dot -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 PASS | `3a09f93ee6a14b5e289a93344e381d7e0c1e26ab46d0b4d5ed182896670de66e` |
| race-L-dot_donut | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -race -count=1 -parallel 1 -tags dot,donut -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 PASS | `ed785cd048c141ebf4ddf0b05879644f319b18910fc71167d484fcf3de89ba70` |
| race-L-spot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -race -count=1 -parallel 1 -tags spot -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 PASS | `de54fe3e219e0bde8359e7c7e32ba9c6a085245df8ee43dacba07acba265ac30` |
| race-L-panel | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -race -count=1 -parallel 1 -tags panel -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | 0 PASS | `75e3e345c93b228c12cb91f2cab31e9e850b27fea21b2f6aa1bbc27a368f1a01` |
| race-relevant-default | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -race -count=1 -parallel 1  ./internal/hardware/mic ./internal/feature/microphone ./internal/boot ./internal/service ./internal/asp ./internal/config ./internal/component/all ./internal/feature/phone ./internal/hardware/privacy ./internal/layout` | 0 PASS | `d5c5815e76f15045ad5caf04aeb16b36087fa071b6e4a97a1eade00b5e79a307` |
| race-relevant-dot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go test -race -count=1 -parallel 1 -tags dot ./internal/hardware/mic ./internal/feature/microphone ./internal/boot ./internal/service ./internal/asp ./internal/config ./internal/component/all ./internal/feature/phone ./internal/hardware/privacy ./internal/layout` | 0 PASS | `943a41bfdd12a1991ea848117c727e1998f476997a8fe4343da3de50b411bf13` |
| vet-default | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go vet  ./...` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-amd64-default | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-amd64-default' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-arm32-default | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build  -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-arm32-default' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| vet-dot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go vet -tags dot ./...` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-amd64-dot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags dot -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-amd64-dot' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-arm32-dot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags dot -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-arm32-dot' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| vet-dot_donut | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go vet -tags dot,donut ./...` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-amd64-dot_donut | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags dot,donut -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-amd64-dot_donut' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-arm32-dot_donut | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags dot,donut -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-arm32-dot_donut' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| vet-spot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go vet -tags spot ./...` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-amd64-spot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags spot -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-amd64-spot' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-arm32-spot | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags spot -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-arm32-spot' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| vet-panel | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `go vet -tags panel ./...` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-amd64-panel | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags panel -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-amd64-panel' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| build-linux-arm32-panel | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -tags panel -o '/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/builds/echod-linux-arm32-panel' ./cmd/echod` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| diffcheck | `/home/gptlab/dev/techo5-v3-stage06-s07fix` | `git diff --check` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| gofmt-correction | `/home/gptlab/dev/techo5-v3-stage06-s07fix/echod` | `test -z "$(gofmt -l internal/hardware/mic/asp_control_test.go)"` | 0 PASS | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |

**Manifests:** `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/receipts/LOGS-SHA256SUMS` SHA256 `b07d94244d08d10dfbb99e8ba898611c412a2f891eb2dd453144342c8fedaa66`; `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/receipts/BINARIES-SHA256SUMS` SHA256 `925f17c09eead38af356507faad88112a60f4221e2ea24a782b52638eab91f07`; `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/receipts/SOURCE-SHA256SUMS` SHA256 `2dc96df448af23e92a1fcfc354de42125f5d475cc67d25e30c5605b4cab9b13a`. Build outputs are local offline binaries, **not** deployed to Echo.

## 6. Deviations / STOP / remaining risks

- **FAIL-07-01 addressed in test-only correction**, no production bug claimed or hidden. Original Stage07 FAIL/STOP handoff remains historically valid and unmodified; the Stage07 retake has **NOT** been executed.
- Test evidence uses simulated native capture and compiled C04 mock/codec, **not** vendor DSP on real Echo, factory filesystem mount, privileged launcher or acoustics. These were prohibited.
- Android Panel NDK/ABI build is outside the correction run and is **NOT RUN / not claimed PASS**, even if the independent toolchain preparation progresses. A future Stage07 retake must build its actual Android arm64-v8a and armeabi-v7a targets; Linux Panel tags are not substitutes.
- No `descriptor_test.go` hygiene edit; no main/docs/C changes; no auto-integration/rollback/force-push.

## 7. Coordinator handoff and next gate

- Independent coordinator must verify exact Go stage **remote commit/tree** and unchanged C input, this 40/40 ledger/compilation/reproduction and clean worktrees, then render **ACCEPTED or REJECTED** independently. Executor leaves **acceptance PENDING**.
- Only after ACCEPTED and under separate authorization may coordinator perform **Go fast-forward only**, with C unchanged, verify remote integration receipts and publish `stages/06_S07_CORRECTION_HANDOFF.md` on docs staging. **Integration remains PENDING** here.
- Only after that separate gate and verified Android toolchains can a new full **Stage07 S01–S10 retake** be launched under its own explicit assignment. No automatic launch here.

## 8. Documentation publication

This executor-only evidence/report lives outside all code and docs checkouts: `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/06_S07_CORRECTION_HANDOFF_EXECUTOR.md`; docs coordinator publishes approved decision separately. No commits/changes to `experimental/v3-design-staging`, either dev-main, either protected main or C stage refs.

## 9. Independent coordinator technical acceptance / Go integration receipt (2026-10-31)

**Original executor evidence preserved unchanged** at `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/06_S07_CORRECTION_HANDOFF_EXECUTOR.md`, SHA256 `5be865b56f1be20dc06dfb4b9745f5fdbca7928ccf27440d881de6ca94695d40`. Original report's PENDING states and pre-integration remote refs are historical executor snapshots, not the final coordinator decision.

**Independent technical decision: ACCEPTED** for the tightly scoped Go-only Stage06-S07 **test-fixture correction gate**, not Stage07 final acceptance. Reviewed actual published Go commit `71788409df4d9eacef21816a8f82fc545a6f4ff1` (tree `0b55a3d840fd2f21238ee11f82313be1c3489965`), with one parent exact accepted Go Stage06 `cce2f220aeab9835e10a7f7ab659f5ca390011e8`. Diff contains exactly one changed file `echod/internal/hardware/mic/asp_control_test.go`, 18 additions and two deletions. `TestL07NativeSoftwareCutAndStockRaw` now constructs a fake capture `cfg := alsaConfig(); cfg.Channels = Channels`, consistent with `stage06Stock()` full MIC+REF input. Assertions independently verify MIC and REF roles, channel indices and frame offsets, silence on all physical channels (including REF L/R where present), and stock raw after SoftwareCut. **Production Go and C files are byte-for-byte unchanged**; no relaxation of native PCM validation, protocol, calibration or Source lifecycle.

**Independent evidence verification:** coordinator read and SHA256-verified original 40/40 exit-0 ledger `results.tsv`, all 40 referenced log contents and their digests, published Go branch/tree, C04 fixture equality, `receipts/LOGS-SHA256SUMS`, `receipts/BINARIES-SHA256SUMS`, and `receipts/SOURCE-SHA256SUMS`. Executor evidence includes pre-fix S07 dot FAIL `bytes=8640 stride=21`, five corrected focused tests, full five-tag Go test matrices, five `TestL` race runs, relevant default/dot race, five vets, C `make test`/ASan/UBSan, compiled Go↔C session, five AMD64 and five Linux ARM32 builds, clean/gofmt/source/fixture checks. All passed per independently hashed logs. The original source report accurately keeps Android ABI Stage07 retake outside this Go-only correction.

**Independent coordinator reruns** on the exact committed Go SHA and C04: focused `TestL07NativeSoftwareCutAndStockRaw` for default, dot, dot,donut, spot, panel **PASS (5/5)**; complete `go test -count=1 -parallel 1 -tags dot ./...` **PASS**; `go test -race -count=1 -parallel 1 -tags dot -run '^TestL'` across native MIC, HA and service **PASS**; compiled independent C04 Go session `go test -count=1 -parallel 1 ./internal/asp` **PASS**; relevant `go vet -tags dot` and `gofmt -l` **PASS**. Independent full-dot log: `/home/gptlab/dev/techo5-v3-stage06-s07fix-evidence-20261031/coordinator-full-dot.log`. Other full Go tags and C sanitizers are PASS by hash-verified executor evidence, not falsely claimed as all independently rerun.

**Exact integration:** before FF the remote Go dev-main was `cce2f220aeab9835e10a7f7ab659f5ca390011e8`; published correction stage ref was `71788409df4d9eacef21816a8f82fc545a6f4ff1`; C dev-main `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`, docs predecessor `00886925a2620bc8850e2596707eaab01529be6a`. All were checked against GitHub. Coordinator checked one-commit ancestry and clean worktrees, then ran `git merge --ff-only 71788409df4d9eacef21816a8f82fc545a6f4ff1` in `/home/gptlab/dev/techo5-v3-dev-main` and direct `git push --porcelain origin HEAD:refs/heads/experimental/external-asp-v3-dev-main` from anfisa-trainer. Verified both published stage ref and Go dev-main equal the exact accepted SHA; C dev-main remains unchanged. Protected mains still Go `62603a7854283a3a07473a17c831c6ef0fd6f8d4` and C `e986ae18643e08e95f29a38aae1dd74762f6362e`. A post-push shell typo at the very end of a postcheck command did not affect the already completed push; the coordinator repeated all final clean-worktree/ref checks separately and obtained PASS. No force/reset/rebase/squash/device operation.

**Final pair for Stage07 Retake:** TECHO5 Go **`71788409df4d9eacef21816a8f82fc545a6f4ff1`** + FrankenASP V3 C **`f6266907edb90e7c9c2f414e8e46e6f215a21b7b`**. Stage06-S07 execution **PASS**, independent delegated technical acceptance **ACCEPTED**, Go-only integration **COMPLETE**. Android NDK r28c is separately installed and verified on trainer (see `../ANDROID_NDK_TOOLCHAIN.md`), but **Stage07 remains FAIL/STOP historically** and must be fully retaken under a separately assigned S01–S10 task. This correction alone does not prove Stage07 S08, Echo, vendor Calibrator, device or acoustics. The existing `stages/07_HANDOFF.md` stays immutable.
