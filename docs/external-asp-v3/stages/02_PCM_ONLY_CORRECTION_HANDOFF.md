# Stage02-R — PCM-only descriptor correction: handoff

## 1. Identity and execution state

- **Stage:** Stage02-R PCM-only descriptor correction; authoritative owner decision `PCM_ONLY_ARCHITECTURE_DECISION.md`, approved 2026-10-31.
- **Execution:** **PASS**. The seven component gates C01–C07 passed; two correction branches are published and independently verified. Full app test status is shown exactly in the matrix below.
- **Technical acceptance: ACCEPTED (independent coordinator review, 2026-10-31).** The owner has directed the coordinator to manage technical decisions and integration. This is a delegated technical decision; not a claim that the owner personally said “Stage02-R accepted”.
- **Integration: COMPLETE (2026-10-31).** Both code dev-main refs were fast-forwarded directly from validation host to the exact reviewed PCM-only Go/C pair, and all remote SHAs and frozen main refs independently verified. Stage03-R can start only under its separate assignment; Stage04 remains blocked until Stage03-R acceptance/integration.
- **Software gate:** PCM-only codec/component Go↔C; **not** live wire, vendor DSP, acoustic or hardware gate.
- **Echo hardware, vendor runtime and acoustic tests:** **НЕ ВЫПОЛНЕНЫ**.
- Docs input: `dmit2k/techo5:experimental/v3-design-staging` **8c8ab3ea3deba1b3799f016e5a4cb88b701065ea**; actual current `SPEC_V2_APPROVED.md` blob **f3e1b00d1f5d184e345c0d2c79da8394435ba3a4**. Historical Stage02/03 SHA are evidence only; PCM-only decision takes precedence.
- Date/time report generated: 2026-10-31T14:26:43+0200, Europe/Riga nominal project date 2026-10-31.

## 2. Exact pair, branches and protected refs

