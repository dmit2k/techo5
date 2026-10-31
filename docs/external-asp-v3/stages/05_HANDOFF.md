# Stage05 — secure session / paired Go↔C wire gate: исполнительский handoff

## 1. Идентичность / контроль статусов

- **Stage:** 05, `stages/05_PAIRED_WIRE_GATE_TASK.md`; 2026-10-31, Europe/Riga; исполнитель: Stage05 code executor на `anfisa-trainer` через Desktop Commander.
- **Docs authority:** `dmit2k/techo5:experimental/v3-design-staging` — **`391b146e35e59170a243a7d594c6a3e3bb281f92`**, локальный и независимый GitHub remote `ls-remote` совпали. Утверждённый SPEC: `SPEC_V2_APPROVED.md` Git blob **`f3e1b00d1f5d184e345c0d2c79da8394435ba3a4`**, независимо `git rev-parse <docs HEAD>:docs/external-asp-v3/SPEC_V2_APPROVED.md`.
- **Принятый вход:** опубликованный `stages/04_HANDOFF.md` §9: independent delegated technical **ACCEPTED** и **integration COMPLETE** по Stage04 (это решение координатора, не дословная декларация владельца). Его точный output pair подтверждён отдельно на обоих remote перед началом.
- **Execution: PASS — final 19/19 Go test/build matrix exit 0; C test, ASan, UBSan and both production ARM32 cross-builds exit 0**. Только software gate; acceptance и integration остаются PENDING.
- **Independent delegated technical acceptance Stage05: ACCEPTED** (2026-10-31; coordinator review §9). Не утверждение о дословной owner acceptance.
- **Integration Stage05: COMPLETE** (2026-10-31). Исполнитель не трогал dev-main; координатор продвинул только Go dev-main и проверил remote pair (§9).
- **Software gate:** compiled native Go/C secure-session PCM-only wire; НЕ полный `mic.Source`/HA, не hardware qualification.
- **Hardware / acoustics / Echo vendor runtime / privileged launcher:** **НЕ ВЫПОЛНЕНЫ**. Никакие Echo, mount, namespace/chroot, deploy, vendor binaries на устройстве не запускались.

## 2. Git-receipt и SHA

