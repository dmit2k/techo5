# 07 — TECHO5 / FrankenASP V3: финальная программная приёмка — executor handoff

**Для документационного координатора. Фактический вердикт: EXECUTION FAIL / STOP. Отчёт НЕ является опубликованным `stages/07_HANDOFF.md` или решением владельца.**

## 1. Идентичность и статусы

- TASK: `stages/07_SOFTWARE_ACCEPTANCE_TASK.md`, 2026-10-31 (Europe/Riga), исполнение на `anfisa-trainer` через Desktop Commander.
- **Published docs HEAD** `dmit2k/techo5:experimental/v3-design-staging`: **739a744e4f0b617453b88413c41ebdc51a5c0ecf**, local=remote, clean; **actual approved SPEC_V2_APPROVED.md blob** **f3e1b00d1f5d184e345c0d2c79da8394435ba3a4** (`git hash-object`=`git rev-parse HEAD:path`). Docs branch originated from clean T01 via docs commit `06bb7e8a56ffd0406626d4f218f7dc796d799d03`, parent frozen T01; historical first-docs SPEC blob c754947d… is not today's PCM-only authority.
- Прочитаны EXECUTION_BRANCH_STRATEGY, MASTER_CONTRACT, PCM_ONLY_ARCHITECTURE_DECISION, TASK07, SPEC §§20–27, HANDOFF_TEMPLATE, принятый `stages/06_HANDOFF.md` §9, REQUIREMENTS_TRACEABILITY и Stage02-R/03-R/04/05 conformance/commands. Relevant AGENTS.md не обнаружен.
- **Stage06 incoming gate:** Execution PASS / independent delegated technical **ACCEPTED** / Integration **COMPLETE** — coordinator decision, не вымышленная дословная декларация владельца. Remote SHA и clean worktrees подтверждены.
- **Stage07 execution: FAIL / STOP** из-за обязательного S07 full tagged test `-tags dot` (reproduced exact failing test independently). **S08 Android ABI BLOCKED** из-за отсутствия NDK. Software final PASS **не предоставлен**. **Owner acceptance Stage07: PENDING.**
- **Code output integration:** ранее принятая Stage06 пара сохраняется без изменений; Stage07 не имел и не делал code integration, stage refs или commits. Docs publication: PENDING документационного координатора.
- **Hardware tests НЕ ВЫПОЛНЕНЫ; acoustics НЕ ВЫПОЛНЕНЫ; Echo vendor runtime НЕ ВЫПОЛНЕН.** Echo/target binaries/mount/chroot/deploy не запускались.

## 2. Ревизии, Git, worktrees

| Поле | TECHO5 Go | FrankenASP V3 C |
|---|---|---|
| Repo | `dmit2k/techo5` | `dmit2k/frankenasp-v3` |
| Integration ref | `experimental/external-asp-v3-dev-main` | `experimental/v3-dev-main` |
| Input accepted / tested / unchanged output | **cce2f220aeab9835e10a7f7ab659f5ca390011e8** | **f6266907edb90e7c9c2f414e8e46e6f215a21b7b** |
| Source tree SHA | 09f798b6d9515f5de7dc1e7e4d2fd235f268e995 | d79887b4cd9944759748dc83378204d34a7c8bf2 |
| Detached Stage07 worktree | `/home/gptlab/dev/techo5-v3-stage07` | `/home/gptlab/dev/frankenasp-v3-stage07` |
| Stage07 branch / commits | none / none | none / none |
| Protected main HEAD, unchanged | 62603a7854283a3a07473a17c831c6ef0fd6f8d4 | e986ae18643e08e95f29a38aae1dd74762f6362e |
| Final status --porcelain | empty/clean | empty/clean |

