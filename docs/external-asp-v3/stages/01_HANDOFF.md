# Stage 01 — Baseline & Git Bootstrap: handoff

## 1. Идентичность и статусы

- **Stage:** 01 — Baseline & Git Bootstrap; **TASK:** [01_BASELINE_BOOTSTRAP_TASK.md](01_BASELINE_BOOTSTRAP_TASK.md); **дата:** 2026-10-31 (Europe/Riga); **исполнитель:** ChatGPT, Stage01 execution; **публикация:** documentation coordinator по прямому распоряжению владельца после исполнения.
- **Нормативный docs input commit:** 06bb7e8a56ffd0406626d4f218f7dc796d799d03, ветка TECHO5 experimental/v3-design-staging. Фактический Git blob docs/external-asp-v3/SPEC_V2_APPROVED.md: **c754947d81c77860f4221135b84956249f956935**. Устаревшие placeholders и прежние blob SHA не применяются.
- **Owner review docs commit:** f13882667abe58e42b50df1448c1cc4f34a021ea (исправлены ошибочный старый SPEC blob в MASTER_CONTRACT и преждевременно заданные пять байт READY в Stage02/05); утверждение Stage01 включает исправленный документационный контракт. Исторический input docs commit 06bb7e8a остаётся фактом времени выполнения.
- **Execution result:** **PASS** для Stage01 baseline/bootstrap и доступных обязательных offline проверок B01–B05; частичное покрытие B04 и недоступные проверки явно описаны в §5–6. Это **не** означает PASS полного CI, Android NDK, Bionic или реального устройства.
- **Owner acceptance:** **ACCEPTED** — явное решение владельца в чате от 2026-10-31: «ок, попробую еще раз тебе довериться принимаю, идем дальше». Принят именно Stage01 baseline/bootstrap с документированными ограничениями B04 (§5–6), а не полный CI, Android NDK, C/Bionic runtime или аппаратные испытания.
- **Integration:** **COMPLETE (только Stage01 bootstrap refs)**: оба существующих integration refs реально проверены на frozen main SHA, без новых code commits/merge/FF. Дальнейшая integration Stage02 не производилась.
- **Software gate:** baseline; не codec/native/C/paired wire/full Source/final software gate.
- **Hardware tests:** **НЕ ВЫПОЛНЕНЫ**. **Acoustic tests:** **НЕ ВЫПОЛНЕНЫ**. **Vendor runtime on Echo:** **НЕ ВЫПОЛНЕН**.
- **Stage02:** условие owner ACCEPTED для Stage01 выполнено. Переход к Stage02 допускается отдельным исполнительским поручением после повторной проверки remote input pair и текущего docs HEAD. Публикация этого handoff не является автоматическим запуском Stage02.

## 2. Ревизии, provenance, worktrees

| Поле | TECHO5 (Go) | frankenasp-v3 (C) |
|---|---|---|
| Origin URL | https://github.com/dmit2k/techo5.git | https://github.com/dmit2k/frankenasp-v3.git |
| Frozen main, input baseline | **62603a7854283a3a07473a17c831c6ef0fd6f8d4** (T01) | **e986ae18643e08e95f29a38aae1dd74762f6362e** (C01) |
| Integration ref | experimental/external-asp-v3-dev-main | experimental/v3-dev-main |
| Integration input / bootstrap SHA | **62603a7854283a3a07473a17c831c6ef0fd6f8d4** | **e986ae18643e08e95f29a38aae1dd74762f6362e** |
| Previous accepted Stage handoff | Нет: первый stage | Нет: первый stage |
| Stage branch | Не создавалась; detached verification | Не создавалась; detached verification |
| Worktree | ${WORKSPACE}/techo5-v3-stage01 | ${WORKSPACE}/frankenasp-v3-stage01 |
| Stage code commits | **Нет** (административное исключение) | **Нет** |
| Tested candidate SHA | 62603a7854283a3a07473a17c831c6ef0fd6f8d4 | e986ae18643e08e95f29a38aae1dd74762f6362e |
| Owner-accepted candidate SHA | **62603a7854283a3a07473a17c831c6ef0fd6f8d4** (Stage01 baseline accepted) | **e986ae18643e08e95f29a38aae1dd74762f6362e** (Stage01 baseline accepted) |
| Output integration SHA | **62603a7854283a3a07473a17c831c6ef0fd6f8d4** (unchanged) | **e986ae18643e08e95f29a38aae1dd74762f6362e** (unchanged) |
| Remote main и integration verified | Оба == T01; 2026-10-31 | Оба == C01; 2026-10-31 |
| Clean verification worktree | git status --short → пусто | git status --short → пусто |