| Поле | Go — dmit2k/techo5 | C — dmit2k/frankenasp-v3 |
|---|---|---|
| Integration ref | `experimental/external-asp-v3-dev-main` | `experimental/v3-dev-main` |
| **Accepted T_in / C_in** | `b0c0d5fa96576eff8f7497d48e39fa30a5348e76` | `f6266907edb90e7c9c2f414e8e46e6f215a21b7b` |
| Docs predecessor | `stages/04_HANDOFF.md` @ docs `391b146e35e59170a243a7d594c6a3e3bb281f92` | same |
| Stage ref | `experimental/external-asp-v3-dev-stage05` | none; **detached**, no artificial C commit |
| Isolated worktree | `/home/gptlab/dev/techo5-v3-stage05` | `/home/gptlab/dev/frankenasp-v3-stage05` |
| Stage commits | `6c9af9ca308e5d3cba2baf86e77ebca20afbee00` → `da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9` | none |
| **Tested candidate** | **`da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9`** (19/19 final Go matrix PASS; clean committed tree) | **`f6266907edb90e7c9c2f414e8e46e6f215a21b7b`** unchanged |
| Accepted tested candidate (delegated technical decision) | **da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9** | **f6266907edb90e7c9c2f414e8e46e6f215a21b7b** unchanged |
| **Output integration (coordinator verified)** | **da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9** (Go FF) | **f6266907edb90e7c9c2f414e8e46e6f215a21b7b** unchanged |
| Protected main | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` (unchanged) | `e986ae18643e08e95f29a38aae1dd74762f6362e` (unchanged) |

Exact remote preflight `git ls-remote origin refs/heads/<integration> refs/heads/main`, local HEAD and `git status --porcelain=v1`; both input dev-main worktrees clean, new stage worktrees started clean. No applicable `AGENTS.md` found under repo/ancestor paths. Stage05 Go ref **published and verified** at `da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9` after final 19/19 PASS; remote receipt and UTC/local verification time in §5. No integration/main/docs ref was pushed.

## 3. Фактически проверяемая Go/C пара и software wire

**T_candidate = `da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9`; C_candidate = `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`**. Всё Go только от T_in; C принятого Stage04 без изменений. Historical `dmit2k/FrankenASP` не использован как runtime-сторона: offline compiled target — собственный C04 repo `src/{descriptor,io,runtime,factory_adapter,child}.c` + `tests/mock_vendor.c` **только тестовой линковкой `-DASP_TEST_ONLY`**.

- **Сборка реального process-level mock:** `cd /home/gptlab/dev/frankenasp-v3-stage05 && make all build/io_faults.so`; `build/runtime_mock` SHA256 **`5c19a4ee07d25c6821aea43964a3f46ac966310e6ba5e461f97a3e344fdd244d`**. `build/codec_probe` independent C decoder. Go package `echod/internal/asp` содержит production `session_linux.go`, не Python/fake wire; Go `Start` использует `exec.Cmd` и настоящие OS pipes.
- **Shared independent fixture manifests:** Go `echod/internal/asp/testdata/SHA256SUMS` и C `testdata/SHA256SUMS`, byte-identical SHA256 **`9d26d6cca1e29b10f1abaee583da12ee918ea4252f16f70c3eb2c01531e1d2f4`**. Header ровно 16 bytes, records 28/8, supported descriptor 120 bytes, READY ровно единственный byte **`0xA5`** (fixture `testdata/readiness.bin`), no platform metadata; input MIC4 S32 24 valid bits + REF stereo S16 96k/1920 frames, два физических buffers 30720+7680 bytes. Валидация unsupported-geometry происходит до READY; output точно **640 bytes mono S16LE** на каждый exchange.
- C emits 64 initial output zeros **однократно**; следующие samples из 8-ms ABI quanta 768→128. Direct Go/C compiled test checked complete first output/next output, retained first output unchanged. C writes diagnostic stdout chatter to *stderr*, preserves wire FD: no chatter in Go READY/PCM pipe. C source direct tests also validate REF L/R and mapping, bit alignments, supported/unsupported.
- Real C short/EINTR fault injection via `build/io_faults.so` in test-only child environment, OS pipe reads/writes; test-only separately compiled tiny C helper covers READY invalid, EOF/short output, stuck write, stuck read, delayed partial response, Start failures and cancellation. Test-only environment injection occurs only through unexported Go `startPinned`, no user-facing launch override or public production test API.
- Secure path is fixed **`/data/techo5-linux/external-asp/bin/external-asp`**, root-owned (production), pinned directory-component `openat(O_NOFOLLOW)`, check inode owner/type/mode; strictly known /data and /data/techo5-linux writable ancestor exceptions; `FD3` executable `/proc/self/fd/3`, `FD4` trusted deployment root, stdin/stdout PCM and direct stderr. Production child env `LANG=C, LC_ALL=C` only; no loader overrides. All tests run with *isolated temporary trusted roots*, owner UID set to testing UID; no `/data` mutation or privileged C launcher execution.
- **Startup 10s one-time budget** covers prepare/start/descriptor/READY; after READY stop-hook unregistered; context for session lifetime preserved. **Exchange 5s one shared budget** for all sequential borrowed writes + full 640-byte read, one outstanding via gate; no idle timeout, no async worker, no retry, no extra frame header.
- **API for Stage06** (in internal/asp, Linux): `Start(ctx context.Context, d asp.Descriptor) (*asp.Session,error)`, `(*Session).Exchange(ctx context.Context, buffers [][]byte) ([]byte,error)`, `Cancel()`, `Close() error`. Input slices borrowed until Exchange return (not retained); each successful output is an independently owned fresh 640-byte slice; no pooling/scratch published; `Cancel` stops immediately on blocked pipe; `Close` joins *sole* `cmd.Wait` goroutine, idempotent; terminal errors stop only their own session. Stage06 must still own Selected/Active, identity/short publication lock, final echod shutdown, Rewire, fallback, and publication; **there is no Source/HA activation in Stage05**.
- Hardware qualification not inferred from mock: real Bionic vendor init, original device namespace, factory profiles/calibration, microphone acoustics and runtime loader remain explicitly untested.

## 4. Actual changed files, provenance, LOC

| Repo / file | Action | Contract |
|---|---|---|
| Go `echod/internal/asp/session_linux.go` | **add**, 381 production Go source lines | Pinned FD launch + lifecycle + one 10s READY budget + 5s borrowed PCM exchange + 640-byte owned result, F3/F4, D10–D11/D25–D34/D40–D42/D50–D55/D57–D58 |
| Go `echod/internal/asp/session_linux_test.go` | **add**, 785 test lines | W01–W08, W10 integration/timeout/security/cancellation/owned-output regression |
| C | **no changes**; same C04 sources/commit | C codec/runtime/factory mock and production crossbuild verified without empty stage05 C branch |

`git diff --name-status T_in..T_candidate` lists only those two added Go files, no deletions. Compared to accepted fresh T01 baseline, current net accumulation = prior Stage02/03 changes plus Stage05 (not a claim of Stage05-only LOC). Two Go Stage05 commits: initial implementation then scoped fix so **any failed Exchange (including caller-side buffer validation)** stops own session. Reference-only inspected `dmit2k/techo5:9ef0792a90e1dd0f24ead40b72490ed844d2169e` original `asp_secure_open_linux.go`, `asp_pinned_exec_linux.go` for owner checks/exceptions/pinning; no old code trees/wrappers/installer/OTA imported. C04 already carries exact MIT notices and historical provenance; Stage05 introduces no new third-party/vendor bytes.

Stock MIC, native tap, SoftwareCut, codec and other packages not modified at Stage05. **The new Go session package is exercised in standalone Go test binaries and is not yet wired into `echod`'s production mic.Source activation; the production `echod` ARM32 build checks compatibility but is not evidence of Source calling this code. Stage06 must implement that wiring.** No new CFG2/TLV, metadata, IDME/Q30 in host, statuses/feedback, ASP Debug, preflight, first-PCM state, retry queue, manager, inode cache, platform admission whitelist, vendor assets, OTA or compatibility layer. C source `ASP1_TEST_MAP` conditional is in `#ifdef ASP_TEST_ONLY`, not linked in production. Evidence negative-scope audit `negative-scope-and-loc.txt`.

