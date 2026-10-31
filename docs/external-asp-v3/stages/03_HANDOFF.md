# TECHO5 / FrankenASP V3 — Stage03 Native Capture, Raw Metadata & SoftwareCut — HANDOFF

## 1. Identity, authority, status

- Stage: **03 — Native Capture, Raw Metadata & SoftwareCut**, normative `stages/03_NATIVE_PRIVACY_TASK.md`; Europe/Riga, 2026-10-31.
- Executor: ChatGPT Stage03 code executor, directly on **anfisa-trainer** via Desktop Commander.
- Docs authority: `dmit2k/techo5:experimental/v3-design-staging` at `66b91721315fdf0a7d9d9a3c13a9f9bad498f111`.
- Approved SPEC blob: `b324d2161e5ea316618f23bbe6d6aedf6c2df5c8`, verified by `git rev-parse HEAD:docs/external-asp-v3/SPEC_V2_APPROVED.md`.
- Prior accepted handoff: `docs/external-asp-v3/stages/02_HANDOFF.md`: execution PASS, technical acceptance ACCEPTED, integration COMPLETE.
- Execution status: **PASS — 18/18 matrix commands exit 0; post-commit focused, race, Go↔C tests PASS**.
- Acceptance: **ACCEPTED (delegated independent technical review, 2026-10-31)**. This review explicitly included the N07 defect correction and post-commit evidence; not a claim that the owner personally uttered acceptance.
- Integration: **COMPLETE (2026-10-31)**. TECHO5 dev-main was fast-forwarded to corrected T03; FrankenASP V3 dev-main remained at C02.
- Software gate: **native component / stock compatibility**, NOT runtime, vendor DSP or paired wire.
- Hardware, vendor-device runtime, acoustic tests, actual ALSA capture, real IDME/eMMC reads, and Echo deployment: **НЕ ВЫПОЛНЕНЫ** (prohibited in Stage03).

## 2. Exact immutable input and candidate