- **Документационный checkout:** ${WORKSPACE}/techo5-v3-design-staging, ветка experimental/v3-design-staging, исходный docs HEAD 06bb7e8a56ffd0406626d4f218f7dc796d799d03; до публикации чистый. Docs publish commit будет отдельным от этого input HEAD; намеренно не записан как «будущий SHA» внутрь самого файла.
- **Integration checkouts:** ${WORKSPACE}/techo5-v3-dev-main и ${WORKSPACE}/frankenasp-v3-dev-main; их соответствующие HEAD T01/C01 и git status --short пустой.
- **Before/after remote verification:** git ls-remote origin refs/heads/main refs/heads/experimental/external-asp-v3-dev-main (TECHO5) и git ls-remote origin refs/heads/main refs/heads/experimental/v3-dev-main (C) показали одинаковые frozen SHA при повторных проверках. Никакого продвижения, reset, force push, FF, merge или rebase не было.
- **Ancestry:** локальные git merge-base --is-ancestor T01 experimental/external-asp-v3-dev-main, обратная проверка dev-main → T01 и аналогичные C01 ↔ experimental/v3-dev-main завершились exit 0; строго эквивалентные исходные точки.
- **Fork main:** dmit2k/techo5:main=T01; **C main:** dmit2k/frankenasp-v3:main=C01. Локальный TECHO5 refs/heads/main в dev-main checkout отсутствует; проверены actual remote main и origin/main, совпадающие с T01. Это не заменяет remote verification.
- **Upstream read-only snapshot:** HuskerMinion/techo5:main=3be491e4dc274eac0355f41abeb5bac76d420417 на момент Stage01 (10 commits ahead of T01); не использован для FF.
- **Legacy comparison only:** dmit2k/techo5 experimental/asp2-host-v1.6.12-test=9ef0792a90e1dd0f24ead40b72490ed844d2169e; старый FrankenASP=52d51de33e3105501b298c1c214db55f05a2eb82, не source базы или новых commits.
- **Production V3 code patch:** отсутствует на обеих сторонах, delta от T01/C01 нулевой. Новых stage01 code refs не создавалось.

## 3. Go/C pair и границы доказательства

- **Замороженная baseline pair:** **T_SHA=62603a7854283a3a07473a17c831c6ef0fd6f8d4; C_SHA=e986ae18643e08e95f29a38aae1dd74762f6362e**.
- Stage01 **не является** codec conformance, compiled C, helper wire, executable/runtime identity или production device admission; в C01 нет production C sources/compiled helper. Для этого этапа Go tests/builds проверяют только stock code T01, а C01 — чистоту Git и readiness inventory.
- Source/fixture manifest, совместный Go/C test binary, compiled C runtime, production launcher/helper hashes, supported fixtures и DSP binding **не применимы/не создавались** в Stage01. Любой следующий codec/runtime receipt должен иметь собственные новые точные SHA и hashes.
- TECHO5 ARMv7 cross-build без CGO — компиляционная проверка пакетов, **не** ARM32/Bionic device build и не запуск на Echo.
- Old FrankenASP использован только для read-only инвентаризации исторического build substrate; **не** является второй стороной этой пары.
- After bootstrap оба remote integration HEAD совпадают с input pair и остаются unchanged; проверено после offline тестов.

## 4. Stock seam inventory, изменения и сохранность

### 4.1 Подтверждённые текущие стыки