## 5. Test commands, evidence and results

**Evidence directory (all outside any code/docs worktree):** `/home/gptlab/dev/techo5-v3-stage05-evidence-20261031/`; final exact Go matrix in `final/`. Reproducer script `run-go-matrix-final.sh`; main cwd `/home/gptlab/dev/techo5-v3-stage05/echod`, C cwd `/home/gptlab/dev/frankenasp-v3-stage05`. Toolchains: Go **go1.26.9 linux/amd64**, local `cc` **GCC13.3.0**, ARM32 GCC 13.3.0 SHA256 `73c40a8b836981da0da98dfae66073ed4b9d485cf786899d0c5ef3eeb311e548`, ld 2.42 SHA256 `d15886a5fd7498e8f96a1977ce1b2bbcb2ab617174245d42d026cc2501354bb6`.

Go environment:
```sh
export PATH=/home/gptlab/dev/techo5-asp-p3-4a-arm32-build-20261027/toolchains/go1.26.9/bin:$PATH
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOFLAGS=-p=1 GOMAXPROCS=2
export ASP_STAGE05_C_MOCK=/home/gptlab/dev/frankenasp-v3-stage05/build/runtime_mock
export ASP_STAGE05_C_FAULTS=/home/gptlab/dev/frankenasp-v3-stage05/build/io_faults.so
export ASP_C_DECODER=/home/gptlab/dev/frankenasp-v3-stage05/build/codec_probe
cd /home/gptlab/dev/techo5-v3-stage05/echod
bash /home/gptlab/dev/techo5-v3-stage05-evidence-20261031/run-go-matrix-final.sh
```