| Record | TECHO5 | FrankenASP V3 |
|---|---|---|
| Repository | `dmit2k/techo5` | `dmit2k/frankenasp-v3` |
| Main frozen SHA | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` | `e986ae18643e08e95f29a38aae1dd74762f6362e` |
| Accepted input integration SHA | `78a01f897b89dd462576e6e617fbc8216c4518dd` | `0f39216a5240306344af7c2457716f4eaa466f40` |
| Integration ref | `experimental/external-asp-v3-dev-main` | `experimental/v3-dev-main` |
| Stage03 branch | `experimental/external-asp-v3-dev-stage03` | **NONE**, unchanged |
| Worktree | `/home/gptlab/dev/techo5-v3-stage03` | no Stage03 worktree |
| Local candidate commit | `4fb329cf789b156bbeb8d13da20b8c4bda70f108` | **unchanged** `0f39216a5240306344af7c2457716f4eaa466f40` |
| Candidate ancestry | T02 `78a01f897b89dd462576e6e617fbc8216c4518dd` → original Stage03 `3aebef58c7abce16764dbbb900c637efcd76e3c8` → accepted fix `4fb329cf789b156bbeb8d13da20b8c4bda70f108` | unchanged |
| Published stage remote / clean worktree | **PASS: published SHA 4fb329cf789b156bbeb8d13da20b8c4bda70f108; clean** | **unchanged** |
| Output integration SHA | **`4fb329cf789b156bbeb8d13da20b8c4bda70f108`** verified | **`0f39216a5240306344af7c2457716f4eaa466f40`** unchanged, verified |

Tests and SHA-bound evidence refer to **T=4fb329cf789b156bbeb8d13da20b8c4bda70f108 / C=0f39216a5240306344af7c2457716f4eaa466f40**. C code was NOT changed and no C Stage03 branch or empty commit exists.

## 3. Changes, contracts, invariants

### TECHO5 files

- `echod/internal/hardware/mic/native.go` **new** — minimal optional `captureReadWithNative` and `NativeBlock`, direct `[][]byte` borrowed input, Stage02 `asp.Descriptor` physical buffer and captured-channel records, sample-wise in-place `zeroNative`, no host DSP adaptation; `readCaptured` shares one SoftwareCut result; `DescriptorWithRaw` attaches unmodified source values to existing codec.
- `echod/internal/hardware/mic/athena.go` **edited** — paired MIC S32 and stereo REF S16 input observed before conversion. `ReadWithNative` returns MIC and two REF physical buffers when hardware REF is actually opened. MIC signed24-in-S32 valid bits=24 with low shift=0 or mainline high shift=8; REF S16 valid bits=16 shift=0. Real MIC/REF capture sample rates and frame count from existing open settings and buffer sizes; no Athena-only host qualification or MIC4/REF2/96k restrictions in codec. No hardware REF -> only MIC buffer, stock software REF remains stock-only. Existing `Read`, `convert`, `softFill`, XRUN `restart` and meter stages remain.
- `echod/internal/hardware/mic/mic.go` **edited** — one `privacy.SoftwareCut()` callback evaluated **after** physical capture completes but before native stock conversion, shared by native in-place cut and stock `broadcastWithCut`; preserves unmodified single built-in DSP, raw copy and owned mono publication.
- `echod/internal/lib/alsa/alsa.go` **edited** — `Capture.Config()` exposes existing hw_params-confirmed capture configuration without additional ioctl or capture.
- `echod/internal/layout/raw_sources.go` **new** — `RawIdentitySources()` and `RawFactorySources()`: byte-exact values from existing stock origin paths `/proc/device-tree/compatible`, `/proc/cmdline`, `/proc/idme/<name>`, `/proc/device-tree/idme/<name>/value`; Athena mainline eMMC fallback via existing boot area only when proc fields absent. The IDME v2 parser reads raw field spans and leaves values byte-exact. Unknown keys pass through; absent sources are omitted, empty present value stays distinct. No `layout.Idme()` normalization, `parseIdme()` strings, Q30, masks, confidence, availability, MIC-slot mapping, synthetic defaults or platform admission.
- `echod/internal/hardware/mic/native_test.go`, `athena_native_test.go`, `echod/internal/layout/raw_sources_test.go` **new** — synthetic fixtures and fake capture backends, NO hardware acquisition.

Final accepted diff: **9 files; +871 / -17 lines** (production and tests combined, no prior wire change, no C changes).

### API/lifetimes for Stage05/06

- `NativeBlock.Buffers` shares underlying original capture bytes; `NativeBlock.Geometry` describes them in transfer order as Stage02 v1 `asp.Descriptor`. The callback is synchronous, called before `pairedCapture.convert()` with the true unconverted MIC and both REF L/R. **Borrowed lifetime ends on next capture Read**; future synchronous IPC must consume them before that. They are NOT published or retained by `Source`.
- Plain ALSA `Read` already returns native interleaved PCM directly; `readCaptured` uses the caller's existing `raw` buffer without a second buffer copy. Panel remains stock socket capture, not mislabeled hardware REF.
- One frame-level SoftwareCut policy read **after** hardware input and before conversion; when true, native MIC/REF zero in place (format width/stride and sample offsets, protecting padding), and the same cut zeroes stock `raw`. Hardware mute REF semantics unchanged: hardware mute without SoftwareCut may still have REF.
- Published mono frames retain stock `append([]int16(nil), frame...)` ownership; `ListenRaw` retains copy; history, sensors, cancellation, subscriptions, MWW/STT/Panel/SIP callers remain on single Built-in path. No External output selected at Stage03.
- Producer of actual raw metadata may call `layout.RawFactorySources`/`RawIdentitySources` at future session initialization (NOT per capture read), and attach to the descriptor via `DescriptorWithRaw` + unchanged Stage02 `asp.Encode`. Actual device provenance and bytes are not hardware-verified.

### Exceptions and intentionally unimplemented areas

- `pairedCapture.restart` retains **stock XRUN decimator reset**, required by baseline; no reset on mute.
- Existing `putRef` stock first-channel stereo mix is retained for Built-in; native REF callback receives **unmixed L/R**.
- Stage03 does NOT implement helper, DSP runtime, IPC pipes/exchange, launch, selected/active, FIFO, session ownership, secure launch, HA mode, or External retention. No legacy CFG2/TLV, H/E, status, preflight, whitelist, synthetic production adapter, pools, queues, extra managers, new DSP transforms, or process logic added.
- Real Echo capture geometry/rates, calibration availability, and firmware-specific metadata remain **unverified on hardware** by design. Stage02 descriptor rejects oversize/invalid raw entries rather than changing wire or guessing values.

## 4. N01–N08 acceptance matrix

All Stage03 native tests use **synthetic data only**; the original executor matrix ran 18/18 PASS. Independent review additionally reproduced and corrected the partial /proc+eMMC regression and reran final-SHA tests.

| ID | Scenario and evidence | Result |
|---|---|---|
| N01 | `TestN01N02PairedOriginalStereoBeforeConversion`: callback receives original paired MIC and REF before stock convert, normal stock output retained | **PASS — synthetic/fake-backend only** |
| N02 | `TestN02PlainBorrowedAndNoREF`, `TestN02PairedS32High24Geometry`, N01 stereo: direct plain interleaved, true no-REF, two independent nonzero hardware REF channels, correct physical records / alignment | **PASS — synthetic/fake-backend only** |
| N03 | `TestN03SparseNativeZeroBounds`, `TestN03PairedCutAlsoMutesStockSoftwareReference`, `TestN03PolicyExactlyOnceAfterPlainRead`, `TestN03PairedPolicyBoundaryTiming`: once after read, shared native + stock mute, sparse stride/canary and software REF | **PASS — synthetic/fake-backend only** |
| N04 | `TestN04N06BorrowedUntilNextReadAndHardwareREFUnchanged`: original backing storage reused on next read, no retained input | **PASS — synthetic/fake-backend only** |
| N05 | `TestN05OwnedStockRetentionHistoryAndRaw`, plus prior `broadcast_test.go` and stock mic/history tests: retained mono and raw frames, history and subscribers, stock stages | **PASS — synthetic/fake-backend only** |
| N06 | N03/N04 and existing stock privacy tests: hardware REF unaffected without software cut, mute/unmute no extra resets, XRUN stock lifecycle remains | **PASS — synthetic/fake-backend only** |
| N07 | `TestN07RawProcAndIdentityBytes`, `TestN07BootAreaRawIDME`, `TestN07NativeDescriptorRawByteExact`: source names, embedded NUL/high bytes, zero-length, absent/unknown, byte-exact codec; no normalization | **PASS — synthetic/fake-backend only** |
| N08 | all relevant mic/privacy/layout/asp tests, `-race`, `go vet`, whole default app regression, tag matrix default/dot/dot+donut/spot/panel and ARM32 cross-build | **PASS — synthetic/fake-backend only** |

### Executed commands/logs

All tests run under `/home/gptlab/dev/techo5-v3-stage03/echod`, env `PATH=/home/gptlab/.local/toolchains/go/bin:$PATH; GOTOOLCHAIN=local; GOPROXY=off; GOSUMDB=off; GOFLAGS=-p=1; GOMAXPROCS=2`; native amd64 Linux. Test runner: `/home/gptlab/dev/techo5-v3-stage03-matrix-20261031.sh`; evidence dir: `/home/gptlab/dev/techo5-v3-stage03-evidence-20261031/`, `results.txt`, each `*.log`, `SHA256SUMS`, `BINARIES_SHA256SUMS`, `env.txt`.

- `go test -count=1 -parallel 1 ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy`
- `go test -race -count=1 -parallel 1 ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy`
- `go vet ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/lib/alsa`
- `go test -count=1 -parallel 1 ./...`
- `go build -o <evidence>/echod-linux-amd64 ./cmd/echod`
- For each tags `dot`, `dot,donut`, `spot`, `panel`: `go test -count=1 -parallel 1 -tags <tags> ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy`; `go test -run '^$' -parallel 1 -tags <tags> ./...`; `go build -tags <tags> -o <evidence>/echod-linux-amd64-<tag-label> ./cmd/echod`.
- `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o <evidence>/echod-linux-arm32 ./cmd/echod`.
- All actual command exits (incl baseline failures, if any): **all 18 commands exit 0; full ledger below**. Target hardware run is not part of these commands.

## 5. Git publication and protected refs

- Direct SSH push from `anfisa-trainer` **only**: `git push --porcelain origin HEAD:refs/heads/experimental/external-asp-v3-dev-stage03` with original Git object and configured `core.sshCommand`, existing repo-scoped deploy key, no intermediate hosts, no force.
- Remote Stage03 SHA: **4fb329cf789b156bbeb8d13da20b8c4bda70f108 (verified after direct SSH push)**.
- Protected Go main, C main, both dev-main and docs branch: **VERIFIED UNCHANGED AFTER PUSH** (exact refs below).
- All three checked worktrees after push: **CLEAN** (TECHO5 Stage03, FrankenASP dev-main, docs staging).
- Code executor **does not edit docs branch**; this handoff remains outside code/docs worktrees for subsequent independent review and coordinator publication to `docs/external-asp-v3/stages/03_HANDOFF.md`.

## 6. Forward handoff / acceptance

- **Next stage**: 04 — C helper runtime from the accepted integrated pair T03/C02, only via its own execution assignment.
- Input for Stage04 is T03=`4fb329cf789b156bbeb8d13da20b8c4bda70f108`, C02=`0f39216a5240306344af7c2457716f4eaa466f40`; both remote dev-main refs verified.
- Independent delegated technical acceptance and coordinator integration are complete; there is no owner Git task.
- No real Echo deployment or hardware tests authorized.

## 7. Final test ledger and evidence (2026-10-31)

- **Gate: Stage03 Execution PASS**, N01–N08 synthetic tests PASS; **18 / 18 matrix commands exited 0**. Technical acceptance **ACCEPTED**; integration **COMPLETE**. Hardware: NOT RUN/PROHIBITED.

| Check | Exit | Exact command (see corresponding `<id>.log`) |
|---|---:|---|
| `native` | `0` | `go test -count=1 -parallel 1 ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy ` |
| `race` | `0` | `go test -race -count=1 -parallel 1 ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy ` |
| `vet` | `0` | `go vet ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/lib/alsa ` |
| `full-default` | `0` | `go test -count=1 -parallel 1 ./... ` |
| `build-default` | `0` | `go build -o /home/gptlab/dev/techo5-v3-stage03-evidence-20261031/echod-linux-amd64 ./cmd/echod ` |
| `test-dot` | `0` | `go test -count=1 -parallel 1 -tags dot ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy ` |
| `compile-dot` | `0` | `go test -run \^\$ -parallel 1 -tags dot ./... ` |
| `build-dot` | `0` | `go build -tags dot -o /home/gptlab/dev/techo5-v3-stage03-evidence-20261031/echod-linux-amd64-dot ./cmd/echod ` |
| `test-dot-donut` | `0` | `go test -count=1 -parallel 1 -tags dot\,donut ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy ` |
| `compile-dot-donut` | `0` | `go test -run \^\$ -parallel 1 -tags dot\,donut ./... ` |
| `build-dot-donut` | `0` | `go build -tags dot\,donut -o /home/gptlab/dev/techo5-v3-stage03-evidence-20261031/echod-linux-amd64-dot-donut ./cmd/echod ` |
| `test-spot` | `0` | `go test -count=1 -parallel 1 -tags spot ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy ` |
| `compile-spot` | `0` | `go test -run \^\$ -parallel 1 -tags spot ./... ` |
| `build-spot` | `0` | `go build -tags spot -o /home/gptlab/dev/techo5-v3-stage03-evidence-20261031/echod-linux-amd64-spot ./cmd/echod ` |
| `test-panel` | `0` | `go test -count=1 -parallel 1 -tags panel ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy ` |
| `compile-panel` | `0` | `go test -run \^\$ -parallel 1 -tags panel ./... ` |
| `build-panel` | `0` | `go build -tags panel -o /home/gptlab/dev/techo5-v3-stage03-evidence-20261031/echod-linux-amd64-panel ./cmd/echod ` |
| `build-arm32` | `0` | `env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o /home/gptlab/dev/techo5-v3-stage03-evidence-20261031/echod-linux-arm32 ./cmd/echod ` |

### Tested commit confirmation after publish

- `git rev-parse HEAD` = `4fb329cf789b156bbeb8d13da20b8c4bda70f108`; parent `78a01f897b89dd462576e6e617fbc8216c4518dd`.
- On published commit: `go test -count=1 -parallel 1 ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy`: **PASS**, `postcommit-focused.log`.
- On published commit: `go test -race -count=1 -parallel 1 ./internal/hardware/mic ./internal/asp ./internal/layout ./internal/hardware/privacy`: **PASS**, `postcommit-race.log`.
- On published commit with `ASP_C_DECODER=/home/gptlab/dev/frankenasp-v3-stage02/build/codec_probe`: `go test -count=1 ./internal/asp`: **PASS**, `postcommit-go-c.log`. No C repo or C sources changed.
- Detailed time-ordered matrix log checksums: `SHA256SUMS`; output binary checksums: `BINARIES_SHA256SUMS`.
- SHA256 `postcommit-focused.log`: `432bb5d94e728ea29102b673679026c8326534e5b11044e696b2455397d71099`.
- SHA256 `postcommit-race.log`: `f2c29a99666722f518b52bfe6dab2c557b0e3dbff51b27bee67bed15db5ff0dc`.
- SHA256 `postcommit-go-c.log`: `b1189b29c98a91db5149fda09523a468bc10fd4977c57aabc2bed1931b4fea2a`.

### Remote refs checked after push (not local guesses)

| Remote ref | Verified SHA |
|---|---|
| `dmit2k/techo5:refs/heads/main` | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` |
| `dmit2k/techo5:refs/heads/experimental/external-asp-v3-dev-main` | `78a01f897b89dd462576e6e617fbc8216c4518dd` |
| `dmit2k/techo5:refs/heads/experimental/external-asp-v3-dev-stage03` | `4fb329cf789b156bbeb8d13da20b8c4bda70f108` |
| `dmit2k/techo5:refs/heads/experimental/v3-design-staging` | `66b91721315fdf0a7d9d9a3c13a9f9bad498f111` |
| `dmit2k/frankenasp-v3:refs/heads/main` | `e986ae18643e08e95f29a38aae1dd74762f6362e` |
| `dmit2k/frankenasp-v3:refs/heads/experimental/v3-dev-main` | `0f39216a5240306344af7c2457716f4eaa466f40` |