- Only accepted remote Go `cce2f220…` used, **not** Stage06 executor local `cc7254208991dbbba002394f9c2c18f6dda35b28` (trees both 09f798b6… but commits differ). C is actual remote integrated C04.
- Remote `git ls-remote` for both dev-main, mains and docs verified pre/post; all existing worktrees checked clean; no historical/doco merge into code lineage. `receipts/FINAL-GIT-PAIR-TOOLCHAIN.txt` and `receipts/lineage-and-files.txt`.
- Previous accepted Stage06 docs handoff published file Git blob **dd47f033f3a9b0fafcf3f0c7e1f55c47e38a4646**.
- A **pre-existing stale** `07_HANDOFF_EXECUTOR.md` from an earlier rejected pair (T 60d176…/C 417891…, docs 778b0…) was preserved byte-identically as `07_HANDOFF_EXECUTOR_STALE_OLD_PAIR_BACKUP.md` (SHA256 0045dec1fc24113f90e91ab60b37ecddb2bc4220c63a6b7c56aec367b6eda919). **Do not use that stale report** for current Stage07; this file replaces it with the verified accepted pair.

## 3. Согласованная Go/C пара, fixtures, ABI

**Pair = T cce2f220aeab9835e10a7f7ab659f5ca390011e8 / C f6266907edb90e7c9c2f414e8e46e6f215a21b7b.**

- Go `echod/internal/asp/testdata/SHA256SUMS` and C `testdata/SHA256SUMS` both SHA256 **9d26d6cca1e29b10f1abaee583da12ee918ea4252f16f70c3eb2c01531e1d2f4**, `cmp` identical and `sha256sum -c` correct cwd exit0; `readiness.bin` SHA256 6922e93e3827642ce4b883c756b31abf80036649d3614bf5fcb3adda43b8ea32 (one-byte READY 0xA5). PCM-only descriptor strictly 16+28×buffers+8×channels, no IDME/platform wire.
- C compiled ELF64 mock runtime `build/runtime_mock` SHA256 **5c19a4ee07d25c6821aea43964a3f46ac966310e6ba5e461f97a3e344fdd244d**; `build/codec_probe` **423084acb8c414d8178b6ebfab39d3fcc40f61eab549f4fafebb585aa249cf3b**; `build/io_faults.so` **4724d98c4cf89fa15454a39ff0161c81b0f55b310a98d49a4bda17363bc9d4d3**; `build/test_alias_policy` **775481c1fb0e4752cf2a4b9f0e0409e9fd819fdb61a70ed20ff61edf6fae4a39**. Go `session_linux.go` is production code under `go test`; C linked with `-DASP_TEST_ONLY` to `tests/mock_vendor.c` (test-only ABI, not proprietary DSP). Independent codec and fault-injector compiled. No saved standalone Go test ELF claimed.
- Toolchains: **go1.26.9 linux/amd64** (`/home/gptlab/dev/techo5-asp-p3-4a-arm32-build-20261027/toolchains/go1.26.9/bin`), native cc **13.3.0**, available ARM GCC cross **13.3.0**. Go runner `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOFLAGS=-p=2 GOMAXPROCS=2`; environment includes `ASP_STAGE05_C_MOCK`, `ASP_STAGE05_C_FAULTS`, `ASP_C_DECODER` pointing only to final C worktree `build` outputs.
- **Fresh Stage07 Linux ARM32 TECHO5 / Bionic C helper / launcher production binary hashes do not exist:** STOP preceded S08 build; prior accepted Stage04–06 crossbuild hashes are historical evidence, not credited as Stage07. NDK Android arm64-v8a/armeabi-v7a builds NOT RUN.

## 4. Реальные изменения / LOC / architecture

**Stage07 changed/added/deleted production files: 0.** Baseline metrics T_final vs T01 and C_final vs C01 are captured in `receipts/metrics.json` and full actual file inventory in `receipts/lineage-and-files.txt`.

| Baseline diff group | Added LOC | Removed LOC | Net | Actual paths |
|---|---:|---:|---:|---|
| Go production (`.go` non-test) | 1193 | 33 | +1160 | 11 |
| Go tests / textual fixture manifests | 2836 | 0 | +2836 | 28 |
| C production `src/`, `launcher/` | 1032 | 0 | +1032 | 16 |
| C build scripts/Makefile | 100 | 0 | +100 | 3 |
| C tests / text fixture manifests | 1076 | 0 | +1076 | 27 |
| C docs | 223 | 1 | +222 | 4 |