| W ID | Test commands / measured evidence | Verdict |
|---|---|---|
| **W01–W02** | `go test -count=1 -parallel 1 ./internal/asp` (native production Go session + compiled C04 `runtime_mock` real pipes; independent fixture, supported READY 0xA5, unsupported pre-READY, first 640 output/padding) ; C `make test` | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W03** | Same tests: C `tests/io_faults.c` short syscall/EINTR, zero progress; EOF/short READY/PCM, broken pipe/child exit; 12 concurrent callers serialize. C `make test` | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W04** | `TestGlobalExchangeTimeout` real blocked read ~5.01s and blocked write ~5.01s; `TestSharedDeadlineAcrossPartialRead` 3s+3s output must fail at 5s; `TestNoIdleTimerAndStartupSurvival` >10s ready child with >5s idle. | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W05** | `TestStartupTimeout` real ~10.08s before READY; `TestStartupCancelAndNoZombie`; `TestNoIdleTimerAndStartupSurvival` successful exchange 10.2s later; `TestProcessFailureAndCancellation/old-session-isolation`. | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W06** | `TestProcessFailureAndCancellation`, `TestStartFailureAndNoFDLeaks` repeated Start failure / success / sole Wait join and `/proc/self/fd` stable; `TestBorrowedInputSizeAndOldOutput` invalid-buffer error terminal; old-session cleanup isolation, idempotent Close. | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W07** | `TestSecurePinnedPathsAndReplacement`, `TestSecureConcurrentRenameRace`, `TestAssetsPinAndReplacementBetweenActivations`, `TestExecutableReinstalledBetweenSessions`, `TestStartFailureAndNoFDLeaks`; permissions/owner/types/symlinks/rename and FD pins; fixed prod env source review. | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W08** | `TestPairedCompiledC` inspects child `/proc/<pid>/fd/3` and `fd/4`, Go OS pipe PCM, C stdout-isolation; `make test` includes offline C launcher alias-policy H02/H08 (no mount). | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W09** | `go test -race -count=1 -parallel 1 ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy`, `go vet ...`, `go test -count=1 -parallel 1 ./...`, default/dot/dot,donut/spot/panel build and compile matrix; ARM32 GOARCH=arm GOARM=7 build; C `make test`, `make asan`, `make ubsan`, C Bionic and static launcher production crossbuild. | **PASS** — measured on final T_candidate/C_candidate and unchanged C04 |
| **W10** | `git diff --name-status T_in..T_candidate`, `git grep`/negative review of new Go session and unchanged C04; no legacy status/feedback or plugin, no new PCM framing, no vendor imports; tests/mocks separated. | PASS source inspection |

**Go final reproducible 19-command runner** `run-go-matrix-final.sh`, results `final/go-results.tsv`, full argv `final/go-commands.txt`, per-command `final/go-*.log` and `final/go-env.txt`; includes:
- `go test -count=1 -parallel 1 ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy`
- `go test -race -count=1 -parallel 1 ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy`
- `go vet ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/lib/alsa`
- `go test -count=1 -parallel 1 ./...`
- `go build -o final/go-builds/echod-linux-amd64 ./cmd/echod`
- For each `dot`, `dot,donut`, `spot`, `panel`: `go test -count=1 -parallel 1 -tags TAG` on focused packages, `go test -run '^$' -parallel 1 -tags TAG ./...`, `go build -tags TAG -o final/go-builds/echod-linux-amd64-NAME ./cmd/echod`
- `go test -c -o final/go-builds/asp-wire-test-amd64 ./internal/asp`
- `env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o final/go-builds/echod-linux-arm32 ./cmd/echod`.

