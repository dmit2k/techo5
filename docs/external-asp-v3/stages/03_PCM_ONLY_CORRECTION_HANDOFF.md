# TECHO5 / FrankenASP V3 — Stage03-R PCM-only native capture correction HANDOFF

## 1. Stage identity, gates and statuses

- Task: **Stage03-R — remove External metadata plumbing; preserve universal native capture**, 2026-10-31 (Europe/Riga), execution on `validation host` through authorized terminal tools and direct validation host SSH Git.
- Normative: `PCM_ONLY_ARCHITECTURE_DECISION.md`, `MASTER_CONTRACT.md`, `EXECUTION_BRANCH_STRATEGY.md`, `SPEC_V2_APPROVED.md` §§3–5,8–9 and `stages/03_NATIVE_PRIVACY_TASK.md`.
- Gate: **Stage02-R Execution PASS / independent technical acceptance ACCEPTED / Integration COMPLETE**, verified in current coordinator-published `stages/02_PCM_ONLY_CORRECTION_HANDOFF.md`.
- Docs staging HEAD: `4130606cf09cba0d8a5fbc1f288195c0582816d4`; current approved SPEC blob: `f3e1b00d1f5d184e345c0d2c79da8394435ba3a4`.
- **Execution: PASS**, 19/19 measured matrix commands exit=0 and committed SHA focused/race/vet checks; PCM-only native component gate only.
- **Independent technical acceptance: ACCEPTED (coordinator review, 2026-10-31); Integration: COMPLETE (verified fast-forward).** Original executor handoff was PENDING; the coordinator issued an independent decision after source/contract review and replay of focused/race/vet/full Go tests, then advanced only the Go dev-main by an ordinary FF. See final §7. Stage04 may now start only under its own Stage04 assignment.
- **Hardware/Echo, actual ALSA capture, real /proc/idme/eMMC/DT reads, vendor runtime, acoustics: НЕ ВЫПОЛНЕНЫ** (explicitly prohibited).

## 2. Exact revision pair / Git receipt