Binary fixture bytes are not counted as text LOC. C `.gitignore` +1 other. No whole old ASP files removed from fresh upstream; no false claim of absent legacy-code deletions.
- Core production Go: `echod/internal/asp/{descriptor.go,session_linux.go}`, `internal/hardware/mic/{asp_control.go,native.go,mic.go,athena.go}`, `internal/lib/alsa/alsa.go`, `internal/feature/microphone/microphone.go`, `internal/config/microphone.go`, `internal/boot/boot.go`, `internal/service/group.go`.
- C: `src/{descriptor,io,runtime,factory_adapter,child,vendor_factory,bionic_entry,...}`, `launcher/{launcher,alias_policy}.*`, build scripts, scoped conformance tests.
- Negative review `receipts/source-review.txt` and `receipts/S09-checks.txt`: no new CFG2/TLV, wire statuses/feedback, ASP Debug, preflight, async PCM queue/retry, whitelist, inode cache, synthetic production vendor adapter, compatibility layers, new management framework. Stock service supervisor's existing `StateRetrying` and `Status` **not ASP additions**. No unrelated stock consumers/calls/wake/mono/raw/stock DSP deleted.
- Factory aliases/mount/FD3+FD4 belong only to C launcher, not host PCM/IDME; C test-only mock not linked production. `LAUNCH_PROVENANCE.md` and `THIRD_PARTY_NOTICES.md` preserve MIT old `dmit2k/FrankenASP@52d51de33e3105501b298c1c214db55f05a2eb82` reference provenance, no code-history merge, no installer/OTA/release tools or proprietary runtime assets.
- Minor separate hygiene: `echod/internal/asp/descriptor_test.go` has `gofmt -l` finding in `TestPCMBlocksAreByteExactAcrossCompiledC`. Formatting-only Stage02-R test issue, no production failure; no edit applied.

## 5. Tests / build matrix (exact commands, cwd, exit)

Evidence root `/home/gptlab/dev/techo5-v3-stage07-evidence-20261031`. Runner `run-stage07-software.sh`, exact execution ledger `results.tsv`, log files `logs/<id>.log`, SHA256 `receipts/LOGS-SHA256SUMS`. Go CWD `/home/gptlab/dev/techo5-v3-stage07/echod`, C CWD `/home/gptlab/dev/frankenasp-v3-stage07`, except fixture check CWD `.../frankenasp-v3-stage07/testdata`. All runs use exact pair above.

