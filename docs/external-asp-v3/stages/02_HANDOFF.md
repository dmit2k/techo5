# TECHO5 / FrankenASP V3 — Stage02 Descriptor Codec & Go/C Conformance — HANDOFF

> **Publication:** coordinator publishes this completed Stage02 delivery report as `docs/external-asp-v3/stages/02_HANDOFF.md` on TECHO5 `experimental/v3-design-staging`. Historical original draft was prepared outside code worktrees (SHA256 `2dc3cd99b8e42887fe0a12a4f3b1ae17775f30c9433132135edbeef4279e9deb`); this publication updates its delivery status while preserving prior failure evidence.

## 1. Identity, authority and statuses

- Stage: **02 — Descriptor Codec & Go/C Conformance**, `stages/02_DESCRIPTOR_CODEC_TASK.md`.
- Execution date/time: **2026-10-31, Europe/Riga**; host `anfisa-trainer`; executor **ChatGPT Stage02 Executor**.
- Normative docs HEAD: `dmit2k/techo5:experimental/v3-design-staging` at `77536bdcd15519eab7d28c0dab2ea5d44f060591`.
- Exact approved SPEC blob: `b324d2161e5ea316618f23bbe6d6aedf6c2df5c8`.
- Stage01 `stages/01_HANDOFF.md`: execution PASS, **owner ACCEPTED**, bootstrap integration COMPLETE, with previously documented limitations.
- **Codec component checks C01–C07: PASS** (all tested on final local Go/C candidate commits).
- **Overall Stage02 execution: PASS (2026-10-31)**. C01–C07 were completed on the exact Go/C candidate commits. Initial HTTPS/SSH push from `anfisa-trainer` failed, but delivery blocker was **resolved** by transferring the original Git objects via SSH to authorized `gptlab-core` and pushing the unchanged commits with its existing GitHub CLI authentication. Both remote stage refs and unchanged main/dev-main refs were independently verified. Focused Go↔compiled C and C codec tests were also re-run successfully after publication.
- **Acceptance: ACCEPTED (delegated technical decision, 2026-10-31).** The owner requested independent analysis and a decision («мы можем принять Stage 02? анализ и принятие решения! ТЫ должна вести меня через реализацию согласованного проекта!»). Acting in the entrusted coordinator/reviewer role, a fresh review accepted the exact Go/C codec pair within Stage02 scope. This does not assert that the owner personally uttered «принимаю Stage02». Independent tests were rerun: C `make test` and `make sanitize` PASS; Go `go test -race -count=1 ./internal/asp`, `go vet ./internal/asp`, and stock `go test -count=1 ./internal/hardware/mic ./internal/layout` PASS; both copies of `SHA256SUMS` verified.
- **Integration: PENDING**. Neither integration ref was moved; Stage03 is NOT authorized.
- Software gate: **codec component only**, not paired runtime/Source or hardware.
- Hardware, Echo deployment, acoustic measurement and vendor DSP runtime on Echo: **НЕ ВЫПОЛНЕНЫ**.

## 2. Exact Git revision ledger