| File / symbol в T01 | Факты и последствия для Stage02+ |
|---|---|
| echod/internal/hardware/mic/mic.go: Source.New/Get, Start, Run, broadcast, Listen, ListenRaw, Close | Один stock capture owner; Source.Run читает 320 samples/20 ms; broadcast выполняет stock SoftwareCut до Decode, затем mixer/AEC/noise/facing/denoise/leveler/history и fanout. Обработка под Source mutex; блокирующий IPC нельзя без анализа вносить в эту критическую секцию. |
| echod/internal/hardware/mic/mic.go: capture.Read/Close, Source.broadcast | Stock raw block после conversion; SoftwareCut очищает stock raw через clear(raw). ListenRaw получает копию уже после обработки, не заменяет native pre-conversion tap. Subscription channels bounded (8); не переносить AST/IPC в subscriber queues. |
| echod/internal/hardware/mic/athena.go: pairedCapture.Read, convert, micBuf/refBuf, putRef | Отдельные native ALSA MIC и REF читаются прежде stock convert; pre-conversion buffers доступны лишь при новом минимальном native tap. Stock putRef формирует обработанный REF (один канал (L+R)/2 и R), он **не байтово идентичен** stereo native L/R. |
| echod/internal/hardware/mic/device_donut.go | Dot3 MIC 4ch S32_LE 16 kHz (24 valid bits high-aligned) + stereo REF S16_LE 16 kHz; pairedCapture → stock 6ch S24_3LE. |
| echod/internal/hardware/mic/device_dot.go | Dot2/Radar: 7 MIC + 2 REF, итог 9ch S24_3LE 16 kHz. |
| echod/internal/hardware/mic/device_cronos.go, device_spot.go | Cronos/Show5: 2 MIC + 2 REF (4ch), Crown/Show8: 4 MIC + 2 REF (6ch), Spot: 4 MIC + 2 REF (6ch), S24_3LE 16 kHz. |
| echod/internal/hardware/mic/device_panel.go, capture_panel.go | Android Panel получает mono S16LE/16 kHz через socket @techo5.mic; REF software-generated from playback; stock output 2ch S24_3LE. Не смешивать с native physical stereo REF. |
| echod/internal/layout/layout.go: readIDME, Idme; layout/athena_mainline.go: parseIdme | readIDME читает /proc/idme/<name> или /proc/device-tree/idme/<name>/value как bytes; Idme/parseIdme преобразуют/обрезают поля в строки. Для approved raw-byte-preserving descriptor Stage02 не использовать нормализованную строку как raw IDME. Mainline Athena читает /dev/mmcblk0boot1; точный field boundary надо исследовать отдельно. |
| echod/internal/hardware/mic/gain.go, echod/internal/service/group.go, echod/internal/boot/boot.go | Rewire закрывает capture, service.Group ждёт Run до Close; lifecycle helper cancellation должен учитывать порядок ожидания и не привязываться к одному capture.Close. XRUN — штатный retry/restart сценарий. |

### 4.2 Native format / geometry границы

- **Stock mono output:** 16 kHz, S16LE, 320 samples/20 ms. Raw stock capture часто 16 kHz S24_3LE, но **не универсальный native input**.
- **Amazon Athena:** 4 MIC S32_LE 96 kHz с 24 valid bits low-aligned; hardware stereo REF S16_LE 96 kHz. **Mainline Athena:** 4 MIC S32_LE 96 kHz с 24 valid bits high-aligned; hardware stereo REF S16_LE 96 kHz. Native за 20 ms: micBuf **30720 bytes**, refBuf **7680 bytes**; далее software conversion to 16 kHz S24_3LE.
- **Donut:** MIC 4ch S32_LE 16 kHz high-aligned 24 valid bits и stereo REF S16_LE 16 kHz; native за 20 ms MIC **5120 bytes**, REF **1280 bytes**.
- **Отсутствие physical REF:** некоторые stock fallback paths используют synthetic/playback REF. External descriptor не должен выдумывать physical REF fields либо переносить host-side REF conversion вместо raw capture. Реальные source availability и IDME metadata требуют дополнительной проверки.
- **Статический анализ не доказывает:** acoustic behavior, calibration semantics/physical MIC-slot mapping, vendor DSP support, device runtime timing, availability реальных /proc/idme fields. Эти решения Stage01 не принимал.

### 4.3 Stock consumers и регрессии для следующих этапов