| Gate | ID / exact executed command | Exit and evidence |
|---|---|---|
| S01, S06 | `c-test`: `make test`; `c-asan`: `make asan`; `c-ubsan`: `make ubsan` | **0/0/0 PASS**. C01–C07: seven exact PCM-only fixtures, 44 malformed/legacy rejections; H01–H08: supported/unsupported DSP admission, independent REF L/R, 64 seed samples, 8-ms quantum, stdout/FD/fault isolation, launcher alias policy and ABI mock. Logs c-test/c-asan/c-ubsan. |
| S01 independent fixtures | `s01-fixtures-corrected`: in C `testdata`, `cmp SHA256SUMS [Go testdata/SHA256SUMS] && sha256sum -c SHA256SUMS` | **0 PASS**. Initial `s01-fixtures` from wrong CWD exit1 due relative paths; corrected, both logs retained; not a source defect. |
| S02 W01–W10 | `s02-compiled-joint`: `go test -count=1 -parallel 1 -v ./internal/asp` | **0 PASS**, compiled exact Go/C OS pipes, 0xA5 READY, 640 bytes/20ms, common 5s, startup 10s and post-READY >10s survival, EOF/partial I/O, secure pinning, sole Wait, owned output. |
| S03 N01–N08/L07–L08 | `s03-native-privacy`: `go test -count=1 -parallel 1 ./internal/hardware/mic ./internal/hardware/privacy ./internal/layout ./internal/lib/alsa` | **0 PASS** default native/SoftwareCut/borrowed/stock/ownership. |
| S04 L03–L06 | `s04-lifecycle`: `go test -race -count=1 -parallel 1 -run '^TestL' ./internal/hardware/mic ./internal/feature/microphone ./internal/service` | **0 PASS**, eight cancellations/stale, Rewire/XRUN/Run-exit, final pre-Wait cancellation. |
| S05 L01/L02/L09 | `s05-selection-ha`: `go test -count=1 -parallel 1 ./internal/feature/microphone ./internal/config ./internal/boot ./internal/service ./internal/component/all` | **0 PASS**, Selected/Active, persistence, helper unavailable, manual-only retry, HA inventory. |
| S07 default | `s07-stock-default`: `go test -count=1 -parallel 1 ./...` | **0 PASS**, full existing stock consumers, phone, detect, wake, voice, history, raw, sensors. |
| **S07 dot** | `s07-stock-dot`: `go test -count=1 -parallel 1 -tags 'dot' ./...` | **1 FAIL**, `TestL07NativeSoftwareCutAndStockRaw`, `bytes=8640 stride=21`; exact log. |
| **S07 focused reproduction** | `go test -count=1 -parallel 1 -tags dot -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic` | **1 FAIL**, same error; `logs/repro-dot-L07.log`, exit receipt. |
| S07 other full tags/race | `dot,donut` run was interrupted when STOP triggered (partial log not credited). Full `spot`, `panel`, `panel,panel_debug` and full relevant concurrent `-race` | **NOT RUN / not credited** due mandated STOP. |
| S08 build/vet | `go vet ./...`; Linux default/dot/dot,donut/spot/panel `go build -tags ... ./cmd/echod`; `CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build ... ./cmd/echod`; C `build_bionic.sh`, `build_launcher.sh` | **NOT RUN** after S07 STOP. No new Stage07 production ARM32 hashes. |
| **S08 Android** | Stock `.github/workflows/build.yml` lines 153–214: NDK **28.2.13676358**, API **26**, `GOOS=android CGO_ENABLED=1 -tags panel`, `GOARCH=arm64 CC=aarch64-linux-android26-clang` and `GOARCH=arm GOARM=7 CC=armv7a-linux-androideabi26-clang` | **BLOCKED: NDK/SDK/target clang absent**, `ANDROID_SDK_ROOT` unset; `receipts/android-ndk-preflight.txt`. Linux `-tags panel` not equivalent. |
| S09 source review | `git diff --name-status/--numstat/--check` both frozen deltas; `git grep` against changed production scope; `gofmt -l`, C provenance/non-import scan | Completed, no forbidden ASP subsystem; test-only gofmt issue above. |
| S10 final receipt | `git show/ls-remote/status`, source metrics, `sha256sum -c`, pair compiled mock identity, traceability review | **PARTIAL**: input/pair/manifests/clean/provenance verified, production target artifact hashes missing because STOP. |

Complete SHA256 checks `receipts/LOGS-SHA256SUMS` (13 files, including partial diagnostic). Key log SHA256: compiled pair `913d0fce7a6b3eec7b17b950676917cd6d6cad4c45bd76de7e977fdcf7f1bb7f`; default full Go `1b6946e94fb9db29df06b2716bfc50ae65500ba5b6ec22720cdb8831587edcb1`; full dot failure `186b0fa791d84196e912572a7fc4543027af6a913176fc6fef885561f891a98b`; focused reproduction `c9258e419c5e9d31414943240e8817541b0f2350a01ec5432bd88d5380cda4f7`.

**Traceability:** `receipts/TRACEABILITY_STAGE07.md` contains 99 individual D/F/V/A rows: D01–D58, F1–F5, V01–V27, A1–A9, mapped to Stage07 S01–S10 and previously accepted C/N/H/W/L tests. **Do not mark all 99 accepted** while S07/S08/S10 incomplete. Hardware A9 separation explicitly retained.