- Remote C Stage03 refs were absent (no empty Stage03 branch). Direct origin push issued on trainer using SSH deploy key; GitHub API not used for commit, no gptlab-core relay and no force push.
- Publication was limited to one TECHO5 Stage03 branch. Documentation staging was not modified. This report is prepared at `/home/gptlab/dev/techo5-v3-stage03-evidence-20261031/03_HANDOFF.md`; coordinator may publish it to `docs/external-asp-v3/stages/03_HANDOFF.md` only under separate authorized workflow.
- Stage03 was accepted and fast-forward integrated at the corrected T03 SHA. Stage04 and External activation were NOT started by Stage03.

## 8. Coordinator independent review and integration receipt

**Original executor report (historical evidence):** SHA256 `b607f4abbd6ff3c94cd7d8a3b9d53d9369b05a0113095b79ba37ce2c39a1e0e1`, 153 lines. It covered original commit `3aebef58c7abce16764dbbb900c637efcd76e3c8`. The initial 18/18 PASS matrix is kept as evidence for that original commit.

**Review finding and correction (N07):** on Athena-mainline, `rawFactorySources` initially read eMMC raw IDME only when /proc/DT exported no values at all. Stock `layout.Idme()` uses fallback per missing field. A new synthetic `TestN07PartialProcBootFallback` exposed a partial /proc serial and eMMC-only `miccal.0`; the test failed on the initial code because raw calibration was omitted. Correction `4fb329cf789b156bbeb8d13da20b8c4bda70f108` preserves any available proc/DT values, fills only missing raw IDs from the original boot-area bytes and retains source identity. No Go-side calibration interpretation, DSP compatibility decision, real hardware read during tests or Stage02 wire change.