- Wake/MWW: echod/internal/feature/detect/engine.go (Listen("wake")); STT pre-roll/history: feature/voice/conversation.go, hawake.go, mic/history.go.
- Calls/SIP: feature/phone/call.go (Listen("call") и существующий callCanceller); talkback/intercom: feature/talkback/talkback.go; recording/announce, clap (ListenRaw), noise/spectrum and sensors.
- Configuration/HA: feature/microphone/microphone.go, internal/config/microphone.go; stock WebRTC AEC и mixer/denoise remain stock, не являются old ASP.
- Реальный lifecycle/cancellation, subscriber frame ownership, privacy SoftwareCut, HA Selected/Active, stock call paths должны покрываться отдельными tests назначенных этапов; baseline PASS этих будущих тестов не означает.

### 4.4 Файлы, provenance, LOC

| Repo / file | Added/changed/deleted в кодовом Stage01 | Назначение |
|---|---|---|
| TECHO5 T01 production | **0 / 0 / 0** | frozen stock baseline |
| frankenasp-v3 C01 production | **0 / 0 / 0** | clean C initial |
| TECHO5 docs: docs/external-asp-v3/stages/01_HANDOFF.md | **added, docs-only**, отдельным docs commit | Публикация настоящего отчёта, не code delta |

- **Метод LOC:** git ls-files -z, production files с расширением .go без суффикса _test.go, physical newline count; test/docs отдельно. TECHO5 whole repo: **1158 production Go files / 222968 LOC**; tests **900 / 119672 LOC**. echod: **1142 / 218367**, tests **891 / 118512**.
- echod/internal/hardware/mic: **21 / 4016** production, tests **17 / 2305**; privacy **10 / 632** (tests 3 / 158); layout **14 / 1080** (tests 8 / 480); boot **3 / 267** (tests 1 / 72); service **2 / 425** (tests 1 / 417).
- C01: только LICENSE (1063 bytes) и README.md (15 bytes); **0 C source files, 0 production C LOC**.
- **Production LOC added/removed/net = 0/0/0** на обеих code сторонах. Старый experimental TECHO5 содержал 34 legacy asp* files под mic/hardware, отсутствующих в T01: это **не** Stage01 удаления. Наличие stock echod/internal/lib/asp/asp.go относится к speaker tuning, а не новой External ASP V3.
- **Неперенесённые legacy:** CFG2/TLV, ASP statuses/feedback/Debug, old ASP wire, preflight, transient managers/queues/retries, platform whitelist/inode cache; никакого bulk copy из старых репозиториев. Ни installer/OTA/release automation, ни proprietary vendor assets/runtime не добавлялись. Reference code не копировался; license/provenance reuse **не применялся** в Stage01.

## 5. Offline tests, builds, compiler readiness и evidence

**Хост:** validation host. Go toolchain ${HOME}/.local/toolchains/go/bin/go, **go1.26.0** (не находится в default PATH). GCC Ubuntu **13.3.0-6ubuntu2~24.04.1**, Python **3.12.3**; make, pkg-config, readelf обнаружены. Нормативный echod/go.mod — go 1.26.0. ARM32 cross-GCC, Android NDK, Bionic vendor toolchain/libasp **недоступны** на этом хосте, не скачивались.

**Общее окружение всех успешных offline Go команд:**

~~~sh
cd ${WORKSPACE}/techo5-v3-stage01/echod
export PATH=${HOME}/.local/toolchains/go/bin:$PATH
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMAXPROCS=2 GOFLAGS=-p=1
~~~