## 6. STOP / defects / minimal fixes

**FAIL-07-01, owner Stage06 (test fixture, not demonstrated production regression):**
- Exact code: `dmit2k/techo5@cce2f220aeab9835e10a7f7ab659f5ca390011e8`, `echod/internal/hardware/mic/asp_control_test.go` `TestL07NativeSoftwareCutAndStockRaw`, lines ~305–320. `stage06Stock()` at line ~75 allocates `FrameSamples * Channels * 3`; `alsaConfig()` in `echod/internal/hardware/mic/native_test.go` at line ~63 sets `Channels: Mics`. Dot constants `Channels=9,Mics=7`: 8640 bytes are incompatible with declared 21-byte frame stride. `plainNative` correctly rejects invalid geometry. This violates required S07 full tagged tests.
- Exact focused command above, exit1, observed `asp_control_test.go:320: cut native capture failed mic: native geometry unavailable: mic: incomplete native capture frame: bytes=8640 stride=21`. Expected test fixture describes actual buffer consistently and verifies both MIC/REF zeros; no skipped/relaxed native validation.
- **Minimal fix proposal, no changes made:** in this test's `fakePlainCapture` use a `cfg` derived from `alsaConfig()` with `cfg.Channels = Channels` (for `stage06Stock()`), preserving MIC+REF mute assertion; alternatively use a MIC-only length based on `Mics` plus separate REF check. Stage06 replacement branch/review, rerun focused `dot`, complete all tagged/race tests and joint S01–S10, independently accept/integrate replacement pair. Do not edit production descriptor/geometry to make a broken test pass.
- No new commits/branches prepared by Stage07.

**BLOCKED-07-02, toolchain:** `.github/workflows/build.yml` requires pinned Android NDK 28.2.13676358/API26 and both Clang targets; neither SDK nor NDK present on `anfisa-trainer`. Minimal fix: provision verified toolchain and re-run actual Android `CGO_ENABLED=1` arm64-v8a and armeabi-v7a Panel ABI builds in verification-only worktrees, hash outputs. No Echo execution.

**Other evidence findings:** initial fixture checksum launched in wrong CWD was corrected PASS (not a code defect); `descriptor_test.go` from Stage02-R lacks gofmt (formatting-only). Stock `service.Group` status/retry are older stock, not V3 status framework. C fixture mock/vendor resources do not prove real hardware qualification. No contradiction justified altering approved PCM-only architecture.

## 7. Передача дальше

- **Final software acceptance FAIL / STOP, owner acceptance PENDING.** Code pair remains Go **cce2f220aeab9835e10a7f7ab659f5ca390011e8** / C **f6266907edb90e7c9c2f414e8e46e6f215a21b7b**, and existing remote dev-main untouched.
- Return failing L07 `dot` fixture to Stage06 owner via separately authorized correction stage, no covert reset/force/squash; provision pinned Android toolchain. Then independently re-approve/integrate only real replacement if any and re-run full software acceptance matrix on exact post-fix accepted Go/C pair.
- **No automatic Stage08, PR, deployment, Echo, vendor runtime or acoustic tests.** Separate owner instruction required.

## 8. Publication by docs coordinator

- This executor report is `/home/gptlab/dev/techo5-v3-stage07-evidence-20261031/07_HANDOFF_EXECUTOR.md`, with evidence `results.tsv`, `logs/`, `receipts/FINAL-GIT-PAIR-TOOLCHAIN.txt`, `receipts/LOGS-SHA256SUMS`, `receipts/metrics.json`, `receipts/TRACEABILITY_STAGE07.md`. Check hashes/real outputs against exact remote refs.
- Coordinator, not executor, may publish **actual FAIL / STOP** in `docs/external-asp-v3/stages/07_HANDOFF.md` on `dmit2k/techo5:experimental/v3-design-staging` after own independent review, via a new docs-only commit and remote postflight. No code integration ref changes.
- Owner decision remains PENDING, docs publication SHA not invented; older 00:13 UTC stale report archived separately and must not be substituted for this report.