**Exact independent post-commit checks at T03/C02:** `go test -count=1 -parallel 1` on mic/asp/layout/privacy: PASS; `go test -race -count=1 -parallel 1` on these packages: PASS; `go vet` on mic/asp/layout/alsa: PASS; Go↔compiled C conformance with `ASP_C_DECODER=/home/gptlab/dev/frankenasp-v3-stage02/build/codec_probe`: PASS. Logs `/home/gptlab/dev/techo5-v3-stage03-review-evidence-20261031/acceptance-postcommit-*.log`, checksums `ACCEPTANCE_SHA256SUMS`. The original executor's full default and full build/tag matrix passed 18/18. The additional independent matrix completed 9/18 commands, **all nine PASS**, including full `go test ./...`, default/dot builds, dot+donut focused tests, race and vet; the redundant remaining tag tests were intentionally not repeated. Its results are in the separate review evidence directory. No unexecuted check is claimed PASS in the independent rerun.

**Verified integration:** clean worktrees, original T02/C02 integration refs and corrected T03 stage ref were verified remotely before any update. Coordinator executed `git merge --ff-only 4fb329cf789b156bbeb8d13da20b8c4bda70f108` in `/home/gptlab/dev/techo5-v3-dev-main`, then direct trainer SSH `git push --porcelain origin HEAD:refs/heads/experimental/external-asp-v3-dev-main`, remotely verified target T03 and original Go main T01. FrankenASP dev-main stayed at C02, its main stayed at C01. Both stage refs/worktrees and docs worktree were checked clean, no force/rebase/squash/new C commit.

**Accepted Stage03 pair**: T03 `4fb329cf789b156bbeb8d13da20b8c4bda70f108` / C03 unchanged `0f39216a5240306344af7c2457716f4eaa466f40`. **Execution: PASS; delegated technical acceptance: ACCEPTED; integration: COMPLETE.** No Echo/device/real ALSA/acoustic/vendor runtime evidence is claimed.