**C exact unchanged C_SHA**: `cd /home/gptlab/dev/frankenasp-v3-stage05 && make test && make asan && make ubsan`, all **exit0**; logs `make-test.log`, `make-asan.log`, `make-ubsan.log`. Crossbuilds with owner-private toolchain from accepted Stage04 (exact commands/scripts, no device action):
```sh
TC=/home/gptlab/dev/techo5-stock-audio-audit-20261021/arm32-tools
BIONIC_CC=$TC/cross/usr/bin/arm-linux-gnueabihf-gcc-13
BIONIC_LD=$TC/binutils-unpacked/usr/bin/arm-linux-gnueabihf-ld
BIONIC_ROOT=/home/gptlab/dev/techo5-stock-audio-audit-20261021/factory-arm32-root
LD_LIBRARY_PATH=$TC/binutils-unpacked/usr/lib/x86_64-linux-gnu
OUT_DIR=/home/gptlab/dev/techo5-v3-stage05-evidence-20261031/c-bionic
bash build_bionic.sh
ARM32_CC=$BIONIC_CC
ARM32_BINUTILS=$TC/binutils-unpacked/usr/arm-linux-gnueabihf/bin
ARM32_SYSROOT=/home/gptlab/dev/techo5-asp2-f2b-minimal-20261028/private/toolchain/sysroot
OUT_DIR=/home/gptlab/dev/techo5-v3-stage05-evidence-20261031/c-launcher
bash build_launcher.sh
```
Both exit0, outputs **byte-identical to accepted Stage04**, ARM EABI5: Bionic `asp1-child` SHA256 **`86d4f4d7827129e944c76187dc4575ef7bc5240f2b99f048e7860ed71057fa05`**, Linux static `asp1-launcher-arm32` SHA256 **`fce1f3e1abc427d7e5402da5c56dd530d5bf2acd7686b4f3aba98d1445b24ab7`**. Final Go native test executable SHA256 **892309a3aaf81e6b077964ee4605b84bddd7147b3f23c28c9907bd8652a999c1**; Linux ARM32 TECHO5 SHA256 **8c3bfc92631976633d09f57069ff07ef960bfe824bbcbb92be66d562bde496b2**; default amd64 SHA256 **74df67ce95309f0d9e1b8ac00d89d6bca08e4e38a2496e255cf625acbb13b549**. All seven final Go build artifacts verified via sha256sum -c final/go-builds/SHA256SUMS; manifest SHA256 **ad7838706398a8f21ee9aa78fd3e4cd68dce92cdf288cfd21c5fbba926a31d06**.

### Final SHA256 / remote receipt

- Tested pair is T **da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9** + C **f6266907edb90e7c9c2f414e8e46e6f215a21b7b**, C unchanged; final/go-results.tsv records **19/19 exit 0**. Every command is recorded in final/go-commands.txt; individual evidence files final/go-*.log.
- Published own Go Stage05 ref verified at **da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9** by independent remote ls-remote after fast-forward stage-only push. Go/C dev-main and both main branches frozen; docs remote HEAD still 391b146e35e59170a243a7d594c6a3e3bb281f92, no docs write. All five local worktrees checked CLEAN. Receipt: final-remote-ref-receipt.txt, time 2026-10-31 17:29:49 EET.
- Compiled offline Go native test ELF: SHA256 **892309a3aaf81e6b077964ee4605b84bddd7147b3f23c28c9907bd8652a999c1**; independently compiled C runtime mock **5c19a4ee07d25c6821aea43964a3f46ac966310e6ba5e461f97a3e344fdd244d**; C codec probe **423084acb8c414d8178b6ebfab39d3fcc40f61eab549f4fafebb585aa249cf3b**.
- Final Linux ARM32 TECHO5 **8c3bfc92631976633d09f57069ff07ef960bfe824bbcbb92be66d562bde496b2**; production ARM32 Bionic child **86d4f4d7827129e944c76187dc4575ef7bc5240f2b99f048e7860ed71057fa05**; static ARM32 launcher **fce1f3e1abc427d7e5402da5c56dd530d5bf2acd7686b4f3aba98d1445b24ab7**. All are compile-only evidence, no Echo runtime.
- Final Go 7-binary manifest: final/go-builds/SHA256SUMS, its own SHA256 **ad7838706398a8f21ee9aa78fd3e4cd68dce92cdf288cfd21c5fbba926a31d06**; verified all seven entries OK.
- Full standalone evidence manifest: EVIDENCE_SHA256SUMS, its own SHA256 **3164f8577de681aa26695718de7af093b90149319e935cee10c0783d308b2472**; **41 individually hashed files** (all 19 Go command logs, C tests/sanitizers/crossbuild logs, Go/C build outputs, Git remote receipt and negative-scope review). See final/go-builds/SHA256SUMS for exact default, dot, dot+donut, spot, panel and ARM32 hashes.
- Final modified source hashes: session_linux.go **da5e43bc177f2fb3403c8b7bebd90560657b336c6c0f38610f0530cea9307f18**, session_linux_test.go **da71a26351742c14b317957321dbdd49daa94bb15aab0768941c735b6ee1292a**; negative-scope-and-loc-final.txt confirms exactly two added Go files (381 production / 785 test lines), no C change and no new forbidden host mechanisms.
- На момент executor handoff acceptance/integration были PENDING; текущие coordinator acceptance/integration receipts — в §9.