## 9. Независимая проверка координатора и распоряжение по дефектам — 2026-10-31

**Итог: Stage07 SOFTWARE FAIL / STOP — техническая приёмка НЕ ПОЛУЧЕНА.** Owner acceptance PENDING. Никаких code FF/merges, новых stage commits или разрешения PR/аппаратных тестов. Исходный handoff исполнителя сохранён без изменений по адресу `/home/gptlab/dev/techo5-v3-stage07-evidence-20261031/07_HANDOFF_EXECUTOR.md`, SHA256 **`da44f0e4fa621fbc1a6f9c5af2244a11cb22b38b63b14415ffe2365b812fb965`**. Старый backup отчёта по отвергнутой паре в расчёт не принимался.

**Проверенные GitHub refs:** Go dev-main `cce2f220aeab9835e10a7f7ab659f5ca390011e8`, FrankenASP V3 C dev-main `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`, docs до данного коммита `739a744e4f0b617453b88413c41ebdc51a5c0ecf`; protected mains не изменены. Все stage07/dev-main/docs worktrees clean. Evidence `EVIDENCE-SHA256SUMS` и `receipts/LOGS-SHA256SUMS` перепроверены.

**FAIL-07-01 (S07):** независимое воспроизведение на фактическом Go SHA командой `cd echod && go test -count=1 -parallel 1 -tags dot -run '^TestL07NativeSoftwareCutAndStockRaw$' ./internal/hardware/mic`: exit **1**, `asp_control_test.go:320: cut native capture failed mic: native geometry unavailable: mic: incomplete native capture frame: bytes=8640 stride=21`. Тест Stage06 создаёт raw по `FrameSamples*Channels*3`, но `alsaConfig()` в `native_test.go` объявляет `Mics` вместо `Channels`; для dot `Channels=9`, `Mics=7`. Production `plainNative` **правильно отвергает** неконсистентную геометрию. Минимальное исправление — Stage06 **test-only fixture**, не ослабление native validation. Диагностические focused прогоны `dot,donut`, `spot`, `panel` PASS, но полная S07 tagged-матрица всё ещё NOT RUN/NOT PASS. Первая диагностическая команда координатора ошибочно запущена не из `echod/` и завершилась `go.mod not found`; она не учитывается, повтор из правильного CWD выполнен.

**BLOCKED-07-02 (S08):** для реальных Android Panel ABI arm64-v8a/armeabi-v7a нужен NDK `28.2.13676358`, API26, `GOOS=android CGO_ENABLED=1`, два Android Clang. На `anfisa-trainer` не найдены `sdkmanager`, NDK/SDK в проверенных стандартных местах и оба требуемых Android clang. Linux `-tags panel` или прошлые ARM32/Bionic binaries не заменяют S08. NDK/SDK не устанавливались, системные настройки не менялись.

**Порядок выхода из STOP:** (1) отдельная Stage06-S07 Go-only correction branch от текущего *опубликованного* Go `cce2f220`, full tagged/race/regression, независимая приёмка и FF только Go после ACCEPTED, публикация `stages/06_S07_CORRECTION_HANDOFF.md`; (2) верифицированный NDK toolchain; (3) новый полный Stage07 S01–S10 retake на новой принятой паре, с реальными Android бинарниками и отдельным `stages/07_RETAKE_HANDOFF.md`. Данный FAIL-handoff оставить исторически неизменным. Задания опубликованы в `06_S07_CORRECTION_PROMPT.md`, `07_SOFTWARE_ACCEPTANCE_RETAKE_PROMPT.md`; публикация не является разрешением автоматического запуска.

**Code pair после этой FAIL-публикации не меняется:** Go **`cce2f220aeab9835e10a7f7ab659f5ca390011e8`**, C **`f6266907edb90e7c9c2f414e8e46e6f215a21b7b`**. Hardware/acoustics/Echo/vendor runtime **NOT RUN**.