| Field | TECHO5 / Go | frankenasp-v3 / C |
|---|---|---|
| Repository | `dmit2k/techo5` | `dmit2k/frankenasp-v3` |
| Accepted T01/C01 input integration SHA | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` | `e986ae18643e08e95f29a38aae1dd74762f6362e` |
| Integration branch | `experimental/external-asp-v3-dev-main` | `experimental/v3-dev-main` |
| New Stage02 branch | `experimental/external-asp-v3-dev-stage02` | `experimental/v3-stage02` |
| Worktree | `/home/gptlab/dev/techo5-v3-stage02` | `/home/gptlab/dev/frankenasp-v3-stage02` |
| Local Stage02 commits | `78a01f897b89dd462576e6e617fbc8216c4518dd` | `0081a6dc9bc5b9a235af38f7f87ffbbcbcd553ad`, then `0f39216a5240306344af7c2457716f4eaa466f40` |
| Tested local candidate SHA | **`78a01f897b89dd462576e6e617fbc8216c4518dd`** | **`0f39216a5240306344af7c2457716f4eaa466f40`** |
| Published Stage02 SHA | **`78a01f897b89dd462576e6e617fbc8216c4518dd`**, verified remote | **`0f39216a5240306344af7c2457716f4eaa466f40`**, verified remote |
| Accepted candidate (delegated technical decision) | **`78a01f897b89dd462576e6e617fbc8216c4518dd`** | **`0f39216a5240306344af7c2457716f4eaa466f40`** |
| Output integration SHA | Unchanged: `62603a7854283a3a07473a17c831c6ef0fd6f8d4` | Unchanged: `e986ae18643e08e95f29a38aae1dd74762f6362e` |
| Worktree status | `git status --porcelain` empty | `git status --porcelain` empty |

Final **locally tested** candidate pair:
`T_candidate=78a01f897b89dd462576e6e617fbc8216c4518dd`;
`C_candidate=0f39216a5240306344af7c2457716f4eaa466f40`.

After successful original-SHA stage publication, remote `main` and code dev-main were independently verified unchanged:
- TECHO5 main/dev-main: `62603a7854283a3a07473a17c831c6ef0fd6f8d4`.
- C main/dev-main: `e986ae18643e08e95f29a38aae1dd74762f6362e`.
- Documentation design-staging: `77536bdcd15519eab7d28c0dab2ea5d44f060591`.
- During the initial failed push, GitHub returned **404 Not Found** for the not-yet-created stage refs. After authenticated publication via `gptlab-core`, both remote refs exist with the exact tested local SHAs. No force, reset, rebase, merge, main or dev-main write occurred; the only separate docs write is this coordinator handoff publication.

## 3. Binary contract, raw metadata, readiness

### Fixed wire records

- Exactly one session descriptor; magic ASCII `ASP1`; version uint16 LE 1.
- Header **24 bytes**: offset 0 magic[4]; offset 4 version uint16; offsets 6/7 buffer/channel counts uint8; offset 8 descriptor length uint32; offset 12 PCM input block length uint32; offset 16 platform raw section length uint16; offset 18 hardware raw section length uint16; offset 20 factory raw section length uint32.
- Buffer record **28 bytes**, for each physical buffer in transfer order: offsets 0/4/8/12/16 = uint32 LE physical offset, physical length, sample rate, frame count, frame stride; offsets 20/21/22/23 = uint8 format, valid bits, LSB shift, channel count; offset 24 = uint16 first-channel record; offset 26 = zero uint16 reserved.
- Channel record **8 bytes**: uint8 buffer index, uint8 role (1 MIC / 2 REF), uint16 captured role index, uint32 frame byte offset.
- Formats 1 S16LE (2 bytes), 2 packed S24_3LE (3 bytes), 3 S32LE (4 bytes); signed two's complement; valid bits include the sign bit (2..storage bits); shift 0..storage-valid; signed24 in S32 valid bits=24, shift 0 or 8.
- All buffers positive size, rate, frames, stride; contiguous physical byte offsets starting at zero, sum matching header total; no required generic upper cap on PCM block other than wire uint32. Channel groups contiguous per buffer and cover every record exactly once; unique, gapless MIC indices 0..M-1 and REF indices 0..R-1; at least one MIC, REF may be absent. No sample byte overlap within a frame.
- Bounds use widened intermediates: `offset+sampleWidth <= stride` and `(frames-1)*stride+offset+sampleWidth <= physicalLength`; no uint32 sum overflow. Buffer reserved must be zero; descriptor length and section lengths must match exactly.
- No imposed MIC4/REF2/96kHz, 2-buffer or DSP compatibility rule in host codec. Valid descriptor does **not** imply vendor DSP acceptance.

### Minimal raw representation

After records, three raw sections are concatenated in fixed header order: **platform/identity (section 1), capture hardware (section 2), factory/IDME (section 3)**. Sections are identified by their fixed header positions; there is no per-entry type registry. Each section length zero means absent.

Each source entry consumes `uint8 sourceNameLength`, `uint16 LE rawValueLength`, exact ASCII source-name bytes and exact original raw value bytes. Names are 1..63 ASCII bytes from `[A-Za-z0-9/_.-]` (e.g. `/proc/idme/miccal.0`); value lengths 0..32768; embedded zeros, trailing whitespace and 0x80..0xFF values are preserved. Source name unique within its section; zero-length present entries and absent entries differ. Values are not semantically parsed or associated with captured MIC slots by Go.

Bound: descriptor **<=65536 bytes**, **<=64 raw entries** overall, individual raw value **<=32768 bytes**; platform/hardware lengths constrained by uint16. Header/buffer/channel records are fixed 24/28/8. This accommodates the 12000-byte raw factory fixture (>8192) without allocating max descriptor bytes per PCM block; `asp_read_fd` allocates exactly the advertised startup descriptor length after checking its header. Actual device-specific raw source lengths **were not measured**; 64 KiB is a conservative explicit format policy bound, not a proven physical device maximum. Oversize rejects rather than truncates.

TECHO5 upstream `echod/internal/layout/layout.go` `Idme()` trims NUL/whitespace; `echod/internal/layout/athena_mainline.go` `parseIdme()` returns interpreted strings. Stage02 does **not** call these for raw values and intentionally does not add the hardware raw-reading tap (future stages only). No normalized `PlatformInfo`, `MicFactoryGain`, Q30, masks, confidence, availability, MIC gain/slot association or platform qualification exists in new Go code.

### Readiness and PCM

- One exact byte **hex `a5` / 0xA5** is the one-time readiness confirmation after successful helper initialization; `testdata/readiness.bin` is exactly one byte. No ASCII `READY`, status/error message protocol, repeated handshake or per-frame envelope.
- Later IPC runtime is NOT implemented in Stage02. Its fixed contract is one native input block of header size `inputBytes` followed by exactly **640 output bytes**, mono S16LE, 16000 Hz, 320 samples / 20ms. C/Go constants and `testdata/output_contract.txt` record these values; this does not process live PCM.
- Byte-exact REF proof: `stereo_s24.pcm` captures distinguishable frame-0 REF L bytes `a1 b2 c3` and REF R bytes `d4 e5 f6`, with further distinct samples in frame 1. Compiled C decoder/test-only probe reads each captured channel independently, without conversion.
- Full field-by-field reference: `frankenasp-v3/WIRE_V1.md` in local Stage02 branch.

## 4. Actual changed files and scope

**Go (21 newly added files):**
- `echod/internal/asp/descriptor.go` — encoder, value types, size/count/geometry/raw validators (201 lines production).
- `echod/internal/asp/descriptor_test.go` — exact independently generated bytes, compiled C cross-conformance, readiness, stereo markers, invalid writer inputs, manifest verification.
- `echod/internal/asp/testdata/`: `SHA256SUMS`, `readiness.bin`, `output_contract.txt`, and seven fixture stems `stereo_s24`, `separate_three`, `s32_low24`, `s32_high24`, `ref_absent`, `packed_sparse`, `large_factory`, each with `.bin` and `.expect`; stereo and separate cases also include `.pcm`.

**C (27 newly added files):**
- `src/descriptor.h` (42 production lines) — bounded API and no-copy raw views;
- `src/descriptor.c` (133 production lines) — explicit LE C decode, validation, exact-sized descriptor read; no DSP/runtime;
- `tests/codec_probe.c` — compiled test-only decoder and test-only PCM sample inspector;
- `tests/make_vectors.py` — independent Python `struct.pack` specification-derived fixture oracle;
- `tests/test_codec.py` — fixture bytes/transcript and bad-input tests, including precise exit 2 expected under sanitizers;
- `Makefile`, `.gitignore`, `WIRE_V1.md`;
- `testdata/` — same 19 files as Go testdata (SHA256SUMS plus 18 checksummed files).

Production new LOC: Go **+201**, C source/header **+175**, no production code removed. Tests, fixtures and documentation counted separately. Old code imported/cherry-picked: **NONE**. No proprietary vendor assets, installer, OTA or release automation included.

Prohibited areas unchanged: `mic.Source`, capture backends, stock PCM conversion, SoftwareCut, HA/config, Selected/Active, secure launcher, process/session lifecycle, FIFO, DSP bridge, vendor runtime, production PCM publication. New production code has no CFG2/TLV, old envelopes, normalized metadata, universal registry or production synthetic adapter. Only test-only C probe handles sample marker inspection.

## 5. Test evidence C01–C07, exact commands and results

Environment on `anfisa-trainer`:
`export PATH=/home/gptlab/.local/toolchains/go/bin:$PATH GOMAXPROCS=2 GOFLAGS=-p=1 ASP_C_DECODER=/home/gptlab/dev/frankenasp-v3-stage02/build/codec_probe`.

Toolchains: **Go 1.26.0 linux/amd64**, **GCC 13.3.0**, **Python 3.12.3**, native offline Linux amd64. Tests run on exact local candidate pair above. The authoritative complete command/exit/log ledger is `/tmp/asp-v3-stage02-final-matrix.tsv` (**SHA256 `3031112c5c1a59637c05a8b4d781fcb56f0fb4789884d3b33563fed4f381a9f4`**).

| Test ID | Working directory | Exact command | Exit | Evidence |
|---|---|---|---|---|
| C07 C build | `/home/gptlab/dev/frankenasp-v3-stage02` | `make clean && make all` | **0 PASS** | `/tmp/asp-v3-stage02-C07_C_BUILD.log` |
| C01/C02/C04/C05/C06 C conformance | C worktree | `make test` | **0 PASS** | `/tmp/asp-v3-stage02-C01_C04_C06_C_TEST.log` |
| C06 ASan+UBSan | C worktree | `make sanitize` | **0 PASS** | `/tmp/asp-v3-stage02-C06_ASAN_UBSAN.log` |
| C01/C02/C03 paired Go writer → compiled C decoder | `/home/gptlab/dev/techo5-v3-stage02/echod` | `go test -count=1 -v ./internal/asp` | **0 PASS** | `/tmp/asp-v3-stage02-C01_C02_C03_GO_C.log` |
| C06 Go race | Go echod directory | `go test -race -count=1 ./internal/asp` | **0 PASS** | `/tmp/asp-v3-stage02-C06_GO_RACE.log` |
| C06 Go vet | Go echod directory | `go vet ./internal/asp` | **0 PASS** | `/tmp/asp-v3-stage02-C06_GO_VET.log` |
| C07 Go build | Go echod directory | `go build ./internal/asp` | **0 PASS** | `/tmp/asp-v3-stage02-C07_GO_BUILD.log` |
| C07 Go test executable | Go echod directory | `go test -c -o /tmp/techo5-v3-stage02-asp.test ./internal/asp` | **0 PASS** | `/tmp/asp-v3-stage02-C07_GO_TEST_BINARY.log` |
| C07 stock regression | Go echod directory | `go test -count=1 ./internal/hardware/mic ./internal/layout` | **0 PASS** | `/tmp/asp-v3-stage02-C07_STOCK_MIC_LAYOUT.log` |
| C07 manifest/hash consistency | C worktree | `cmp -s testdata/SHA256SUMS /home/gptlab/dev/techo5-v3-stage02/echod/internal/asp/testdata/SHA256SUMS && (cd testdata && sha256sum -c SHA256SUMS)` | **0 PASS** | `/tmp/asp-v3-stage02-C07_MANIFEST_COMPARE.log` |

**Detailed cases:** seven independently specified valid binary descriptors; S16LE, S24_3LE, S32LE, signed24 low/high, packed 20-bit, channel order, one or three physical buffers, REF absent, separated REF, interleaved stereo REF, named raw values with NUL/high-bit bytes and absent/zero-length sources. Raw factory fixture 12082 descriptor bytes (12,000 raw bytes). Seven unaligned pointer C decodes pass. Forty-one malformed/truncated cases **reject with precise C exit 2** under both normal and ASan/UBSan builds; no sanitizer crash is accepted as a valid rejection. Invalid cases include magic CFG2, version, header/section/count mismatches, reserved bytes, PCM formats/bits/shifts, bad offsets/gaps/overlaps, roles/role-index continuity, sum overflow, last sample beyond physical length, duplicate or invalid raw source, oversized value/descriptor, and multiple truncated reads.

**Manifest:** `testdata/SHA256SUMS` identical in both repos, file SHA256:
`7908f5a59cf32b52a568c26af56e6776225ae3d7b8bac63f0aaa4e448fcce2ef`.
It lists SHA256 for all 18 other testdata files; manifest/contents passed in both Go/C. Generated fixture bytes are independent of either production codec, from committed `tests/make_vectors.py`.

**Compiled artifact identities (native offline):**
- C `build/codec_probe`: SHA256 `2d26153ed8efa8cd401daad5da2445ee2ddb68dd434ec3103c80b8ae4a517cbb`.
- C `build/codec_probe_san`: SHA256 `bc30c2d38d7c2decfa4881a07ec152fcdc6e4049003c44c4d439f00dbaeb3bdb`.
- Go compiled `/tmp/techo5-v3-stage02-asp.test`: SHA256 `e5378d716dd8dd08330854d1e4579fb0d02908c442d61442b38534dc45fd8158`.

**Log hashes:** `C_TEST` e15ed2e4acb3eb2cf42776cd103c02feab56715034e03631755485cb5fa87b53; `ASAN/UBSAN` 1a4c0227ca6b4a509acf38a467b18f9a45f41ddf9e6a4f49b3ab5aa2cedefd3c; `GO_C` 5f0152e946b09a4a62a143d5c32b2d4ef82519526efe55febcfdbaf4f7a7f344; `GO_RACE` 63e54d44975dcdb44fae015fc14740a02de21a61cda1cbc69b0c42f8f11d878f; `GO_VET` e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855; `STOCK` 40aaba8048070a859d108b24435491c88a4fe64d40b6b58a215aabd083cc829e.

All tests are codec component/offline; no whole application Stage02 ARM32/Android build, vendor ABI qualification, real raw IDME device measurement, real launch/PCM exchange, capture mute, SoftwareCut, FIFO, ownership, process lifecycle, hardware or acoustic acceptance was attempted. Stage02 does not claim these tests.

## 6. Resolved delivery blocker: Stage02 stage branch push

**Exact attempted command:**
`git -C /home/gptlab/dev/techo5-v3-stage02 push origin refs/heads/experimental/external-asp-v3-dev-stage02:refs/heads/experimental/external-asp-v3-dev-stage02`.

**Observed:** `fatal: could not read Username for 'https://github.com': No such device or address` (**exit 128**). Because the command was first in a `set -e` pair, the C push was **not attempted**. Host has no Git HTTPS credential helper, no `gh` binary, and SSH to `git@github.com` was rejected with `Permission denied (publickey)`. The connected GitHub API returned 404 for both Stage02 refs. Local commits and all worktrees remain intact. No reset/rebase/force-push or branch deletion performed.

**Actual remediation (completed):** `gptlab-core` has a working GitHub CLI session for `dmit2k` and SSH access to `anfisa-trainer`. The coordinator checked frozen main/dev-main and absence of the stage refs, fetched original Git objects to temporary isolated bare repositories over SSH, verified ancestry and exact candidate commit IDs, then performed these pushes from the core host:

```sh
git -C <isolated-core-go.git> -c credential.helper='!gh auth git-credential' push https://github.com/dmit2k/techo5.git 78a01f897b89dd462576e6e617fbc8216c4518dd:refs/heads/experimental/external-asp-v3-dev-stage02
git -C <isolated-core-c.git> -c credential.helper='!gh auth git-credential' push https://github.com/dmit2k/frankenasp-v3.git 0f39216a5240306344af7c2457716f4eaa466f40:refs/heads/experimental/v3-stage02
```

Both remote branch HEADs **were verified equal** to the exact tested pair via GitHub API. Both protected main and both code integration dev-main remain frozen at the Stage01 SHAs. The temporary transfer directories were removed. No code SHA changed, so no test invalidation occurred. Recheck: `go test -count=1 ./internal/asp` (with `ASP_C_DECODER` set to the compiled C probe) PASS and `make test` (compiled C, stereo L/R markers, 7 fixtures, 41 malformed cases) PASS. Initial push failure remains documented as a historical incident, not the final execution status.

Authentication failure was an **execution-environment issue**, not a descriptor design defect. It is **RESOLVED**; C07 remote publication now satisfies the exact-SHA gate.

## 7. Forward handoff and acceptance

- CODE is locally committed and offline-tested; **both original candidate commit SHAs are published as remote Stage02 refs and verified**.
- **Delegated technical acceptance ACCEPTED**, **integration PENDING**. Stage03 **must not start** until both published Stage02 candidate refs are explicitly integrated via a separately authorized Git integration action into their respective dev-main branches and both resulting remote HEADs verified.
- The documentation coordinator publishes this Stage02 report with the delegated technical acceptance decision and its evidence. The code executor did not change design-staging; actual dev-main integration remains PENDING.
- No hardware/vendor/acoustic claims are made and no Echo operations are permitted by this work.

## 8. Documentation publication boundary

The original draft at `/home/gptlab/dev/v3-stage02-handoff/02_HANDOFF.md` was retained as historical evidence of the initial publication blocker. This completed report is published by the coordinator to `docs/external-asp-v3/stages/02_HANDOFF.md` on `experimental/v3-design-staging`. Stage02 **execution PASS**, delegated technical acceptance **ACCEPTED**, integration **PENDING**. Any later owner acceptance or integration must be recorded separately with actual verified SHAs; no future publication commit is invented inside this file.