| Check | Команда / target | Pair | Result | Evidence on validation host |
|---|---|---|---|---|
| B01 | git ls-remote; git rev-parse; git merge-base --is-ancestor в обе стороны; git status --short | T01/C01 | **PASS**, exact refs, clean | §2; повторная remote проверка |
| B02 | git show/diff; static source grep/inspection текущих mic/layout/service/consumers | T01/C01 | **PASS** stock inventory | §4 |
| B03 | go test ./internal/hardware/mic ./internal/hardware/privacy ./internal/config ./internal/feature/microphone ./internal/service ./internal/boot | T01/C01 | **PASS exit 0**; feature/microphone [no test files] | /tmp/techo5-v3-stage01-B03.log |
| B04 focused | Для каждого tags: default, dot, dot,donut, spot, panel — go test -tags "$tags" <6 B03 packages> (default без -tags) | T01/C01 | **5/5 PASS exit 0** | /tmp/techo5-v3-stage01-focused-tests.log |
| B04 build | Для пяти перечисленных tags: go build -tags "$tags" ./... на native linux/amd64 | T01/C01 | **5/5 PASS exit 0** | /tmp/techo5-v3-stage01-build-matrix.log |
| B04 ARMv7 | GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -tags "$tags" ./... для тех же пяти tags | T01/C01 | **5/5 PASS exit 0** | /tmp/techo5-v3-stage01-build-matrix.log |
| B04 vet | GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go vet -tags dot,donut ./... | T01/C01 | **PASS exit 0** | /tmp/techo5-v3-stage01-vet-donut.log (empty successful log) |
| B04 Panel debug | go test -tags panel,panel_debug ./internal/lib/loopback | T01/C01 | **PASS exit 0** | /tmp/techo5-v3-stage01-panel-debug.log |
| B04 optional full CI | go test ./... (default) | T01/C01 | **INCOMPLETE / timeout exit 124 at 240s**; завершённые пакеты без observed FAIL; не считать PASS | /tmp/techo5-v3-stage01-go-test-default.log |
| B04 optional tag full CI | go test -tags dot ./... | T01/C01 | **INTERRUPTED** после budget; прочие full-tag suites не запускались | /tmp/techo5-v3-stage01-go-test-dot.log |
| B04 Android | Android Panel NDK/API26 build | T01/C01 | **BLOCKED / NOT RUN** — NDK отсутствует | toolchain inventory |
| B05 C | Native ARM32 Linux/Bionic compiled runtime | T01/C01 | **BLOCKED / NOT APPLICABLE** — в C01 нет C implementation, cross/Bionic prerequisites отсутствуют | C01 tree/inventory |

**Лог SHA256:**

| Path | SHA256 |
|---|---|
| /tmp/techo5-v3-stage01-B03.log | ec698df80b036bc72a651e1c8a689d7a9744aebaaf991d3c5965773b83caa311 |
| /tmp/techo5-v3-stage01-focused-tests.log | ccafb9555ddcb56b9d1762311fa9b01aba786d1b88944e713a4b2daf0c3ad082 |
| /tmp/techo5-v3-stage01-build-matrix.log | ef3dc46a3815714fc074f280e566276bb1c4d36e7e2b1b96c87e59cc674d35f7 |
| /tmp/techo5-v3-stage01-tests-matrix.log | 6cd287493564a99a64cf0b63b740de1ed62370fbba95b7ed13e186c61b667066 |
| /tmp/techo5-v3-stage01-go-test-default.log | c63c2f3297910b0e4ad9cded0be609c5d252e2d75439dcdc974d5a174e1d7a90 |
| /tmp/techo5-v3-stage01-vet-donut.log | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |

**Диагностика baseline execution:** первый запуск без явного Go PATH завершился exit 127; с найденным go1.26.0 первоначальный параллельный прогон наткнулся на host process budget/resource temporarily unavailable/newosproc; повтор с GOMAXPROCS=2 и GOFLAGS=-p=1 успешно прошёл B03 и матрицу. Это ограничения host execution, а не доказанные дефекты stock source. Full-suite 240s timeout не является PASS и не трактуется как observed test regression. Не запускались race tests, vendor/ABI/hardware/acoustic tests. Тестовые target ARM binaries на устройстве не исполнялись.

**CI inventory:** .github/workflows/check.yml содержит stock build defaults/dot/dot,donut/spot/panel и ARMv7/gofmt; .github/workflows/build.yml содержит дополнительные full test-tag suites, ARM vet и Android Panel NDK 28.2.13676358 (API26; workflow Go 1.26.8). Использован host Go 1.26.0, не выдаётся за идентичную CI-конфигурацию. Исторический dmit2k/FrankenASP build_bionic.sh/BIONIC_FACTORY.md просмотрены только как reference: для настоящего C Stage потребуются совместимые ARMv7 toolchain/ld, Bionic runtime/linker и отдельно получаемые владельцем vendor libs/calibration, без включения в Git.

## 6. Отклонения, blockers, risks и STOP