| Field | TECHO5/Go | FrankenASP V3/C |
|---|---|---|
| Origin | `dmit2k/techo5` | `dmit2k/frankenasp-v3` |
| Input accepted integration SHA | `4fb329cf789b156bbeb8d13da20b8c4bda70f108` (T03) | `0f39216a5240306344af7c2457716f4eaa466f40` (C02) |
| Integration branch | `experimental/external-asp-v3-dev-main` | `experimental/v3-dev-main` |
| Correction stage branch | `experimental/external-asp-v3-dev-stage02-pcmonly` | `experimental/v3-stage02-pcmonly` |
| Worktree | `${WORKSPACE}/techo5-v3-stage02-pcmonly` | `${WORKSPACE}/frankenasp-v3-stage02-pcmonly` |
| Tested and committed candidate | **`15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc`** | **`77149e6494ac9784d98464ff984c84887a28ced9`** |
| Published remote stage SHA | **`15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc`** | **`77149e6494ac9784d98464ff984c84887a28ced9`** |
| Accepted tested candidate (delegated technical decision) | **`15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc`** | **`77149e6494ac9784d98464ff984c84887a28ced9`** |
| Output integration HEAD (coordinator verified) | **`15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc`** | **`77149e6494ac9784d98464ff984c84887a28ced9`** |
| Protected main (unchanged) | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` | `e986ae18643e08e95f29a38aae1dd74762f6362e` |
| Post-push worktree status | clean | clean |

Exact paired conformance receipt is **T_candidate=15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc; C_candidate=77149e6494ac9784d98464ff984c84887a28ced9**. Direct `git push --porcelain origin HEAD:refs/heads/<correction-branch>` was executed on validation host using existing repo-scoped SSH keys. Both remote stage refs verified by `git ls-remote`. The executor performed no merge/FF, reset, rebase, force push, GitHub commit recreation, docs edits or device operations. **Subsequently** the coordinator performed two explicitly authorized FF-only integration updates, documented in §9; no force/rebase/device actions.

## 3. Exact PCM-only wire contract

- **16-byte fixed LE header**: bytes 0–3 magic `ASP1`; 4–5 uint16 version=1; 6 uint8 physical buffer count, 7 uint8 channel count; 8–11 uint32 exact descriptor length, 12–15 uint32 physical input block bytes.
- **28-byte physical buffer record**, starting **at byte 16**: offsets 0/4/8/12/16 uint32 physical offset/length/rate/frames/stride, offset 20 format, 21 valid bits, 22 LSB shift, 23 number of channels, 24 uint16 first channel record, 26 uint16 reserved zero.
- **8-byte channel record**: byte 0 physical buffer index; byte 1 role (1 MIC, 2 REF); bytes 2–3 LE uint16 captured role index; bytes 4–7 LE uint32 sample offset inside frame.
- Formats S16LE=1, S24_3LE=2, S32LE=3, signed two's-complement; valid bits include sign (2..storage width), shift 0..(storage−valid). Signed24 in S32 at shift 0 and 8 is tested.
- **Strict descriptor length = `16+28*buffers+8*channels`**, where both counts are uint8 in 1..255. Maximum `16+255*28+255*8=9196` bytes. No PAD/TLV/raw section/trailing descriptor bytes. Input block length is independent uint32 and equals the sum of contiguous physical buffer lengths.
- Physical buffer offset continuity, nonzero rates/frames/length/stride, channel coverage/nonoverlap, buffer index consistency, reserved zero, captured MIC and REF indices each unique and gapless from 0, ≥1 MIC and absent REF permitted.
- Last sample safety: `offset+width<=stride`, `(frames-1)*stride+offset+width<=physicalLength`; widened arithmetic checks input total overflow.
- Old metadata-bearing **24/28/8 ASP1 v1 descriptors are rejected**, with no compatibility reader or migration negotiation. All Platform/Hardware/Factory/RawField/source-name/value-length/64KiB/64 entries/32KiB limits removed from production codec and fixture files. Byte-size checks are against actual encoded vectors, not static grep alone.
- Readiness **exactly one byte 0xA5**, unchanged; fixed unframed PCM exchange with one 640-byte mono S16LE 16kHz/320-sample response. No live process/DSP adaptation in this correction.
- Both hardware stereo REF channels L/R retain distinct nonzero markers `a1b2c3` and `d4e5f6` in the same original interleaved frame. Also tested independent three-buffer MIC/REF, REF absent, packed S24, signed24 low/high.
- Full wire documentation: `frankenasp-v3/WIRE_V1.md` (correction branch). The generic host never performs factory gain/IDME mapping; vendor resources remain specific to subsequent FrankenASP launcher implementation.

## 4. Changes and non-changes

**TECHO5 changed files** (`git diff --name-status 4fb329cf789b156bbeb8d13da20b8c4bda70f108..15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc`):
```
M	echod/internal/asp/descriptor.go
M	echod/internal/asp/descriptor_test.go
M	echod/internal/asp/testdata/SHA256SUMS
D	echod/internal/asp/testdata/large_factory.bin
D	echod/internal/asp/testdata/large_factory.expect
A	echod/internal/asp/testdata/max_255.bin
A	echod/internal/asp/testdata/max_255.expect
M	echod/internal/asp/testdata/packed_sparse.bin
M	echod/internal/asp/testdata/packed_sparse.expect
M	echod/internal/asp/testdata/ref_absent.bin
M	echod/internal/asp/testdata/ref_absent.expect
M	echod/internal/asp/testdata/s32_high24.bin
M	echod/internal/asp/testdata/s32_high24.expect
M	echod/internal/asp/testdata/s32_low24.bin
M	echod/internal/asp/testdata/s32_low24.expect
M	echod/internal/asp/testdata/separate_three.bin
M	echod/internal/asp/testdata/separate_three.expect
M	echod/internal/asp/testdata/stereo_s24.bin
M	echod/internal/asp/testdata/stereo_s24.expect
M	echod/internal/hardware/mic/native.go
M	echod/internal/hardware/mic/native_test.go
```

**FrankenASP V3 changed files** (`git diff --name-status 0f39216a5240306344af7c2457716f4eaa466f40..77149e6494ac9784d98464ff984c84887a28ced9`):
```
M	WIRE_V1.md
M	src/descriptor.c
M	src/descriptor.h
M	testdata/SHA256SUMS
D	testdata/large_factory.bin
D	testdata/large_factory.expect
A	testdata/max_255.bin
A	testdata/max_255.expect
M	testdata/packed_sparse.bin
M	testdata/packed_sparse.expect
M	testdata/ref_absent.bin
M	testdata/ref_absent.expect
M	testdata/s32_high24.bin
M	testdata/s32_high24.expect
M	testdata/s32_low24.bin
M	testdata/s32_low24.expect
M	testdata/separate_three.bin
M	testdata/separate_three.expect
M	testdata/stereo_s24.bin
M	testdata/stereo_s24.expect
M	tests/codec_probe.c
M	tests/make_vectors.py
M	tests/test_codec.py
```

- Go `echod/internal/asp/descriptor.go`: remove raw metadata types/sections/validation, encode exactly 16-byte header and PCM records.
- Go `echod/internal/hardware/mic/native.go`: remove dependent `NativeBlock.DescriptorWithRaw`, replace with `Descriptor()` that returns already captured PCM geometry; borrowed native PCM remains untouched.
- Go `native_test.go`: replace obsolete metadata-encoding test with descriptor PCM-only/no-copy assertion; do not change native capture/SoftwareCut behavior.
- C `src/descriptor.h/.c`: remove raw fields/parse and 64KiB bounding; enforce 9196-byte format with exact size check **before allocating** in `asp_read_fd`.
- C test-only `codec_probe.c`: exact byte decode, unaligned, strict isolated descriptor EOF check, stereo sample markers, byte-exact PCM test-only roundtrip; this is not a production DSP adapter.
- Python independent `make_vectors.py` uses Python `struct.pack`, not the production Go writer/C decoder. Legacy `large_factory` fixtures removed from both repos; `max_255` created. All other fixture binaries regenerated under new wire.
- **Preserved:** `layout.RawFactorySources` and stock IDME/capture paths (Stage03-R work); `mic.Source`, native capture backends, SoftwareCut, stock conversion, HA/config, secure helper launch, process/session lifecycle, vendor DSP/FIFO and production PCM publication untouched. No new hardware reads.
- No old CFG2/TLV/envelopes, dual decoder, fake calibration, installer/OTA, proprietary assets or legacy source imported. No protected main/dev-main/docs branch changes.
- Git diff summary: Go 21 files / +611/−137 lines including tests and binary file metadata; C 23 files / +781/−334 lines including fixtures/docs. Use `git diff --numstat` for per-file production LOC; binary and test files intentionally not attributed to production LOC.

## 5. Independent fixtures / SHA256

- The two repositories contain **identical manifest and fixture bytes**, verified with `cmp`, recursive `diff -qr` and `sha256sum -c`.
- **`testdata/SHA256SUMS` SHA256: 9d26d6cca1e29b10f1abaee583da12ee918ea4252f16f70c3eb2c01531e1d2f4**.
- Seven PCM-only fixtures: `stereo_s24` (92 bytes), `separate_three` (132), `s32_low24` (60), `s32_high24` (60), `ref_absent` (60), `packed_sparse` (60), `max_255` (**9196**). Each has `.bin` and `.expect`; physical PCM vectors for stereo and separate captures; `readiness.bin` exactly **A5**; `output_contract.txt` 640-byte output statement. All 18 fixture files are hashed in manifest.
- Native compiled C probe SHA256: **423084acb8c414d8178b6ebfab39d3fcc40f61eab549f4fafebb585aa249cf3b**; sanitized C probe SHA256: **7caaa5f04f5e772a3b29871914f935e63ce1ed8fe6de76715b28e88c5cf3dc93**.
- Go default amd64 build SHA256 **6fc3c1147a0d6eefaea5f2fc032036a5edf2d55c5c3d6ce0bd87b6d1ede9f25d**; armv7 cross-build SHA256 **1e06e3210fadf9fa3c5cb9b7c017a0aab2ac0dcca3a79249b38ec83928da3cb6**.
- Exact expected bytes produced independently; C textual transcript compared with oracle, Go encoder bytes compared with exact binary fixture, Go→compiled C reader checked on same candidate pair.

## 6. C01–C07 tests, build matrix and evidence

Host `validation host`, toolchains Go 1.26.0 linux/amd64, GCC 13.3.0, Python 3.12.3. Commands run with:
```sh
export PATH=${HOME}/.local/toolchains/go/bin:$PATH
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOFLAGS=-p=1 GOMAXPROCS=2
export ASP_C_DECODER=${WORKSPACE}/frankenasp-v3-stage02-pcmonly/build/codec_probe
```

Matrix script: `${WORKSPACE}/techo5-v3-stage02-pcmonly-evidence-runner.sh`; corresponding execution logs and SHA256 manifest are in `${WORKSPACE}/techo5-v3-stage02-pcmonly-evidence-20261031/`. Final reconciled ledger: `results-clean.tsv`; raw `results.tsv` retains the truth about duplicate invocations interrupted when authorized terminal tools returned `upstream_unavailable`. Only one uninterrupted successful instance per test is claimed below. No interrupted invocation is silently presented as a test PASS.

| Test | CWD | Exact command | Exit | Evidence |
|---|---|---|---|---|
| `c-compile` | `${WORKSPACE}/frankenasp-v3-stage02-pcmonly` | `make clean && make all` | **0** | `c-compile.log` |
| `c-conformance` | `${WORKSPACE}/frankenasp-v3-stage02-pcmonly` | `make test` | **0** | `c-conformance.log` |
| `c-sanitizers` | `${WORKSPACE}/frankenasp-v3-stage02-pcmonly` | `make sanitize` | **0** | `c-sanitizers.log` |
| `c-fixture-manifest` | `${WORKSPACE}/frankenasp-v3-stage02-pcmonly` | `cd testdata && sha256sum -c SHA256SUMS` | **0** | `c-fixture-manifest.log` |
| `go-c-exact` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -count=1 -parallel 1 -v ./internal/asp` | **0** | `go-c-exact.log` |
| `go-race` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -race -count=1 -parallel 1 ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy` | **0** | `go-race.log` |
| `go-vet` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go vet ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/lib/alsa` | **0** | `go-vet.log` |
| `go-stock` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -count=1 -parallel 1 ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy` | **0** | `go-stock.log` |
| `go-default-build` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go build -o '${WORKSPACE}/techo5-v3-stage02-pcmonly-evidence-20261031/echod-linux-amd64' ./cmd/echod` | **0** | `go-default-build.log` |
| `go-armv7-build` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o '${WORKSPACE}/techo5-v3-stage02-pcmonly-evidence-20261031/echod-linux-armv7' ./cmd/echod` | **0** | `go-armv7-build.log` |
| `tag-dot` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -count=1 -parallel 1 -tags 'dot' ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy` | **0** | `tag-dot.log` |
| `tag-dot-donut` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -count=1 -parallel 1 -tags 'dot,donut' ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy` | **0** | `tag-dot-donut.log` |
| `tag-spot` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -count=1 -parallel 1 -tags 'spot' ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy` | **0** | `tag-spot.log` |
| `tag-panel` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -count=1 -parallel 1 -tags 'panel' ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy` | **0** | `tag-panel.log` |
| `shared-manifest` | `${WORKSPACE}/frankenasp-v3-stage02-pcmonly` | `cmp testdata/SHA256SUMS '${WORKSPACE}/techo5-v3-stage02-pcmonly/echod/internal/asp/testdata/SHA256SUMS' && diff -qr testdata '${WORKSPACE}/techo5-v3-stage02-pcmonly/echod/internal/asp/testdata'` | **0** | `shared-manifest.log` |
| `go-all` | `${WORKSPACE}/techo5-v3-stage02-pcmonly/echod` | `go test -count=1 -parallel 1 ./...` | **0** | `go-all.log` |

- **C01**: Python independently derived exact byte fixtures; Go writer exact equality, compiled C decode transcripts, one-byte READY.
- **C02**: S16/S24_3/S32, 24 signed low/high, physical buffer/channel order, three physical buffers, packed valid bits/shift/stride, REF absent.
- **C03**: **actual encoded length and exact header boundaries** verified for all seven vectors, no metadata section possible, old 24-byte descriptor rejected, isolated trailing descriptor rejection.
- **C04**: **44 malformed/truncated rejection tests** require exact C exit **2**, not an arbitrary crash, including invalid sizes, old wire, buffers, roles, role indices, last sample, frame count, stride, offsets, overflow, old CFG2, truncation.
- **C05**: L/R distinct nonzero bytes and test-only roundtrip of complete stereo and separate physical PCM blocks; `max_255.bin` exactly 9196 bytes.
- **C06**: C compiler strict warnings-as-errors; C ASan/UBSan and unaligned decoder; Go unit/race/vet. C failure cannot be credited as an ASan crash.
- **C07**: both components compiled at exact candidate commits, Go and C fixtures manifest SHA256/contents equal, native Go default app build and ARMv7 cross-compile; focused stock/tag matrix.
- **Postcommit verification**: `${WORKSPACE}/techo5-v3-stage02-pcmonly-evidence-20261031/postcommit-c.log` and `postcommit-go-c.log` pass on actual committed SHA pair.
- **Baseline status**: no observed new failures in focused stock tests or supported tag matrix. Full application `go test ./...` is tracked independently in `go-all.log` (no fake PASS).
- Not run: Android NDK/Bionic vendor backend compilation, vendor DSP/launcher, hardware/real calibration, Echo, acoustic tests. Stage02-R is codec-only; no vendor DSP support claim.

## 7. Risks / limitations / corrective status

1. **No actual PCM runtime helper qualification:** C decoder does not prove DSP support for all represented layouts. The DSP-specific launcher aliases/Calibrator and resource handling belong to Stage04.
2. **Stage03-R still required:** the Stage03-added External raw readers remain in `echod/internal/layout/raw_sources.go` by explicit authorization; the standalone PCM-only encoder does not invoke them. Their dedicated deletion and N01–N08 no-IDME confirmation belong to Stage03-R, not Stage02-R.
3. **Wire incompatibility intentional:** old 24-byte descriptor rejected; Go and C must be integrated as exact accepted pair before any runtime.
4. **Full app suite:** PASS (exit 0); see ledger for actual state. Codec C01–C07 and stock-focused/tag tests passed.
5. No secret credentials, proprietary vendor binaries or altered Git permissions committed.
6. Transient authorized terminal tools `upstream_unavailable` led to duplicated runner invocations; two duplicate full-suite children were explicitly terminated. The raw ledger records exit 143 for those duplicates; one surviving run is recorded separately. This is an execution-system artifact and not a source-code test failure.

## 8. Coordinator publication / next gate

- Prepared **outside all code/docs worktrees** as `${WORKSPACE}/techo5-v3-stage02-pcmonly-evidence-20261031/02_PCM_ONLY_CORRECTION_HANDOFF.md`.
- Intended docs publication path: `docs/external-asp-v3/stages/02_PCM_ONLY_CORRECTION_HANDOFF.md` on `experimental/v3-design-staging`, by **documentation coordinator only**.
- **Coordinator technical acceptance ACCEPTED; integration COMPLETE**. Both remote dev-main SHA were verified equal to the exact tested candidates, after original-SHA fast-forward. Stage03-R is now unblocked by its separate assignment. Stage04 remains prohibited until subsequent Stage03-R acceptance/integration.
- This report began as an executor handoff (original SHA256 below), with acceptance/integration pending. This coordinator-published edition **includes the independent acceptance decision and actual integration receipt in §9**. The code executor did not change documentation ref.

## 9. Independent coordinator acceptance and two-repository integration receipt (2026-10-31)

**Decision:** Stage02-R codec/conformance **ACCEPTED**. The coordinator independently reviewed production Go `asp.Encode()`, C `asp_decode()`/`asp_read_fd()`, actual physical MIC/REF representation, the complete executor test ledger and the approved PCM-only contract. No raw IDME/calibration/identity fields, source identifiers or legacy decoder remain in this codec pair. Descriptor is exactly 16+28×buffer_count+8×channel_count, max 9196 bytes, with readiness `0xA5`. The isolated old 24-byte descriptor and malformed cases reject cleanly. Native capture/SoftwareCut/stock code was not refactored by Stage02-R; External-only raw reader deletion and native geometry edge cases belong to Stage03-R.

**Re-run on committed candidate SHAs:** C `make test` and `make sanitize` PASS, C `testdata/sha256sum -c` PASS (run as `cd testdata && sha256sum -c SHA256SUMS`), Go `go test -count=1 -parallel 1 -v ./internal/asp` with `ASP_C_DECODER` pointing to the compiled C probe PASS, and `go test -race -count=1 -parallel 1 ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy` + `go vet ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/lib/alsa` PASS. Both complete Go/C fixture directories compared byte-exact using `diff -qr`, and their identical SHA256 manifests verified. Evidence directory `SHA256SUMS` verified all original logs and ledger. The executor's full Go application test and tags remain recorded verbatim in §6; no independent full-app re-run is claimed.

**Integrated pair (GitHub remote refs independently verified):**
- TECHO5 `experimental/external-asp-v3-dev-main`: **`15b2f1f73106fb9149fcdfacbebe4335e2e6fbbc`** (fast-forward from `4fb329cf789b156bbeb8d13da20b8c4bda70f108`; one new commit).
- FrankenASP V3 `experimental/v3-dev-main`: **`77149e6494ac9784d98464ff984c84887a28ced9`** (fast-forward from `0f39216a5240306344af7c2457716f4eaa466f40`; one new commit).
- Both published correction stage refs remain at exactly these SHAs, both protected main refs remain at TECHO5 `62603a7854283a3a07473a17c831c6ef0fd6f8d4` and FrankenASP `e986ae18643e08e95f29a38aae1dd74762f6362e`. Docs/stage/integration worktrees verified clean at integration time.
- Integration commands on `validation host` were `git merge --ff-only <accepted SHA>` in each integration worktree followed by direct authorized `git push --porcelain origin HEAD:refs/heads/<corresponding-dev-main>`. Both remote refs verified by `git ls-remote`; no force, reset, cherry-pick, squash or additional code commit.

**Original executor handoff SHA256:** `137ffce05c4fa9f032a21d41cf22ee82febc0b334a39e84a96801526913b4171`, preserved unchanged at `${WORKSPACE}/techo5-v3-stage02-pcmonly-evidence-20261031/02_PCM_ONLY_CORRECTION_HANDOFF.md`. Historical report statements about PENDING dev-main capture the pre-coordinator snapshot; this section and the revised status table are authoritative final acceptance/integration facts.

**Final gate:** Execution **PASS**; delegated technical acceptance **ACCEPTED**; integration **COMPLETE**. The exact **T02-R/C02-R** pair above is now the mandatory input to **Stage03-R**. No Stage03-R code was started by this acceptance; no Echo/vendor runtime or acoustics were exercised. **Stage04 remains blocked** until Stage03-R independent acceptance and integration.