## 6. STOP / deviations / residual risks

No architecture contradiction identified in Stage05 scoped PCM-only wire. Real device execution, actual root-owned `/data` tree, vendor factory config/calibration/ABI/DSP and Echo acoustic runtime **unverified**. C mock validates compiled ABI/PCM behavior, NOT proprietary DSP on Echo. Test-only trusted root uses user UID, production enforces root UID0. Go stage05 deliberately does not exercise mic.Source Selected/Active/HA, fallback state or synchronized publication: **these remain Stage06**, not claimed as passing W-matrix source activation. No privileged native launcher invocation, mount, calibration read or Echo action.

For any discovered blocker: cite file/SHA/repro; do not reinterpret architecture. No automatic fallback retries, compatibility status or new integration manager.

## 7. Next stage / coordinator actions

**Next: Stage06** — permitted ONLY AFTER Stage05 software execution PASS, independent technical **ACCEPTED** and **integration COMPLETE**, with Go/C remote integration SHA pair verified and published in `stages/05_HANDOFF.md`. Историческое состояние executor handoff было PENDING; coordinator §9 фиксирует принятый Gate **ACCEPTED + COMPLETE**. Coordinator review, FF-only Go integration, remote SHA verification and this docs handoff publication are completed and recorded in §9. The executor must not invent or write future integration SHA.

Stage06 gets `echod/internal/asp/session_linux.go`, `Descriptor` PCM-only codec, `Start`/`Exchange`/`Cancel`/`Close` lifetime and exact Go+C tested pair, and owns the final Selected/Active/Source identity/publication/echod shutdown wiring. It must not reinterpret software validation as Echo/device qualification. No hardware deployment authorization is implied.

## 8. Publication boundary

Executor evidence and this handoff reside **outside** code/docs checkouts at `/home/gptlab/dev/techo5-v3-stage05-evidence-20261031/05_HANDOFF.md`. Docs coordinator alone publishes `docs/external-asp-v3/stages/05_HANDOFF.md` on documentation branch after independent review/decision/integration receipt; no documentation commits, merges, FFs or pushes by Stage05 executor.

**Final coordinator status:** execution PASS; delegated technical acceptance **ACCEPTED**; integration **COMPLETE**. Docs и Go dev-main опубликовал координатор после тестов; исполнитель их не менял (§9).

## 9. Independent coordinator technical acceptance and integration receipt (2026-10-31)

**Original executor handoff preserved:** `/home/gptlab/dev/techo5-v3-stage05-evidence-20261031/05_HANDOFF.md`, SHA256 `24deaaeba7d2907c6835f0b2336d3504061d8c775ed8546d6c94c58e1886001f`. Original PENDING statuses in the historical executor snapshot were correct at the time. This published coordinator edition includes subsequent independently verified ACCEPTED/COMPLETE receipts.