1. **Нет полного B04 CI proof:** T01, echod, go test ./...; 240s timeout exit 124; следующие full-tag test suites не завершены. **Expected:** все выбранные offline CI suites exit 0. **Минимальный follow-up:** отдельный ресурсно ограниченный/full CI job с достаточным timeout, зафиксировать версии Go и artifacts; назначить до финального software acceptance, не подменять B03 PASS.
2. **Android NDK отсутствует:** T01, .github/workflows/build.yml, Android Panel API26; host не имеет NDK. **Expected:** соответствующая compile matrix при наличии toolchain. **Follow-up:** предоставить версионированный NDK, выполнить compile в отдельной build/CI задаче, не запускать на Echo.
3. **C ARM32/Bionic prerequisites отсутствуют:** C01 содержит только LICENSE/README и не содержит C helper; не существует helper для тестирования, cross compiler/vendor runtime недоступны. **Follow-up:** Stage03/04 по dependency plan подготавливают минимальный C runtime и отдельные разрешённые toolchains/assets, затем собственные tests. Stage01 не мог дать C build PASS.
4. **Raw IDME/REF seam требует implementation evidence:** T01, mic/athena.go:pairedCapture.convert/putRef и layout/layout.go:Idme / athena_mainline.go:parseIdme. Применение stock normalized REF или trimmed IDME к raw descriptor нарушит MASTER_CONTRACT §3 / approved V2. **Follow-up:** Stage02 строго задаёт byte-exact native tap/descriptor fixture semantics на actual sources без новых platform metadata и домыслов о Q30.
5. **Lifecycle/privacy риск следующих этапов:** T01, mic/mic.go:Source.broadcast / service/group.go; blocking IPC под stock mutex, повторный SoftwareCut, reaper Wait races и publish scratch buffer недопустимы. **Follow-up:** специализированные tests на Stage05–06 для cancellation, XRUN/Rewire, ownership, stock consumers, secure launch, HA и fallback.
6. **Vendor/device claims:** физические MIC-slot mapping, native IDME на конкретном Echo, заводская калибровка, helper DSP backend, acoustic performance, реальные latency/mute semantics **не проверялись**. Software mocks/компиляция не дают права утверждать hardware PASS.

**STOP на Git divergence/dirty refs:** не наступил: обе code main/dev-main пары совпали, нет новых production изменений. **Нарушение архитектуры Stage01:** не выявлено. Пункты 1–6 остаются документированными ограничениями, не скрытыми PASS.

## 7. Передача Stage02 и owner decision

- **Следующий по плану:** Stage02 — Descriptor Codec; owner acceptance Stage01 получен 2026-10-31, перед стартом Stage02 необходимо повторно проверить оба remote integration refs и обновлённый docs HEAD.
- **Exact input pair:** TECHO5 experimental/external-asp-v3-dev-main **62603a7854283a3a07473a17c831c6ef0fd6f8d4**; C experimental/v3-dev-main **e986ae18643e08e95f29a38aae1dd74762f6362e**.
- Stage02 получает stock mic/pairedCapture/native conversion, layout raw IDME источники из §4, documented toolchain/env из §5, и необходимость сохранять byte-exact buffers/metadata по SPEC V2 и MASTER_CONTRACT. Ни один новый Stage02 API не считается уже созданным.
- **Owner acceptance: ACCEPTED (2026-10-31)**. Решение владельца в чате: «ок, попробую еще раз тебе довериться принимаю, идем дальше». Принята frozen pair из §3, ограничения B04 и будущие toolchain/CI проверки из §6 сохраняются. **Bootstrap integration remote status: COMPLETE** только для frozen initial refs. Следующая реализация Stage02 не осуществлялась при фиксации решения.
- Аппаратное тестирование, Echo deploy, vendor runtime launch, acoustics **не разрешены автоматически**.

## 8. Публикация координатором

- Этот файл опубликован исключительно в TECHO5 **experimental/v3-design-staging** как отдельный documentation-only commit. SHA самого публикационного коммита не вставляется в его содержимое, чтобы не создавать фиктивную самоссылку.
- Отчёт не изменяет T01, C01, code dev-main refs, main или legacy references, не интегрирует docs в production.
- Сверка после push: git ls-remote origin refs/heads/experimental/v3-design-staging; git show HEAD:docs/external-asp-v3/stages/01_HANDOFF.md; git status --short; повторить независимую remote verification для обоих code repos. Фактический docs publish SHA и результат remote проверки сообщаются отдельно.