| Item | TECHO5 Go | FrankenASP V3 C |
|---|---|---|
| Repository | `dmit2k/techo5` | `dmit2k/frankenasp-v3` |
| Input accepted integration | `15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc` | `77149e6494ac9784d98464ff984c84887a28ced9` |
| Accepted output dev-main (coordinator verified) | **`b0c0d5fa96576eff8f7497d48e39fa30a5348e76` (FF)** | **`77149e6494ac9784d98464ff984c84887a28ced9` unchanged** |
| Stage branch | `experimental/external-asp-v3-dev-stage03-pcmonly` | **none, unchanged** |
| Worktree | `${WORKSPACE}/techo5-v3-stage03-pcmonly` | none |
| Tested / published candidate | **`b0c0d5fa96576eff8f7497d48e39fa30a5348e76`** | **`77149e6494ac9784d98464ff984c84887a28ced9` unchanged** |
| Frozen main | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` | `e986ae18643e08e95f29a38aae1dd74762f6362e` |
| Accepted SHA (delegated technical review, not a verbatim owner declaration) | **`b0c0d5fa96576eff8f7497d48e39fa30a5348e76`** | **`77149e6494ac9784d98464ff984c84887a28ced9`** |
| Coordinator output integration | **`b0c0d5fa96576eff8f7497d48e39fa30a5348e76`** | **`77149e6494ac9784d98464ff984c84887a28ced9`** |

The accepted input pair is **T02-R `15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc` / C02-R `77149e6494ac9784d98464ff984c84887a28ced9`**, not historical T03/C02. The requested output candidate pair is **T03-R `b0c0d5fa96576eff8f7497d48e39fa30a5348e76` / C unchanged `77149e6494ac9784d98464ff984c84887a28ced9`**. No changes to C, no C Stage03 branch, no changes to either main, dev-main, or docs branch. The published remote Go correction stage SHA was verified with `git ls-remote` equal to `b0c0d5fa96576eff8f7497d48e39fa30a5348e76`, and local Stage03-R, C dev-main and docs staging worktrees were verified clean.

Go commits relative to accepted T02-R:

```text
b0c0d5fa96576eff8f7497d48e39fa30a5348e76 mic: extend synthetic PCM-only geometry regressions
e90c9aa64d44110bfd5dd2589ee93d14ccd64bc5 mic: remove External raw metadata readers and validate native plain geometry
```

## 3. Changed files and exact removal

```text
M	echod/internal/hardware/mic/native.go
A	echod/internal/hardware/mic/native_pcmonly_test.go
M	echod/internal/hardware/mic/native_test.go
D	echod/internal/layout/raw_sources.go
D	echod/internal/layout/raw_sources_partial_fallback_test.go
D	echod/internal/layout/raw_sources_test.go
```

- Deleted `echod/internal/layout/raw_sources.go`, `raw_sources_test.go`, `raw_sources_partial_fallback_test.go` — all V3-only External raw identity/IDME/eMMC fallback reader and tests, including partial-IDME workarounds. No alternative External metadata path was added.
- `echod/internal/hardware/mic/native.go`: `plainNative` validates known ALSA signed PCM formats S16LE/S24_3LE/S32LE and physically consistent width/rate/frames, known MIC slots or complete stock MIC+REF channel layout only. Unknown format or extra/missing ambiguous channels cannot silently become S24 or REF. A requested native callback receives a clear error and never an invented descriptor; SoftwareCut fails closed on the captured block. Plain stock `Read()` with no native consumer retains existing behavior.
- Added/extended synthetic-only native capture regression tests: no-REF PCM-only descriptor, opaque absent/partial/giant IDME independence, unknown format/roles rejection, S16/S32 positive geometry, hardware REF L/R independently byte-exact, borrowed lifetime, software privacy and stock output retention.
- Preserved untouched stock `layout.Idme`, `readIDME`, `idmeFallback`, stock MAC/Color, Athena stock eMMC fallback, analog gain/hardware capture, `pairedCapture.convert`, XRUN decimator restart, Built-in mixer/AEC/denoise/leveling, noise/facing/standout/history/subscribers/calls/Panel. `git diff 15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc..b0c0d5fa96576eff8f7497d48e39fa30a5348e76` confirms no change to `echod/internal/layout/layout.go`, `athena_mainline.go` or `mac_idme.go`.
- `NativeBlock.Buffers`: borrowed original MIC/REF physical bytes, only valid until next capture Read. `Descriptor()`: only previously captured `asp.Descriptor` PCM geometry, **16+28×buffers+8×channels**, no metadata sections, no factory identity or calibration, no new wire format.
- `pairedCapture.ReadWithNative`: original MIC S32 and hardware stereo REF S16 available synchronously before stock conversion; REF absent stays absent as physical input. No DSP-specific resampling, downmix, calibration, MIC-slot mapping or vendor qualification in Go. Stock software REF remains stock-only.
- One SoftwareCut value obtained after capture at the common boundary, used for in-place MIC/REF sample clearing and the existing stock frame clear. No second privacy gate, resets on mute, new queues, copied native input, External selector or new DSP runtime.

## 4. Acceptance N01–N08 + correction regressions

| Test | Result and synthetic evidence |
|---|---|
| N01 | **PASS** paired native before stock convert; old Read path stays stock; `TestN01N02PairedOriginalStereoBeforeConversion`. |
| N02 | **PASS** plain MIC-only/no-REF, full interleaved stereo REF with independent nonzero markers, paired MIC and stereo REF byte-exact, no format conversion or input copies. |
| N03 | **PASS** single SoftwareCut decision after hardware MIC+REF input before paired conversion; native/stock both zeroed; sparse S32 packed sample/stride canaries and stock software REF covered. |
| N04 | **PASS** borrowed input backing store reused on next read, no copied retained native PCM. |
| N05 | **PASS** stock mono and ListenRaw/history retained frame stability, existing sensor/AEC/leveling code unchanged, stock tests. |
| N06 | **PASS** hardware-mute REF unchanged when software cut false, mute/unmute preserve stock lifecycle and XRUN reset only on overrun. |
| N07 | **PASS** no External metadata reader or descriptor fields; synthetic absent, partial, huge opaque factory bytes do not impact descriptor validity. Invalid ALSA formats, width mismatches, unknown extra channels rejected as native; no REF encoded with zero REF channels; S16LE/S32LE positive cases. |
| N08 | **PASS** relevant focused/race/vet and Go/C codec, complete default Go suite, default+dot+dot,donut+spot+panel focused/full compile/build, ARMv7 cross-build. |

All tests use only fake capture backends and synthetic fixtures. The complete default Go regression covers stock mic, privacy, hardware speaker, history, sensor and higher-level consumers. This is **not** hardware or DSP runtime acceptance.

## 5. Tests, build tags, logs, source SHA and limits

- Working directory `${WORKSPACE}/techo5-v3-stage03-pcmonly/echod`; Go `go version go1.26.0 linux/amd64`. `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOFLAGS=-p=1 GOMAXPROCS=2`, toolchain `${HOME}/.local/toolchains/go/bin`.
- Code source commit identity `b0c0d5fa96576eff8f7497d48e39fa30a5348e76`. Production code is identical across both Stage03-R commits; the second commit adds synthetic tests only. The matrix was initiated before the final two commits while source edits were being applied, so the original `environment.txt` records the starting HEAD, **not** a finished-SHA test identity; final focused/race/vet logs below explicitly recheck the clean committed tree.
- Matrix command runner `${WORKSPACE}/techo5-v3-stage03-pcmonly-matrix.sh`, full raw ledger `${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/results.tsv`. Logs and generated binaries reside in `${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031`. Each line of the ledger is the actual command and exit; no inferred PASS.

| Check | Exit | Exact executed command | Evidence SHA256 |
|---|---:|---|---|
| `focused` | `0` | `go test -count=1 -parallel 1 -v ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/asp` | `c0e91fef8d8f185bd1d1759e78c6e630a24060a7269d2d9aa0ea478437665485` |
| `race` | `0` | `go test -race -count=1 -parallel 1 ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/asp` | `ea432e2423c67e50166821931871fe9261c93eb3f7bc8269775287f7f00df776` |
| `vet` | `0` | `go vet ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/asp ./internal/lib/alsa` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `go-c` | `0` | `go test -count=1 ./internal/asp` | `a4a9194546c2640a3087bfd86807ccb35525301cdff48b0ae0888db50f60ffcc` |
| `full-default` | `0` | `go test -count=1 -parallel 1 ./...` | `0f6a4b4ec0ac3a01773283ce9df76412227a461cfa3a60717c6060ad741520f6` |
| `build-default` | `0` | `go build -o ${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/echod-default ./cmd/echod` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `test-dot` | `0` | `go test -count=1 -parallel 1 -tags dot ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/asp` | `5c447223ba49657f3fd1926d03875ba33df4d3a8427e133668c10d08e44b9ab7` |
| `compile-dot` | `0` | `go test -run \^\$ -parallel 1 -tags dot ./...` | `7471c9bd57d0bf7f87d96e57a5805d185d2babe8f91435b615e0817c94a582eb` |
| `build-dot` | `0` | `go build -tags dot -o ${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/echod-dot ./cmd/echod` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `test-dot-donut` | `0` | `go test -count=1 -parallel 1 -tags dot\,donut ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/asp` | `6819b98b8d247bc8bdb915959ca0ad9c1fab71dde67ebc8199b35e9a5e90e729` |
| `compile-dot-donut` | `0` | `go test -run \^\$ -parallel 1 -tags dot\,donut ./...` | `1cd00b1d39c5d74cb4794af7fe6b0f60b56079f333e5ca7e70e9e7ce81531ecb` |
| `build-dot-donut` | `0` | `go build -tags dot\,donut -o ${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/echod-dot-donut ./cmd/echod` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `test-spot` | `0` | `go test -count=1 -parallel 1 -tags spot ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/asp` | `550abd32b8e2220bbd7c81665a607449a4ce333823f20a9f195aa43ab26c6aad` |
| `compile-spot` | `0` | `go test -run \^\$ -parallel 1 -tags spot ./...` | `5a64cb6cca86300cf6191d54db3de6844d12741fd1478203a631a26e32026ef8` |
| `build-spot` | `0` | `go build -tags spot -o ${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/echod-spot ./cmd/echod` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `test-panel` | `0` | `go test -count=1 -parallel 1 -tags panel ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/asp` | `270f590114001a6908e710e4d371100fcdbdb2b844f2c525d6cec43e69246b25` |
| `compile-panel` | `0` | `go test -run \^\$ -parallel 1 -tags panel ./...` | `e764f94a3329fe0ffd459adff75a87ed7ac263ec9991998510055176d6a41e6d` |
| `build-panel` | `0` | `go build -tags panel -o ${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/echod-panel ./cmd/echod` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `armv7` | `0` | `env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o ${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/echod-armv7 ./cmd/echod` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |

### Additional post-commit exact-SHA checks

| Evidence file | SHA256 |
|---|---|
| `final-sha-focused.log` | `4ef5a493f1143adaeff7902370d85660716be9dd3e4708498ac48788e215d51e` |
| `final-sha-race.log` | `499563208d245825c5f200a75c1be1b29dd7090705863b3dde4725b950a4411c` |
| `final-sha-vet.log` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `final-sha-go-c.log` | `bd44f0b8aef95785ea8f5687df460be944c1f9467529ab7cd3e17c4608d2f9a3` |

- Go/C codec checked with `ASP_C_DECODER=${WORKSPACE}/frankenasp-v3-stage02-pcmonly/build/codec_probe` from accepted C02-R, **without changing C sources**.
- Test log SHA256 manifest: `${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/LOG_SHA256SUMS`; binary SHA256 manifest: `${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/BUILD_SHA256SUMS`.
- Android NDK/Bionic vendor runtime, factory Calibrator, real PCM geometry and real Echos were **not executed/verified**. ARMv7 is a cross-build proof, not a device deployment. No Stage04 helper/runtime/secure launch, live pipe exchange, HA mode or External output activation was implemented.
- No new platform whitelist, MIC4/REF2/96k admission, Q30/gain masks, metadata/provider abstraction, old CFG2 or dual codec.
- No STOP/blockers observed in measured software matrix. Unsupported plain native layouts explicitly fail for an attempted native consumer instead of inventing a descriptor; ordinary stock path remains unaffected.

## 6. Remote publication, clean ref post-check and coordinator handoff

Remote refs verified after direct validation host SSH push:

| Ref | SHA |
|---|---|
| Go `main` | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` |
| Go `experimental/external-asp-v3-dev-main` | `15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc` |
| Go `experimental/external-asp-v3-dev-stage03-pcmonly` | `b0c0d5fa96576eff8f7497d48e39fa30a5348e76` |
| Go `experimental/v3-design-staging` | `4130606cf09cba0d8a5fbc1f288195c0582816d4` |
| C `main` | `e986ae18643e08e95f29a38aae1dd74762f6362e` |
| C `experimental/v3-dev-main` | `77149e6494ac9784d98464ff984c84887a28ced9` |

This document is generated **outside** code/docs checkout at `${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/03_PCM_ONLY_CORRECTION_HANDOFF.md`, intended for docs coordinator publication to `docs/external-asp-v3/stages/03_PCM_ONLY_CORRECTION_HANDOFF.md` on `experimental/v3-design-staging`. The code executor has not committed to docs, integrated Stage03-R into either dev-main or touched any protected main.

**Historical executor gate:** Stage04 was blocked at the point the executor delivered the original report. **Updated coordinator gate:** independent Stage03-R technical **ACCEPTED**, Go dev-main integration **COMPLETE**, and published docs acceptance recorded here in §7. Stage04 can start under its own task, without hardware/Echo actions.

## 7. Independent coordinator review, acceptance and verified integration (2026-10-31)

**Original executor evidence:** SHA256 of the untouched, 128-line handoff at `${WORKSPACE}/techo5-v3-stage03-pcmonly-evidence-20261031/03_PCM_ONLY_CORRECTION_HANDOFF.md` is `97686dd5ac9b99a6aa5accd4e2446cf5c4415b0715a7ab764c8abb31ea5376ad`. Original output was a published stage branch with acceptance/integration PENDING. The pre-integration remote refs, test matrix and limitations in §§2–6 are historical executor snapshots; the accepted integration facts are recorded below.

**Decision: ACCEPTED**, based on independent review against PCM-only `PCM_ONLY_ARCHITECTURE_DECISION.md`, `MASTER_CONTRACT.md`, `SPEC_V2_APPROVED.md` and `stages/03_NATIVE_PRIVACY_TASK.md`. The reviewer verified the six-file diff from T02-R, including deletion of External-only `layout/raw_sources.go`, `raw_sources_test.go` and `raw_sources_partial_fallback_test.go`. No replacement External source reader was introduced. Stock `layout.Idme/readIDME/idmeFallback`, MAC/Color and capture implementation remain intact. `NativeBlock.Descriptor()` is PCM geometry only. Unknown ALSA formats, channel roles or widths are not guessed; requesting optional native access returns an error on unknown layouts while stock-only Read remains available. Hardware REF stereo L/R and paired raw samples remain borrowed and byte-exact; a single SoftwareCut decision still gates native and stock paths.

**Test provenance:** executor matrix 19/19 exit=0 plus final committed SHA focused/race/vet/Go↔C PASS; hashes `LOG_SHA256SUMS` (27) and `BUILD_SHA256SUMS` (6) independently verified. Coordinator additionally re-ran on committed `b0c0d5fa96576eff8f7497d48e39fa30a5348e76` focused MIC/layout/privacy/asp tests, `go test -race` on the same four packages, `go vet` on MIC/layout/privacy/asp/alsa, and a complete `go test -count=1 -parallel 1 ./...` with the accepted compiled C02-R probe — all PASS. No claim of independent rerun of all tagged or target-device tests; executor evidence covers default/dot/dot+donut/spot/panel and ARMv7 cross-build.

**Integration receipt:** input refs TECHO5 dev-main `15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc`, Stage03-R stage ref `b0c0d5fa96576eff8f7497d48e39fa30a5348e76`, FrankenASP dev-main `77149e6494ac9784d98464ff984c84887a28ced9`, documentation branch prior HEAD `4130606cf09cba0d8a5fbc1f288195c0582816d4`. From clean worktrees, after remote ref/ancestry and test checks, coordinator performed exactly one TECHO5 `git merge --ff-only b0c0d5fa96576eff8f7497d48e39fa30a5348e76` and `git push --porcelain origin HEAD:refs/heads/experimental/external-asp-v3-dev-main` on `validation host`, verified remote integration Go SHA equals T03-R. C dev-main was not moved, no empty C stage branch or C commit was created. Protected Go main `62603a7854283a3a07473a17c831c6ef0fd6f8d4`, C main `e986ae18643e08e95f29a38aae1dd74762f6362e` unchanged; worktrees clean. No force, rebase, reset, device access or old handoff rewrite.

**Final accepted Stage04 input pair:** **Go T03-R `b0c0d5fa96576eff8f7497d48e39fa30a5348e76` / FrankenASP C02-R `77149e6494ac9784d98464ff984c84887a28ced9`**.

**Final status:** Stage03-R Execution **PASS** · independent delegated Technical Acceptance **ACCEPTED** · Integration **COMPLETE**. This does not claim owner personally uttered an acceptance statement. Stage04 may begin under the PCM-only restart task; actual vendor Calibrator/launcher resource mapping, ARM32/Bionic production H07, live IPC and hardware tests remain outside Stage03-R scope.