**Decision: Stage05 software gate ACCEPTED** (independent delegated technical acceptance, not a verbatim owner declaration). Reviewed against `PCM_ONLY_ARCHITECTURE_DECISION.md`, `MASTER_CONTRACT.md`, `SPEC_V2_APPROVED.md`, `stages/05_PAIRED_WIRE_GATE_TASK.md` and accepted Stage04. Go candidate `da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9` is precisely two commits after Stage04 Go input; GitHub independent compare confirms exactly two new files, `session_linux.go` (381 production lines) and `session_linux_test.go` (785 test lines). C `f6266907edb90e7c9c2f414e8e46e6f215a21b7b` remains unchanged. The reviewed Go code uses pinned FD3 executable/FD4 deployment, strict path ownership, one 10-second startup budget through READY, one global 5-second budget across synchronous PCM writes and 640-byte read, per-session Cancel/Close with one Wait/reaper, borrowed PCM and independently owned output. No HA/Source/Selected/Active activation claimed: it belongs to Stage06.

**Independent tests on exactly this pair:**
- `go test -count=1 -parallel 1 -v ./internal/asp` with compiled C04 runtime mock and C fault injector plus standalone compiled codec probe: **PASS**. Real Go/C OS pipes and wire test W01–W08 confirm one byte READY `0xA5`, exact buffers/640 output, C short I/O/EINTR, unsupported before READY, cancellation, 5-second blocked read/write and shared partial-response budget, startup 10-second timeout, survival after >10 seconds from READY, secure pinned path/rename race and installation replacement, FD leaks, one reaper, output ownership.
- `go test -race -count=1 -parallel 1 ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy`: **PASS**; `go vet ./internal/asp ./internal/hardware/mic ./internal/layout ./internal/hardware/privacy ./internal/lib/alsa`: **PASS**.
- Independent full `go test -count=1 -parallel 1 ./...`: **PASS**. Coordinator log at `/home/gptlab/dev/techo5-v3-stage05-evidence-20261031/coordinator-full-go.log`.
- C04 unchanged `make test`, `make asan`, `make ubsan`, codec fixture manifest: **PASS**. Executor's verified final Go 19/19 commands include default, dot, dot+donut, spot, panel and Linux ARM32 builds; coordinator verified their original logs and full 41-entry evidence manifest, but does not claim to independently rerun all tagged/crossbuild variants. Production ARM32 Go, Bionic C child and static launcher are evidenced by hashes/crossbuild logs, not target runtime execution.

**Actual Go-only integration:** on clean worktrees with remote refs verified Go dev-main `b0c0d5fa96576eff8f7497d48e39fa30a5348e76`, Go Stage05 `da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9`, C dev-main `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`, docs predecessor `391b146e35e59170a243a7d594c6a3e3bb281f92`, coordinator executed `git merge --ff-only da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9` in `/home/gptlab/dev/techo5-v3-dev-main` and a direct authorized `git push --porcelain origin HEAD:refs/heads/experimental/external-asp-v3-dev-main` from `anfisa-trainer`. Remote Go stage ref and Go dev-main now both equal the exact accepted candidate; C dev-main unchanged. Protected mains unchanged: Go `62603a7854283a3a07473a17c831c6ef0fd6f8d4`, C `e986ae18643e08e95f29a38aae1dd74762f6362e`. No force/rebase/reset/squash, no device action.

**Accepted Stage06 input:** TECHO5 Go **`da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9`** + FrankenASP V3 C **`f6266907edb90e7c9c2f414e8e46e6f215a21b7b`**. Execution **PASS**, delegated technical acceptance **ACCEPTED**, integration **COMPLETE**. Stage06 production Source/lifecycle work is now unblocked by its separate assignment. No live FrankenASP factory launcher/Calibrator/Echo/acoustic qualification has occurred.
