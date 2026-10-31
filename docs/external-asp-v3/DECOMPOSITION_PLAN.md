# План исполнения чистой External ASP V3

**Пакет принят владельцем 2026-10-31; Stage01 принят, Stage02–07 требуют отдельных поручений и stage acceptance/integration.** Архитектура согласована. **7 этапов: bootstrap, 5 этапов реализации, программная приёмка**. Отдельная аппаратная/акустическая проверка не входит в эти полномочия.

## 1. Проверенный baseline на 2026-10-30, Europe/Riga

Проверка выполнена через Desktop Commander на anfisa-trainer, без изменения code refs, сборок и выполнения тестов реализации.

| Объект | Проверенная ревизия | Назначение |
|---|---|---|
| Docs `experimental/v3-design-staging` | `0791797065a152c39e0eb401d4e35a8a2f276964` | Исходный HEAD этой документационной работы; worktree и remote совпали |
| APPROVED V2 blob | новый blob после публикации исправленного комплекта | Проверен git rev-parse, неизменяемый |
| Fork TECHO5 main | `62603a7854283a3a07473a17c831c6ef0fd6f8d4` | Frozen Go main, no FF |
| Исторический HuskerMinion/techo5 main | `5cb2afbd9eabc061e380f2b523ab2739e2c4747d` | Read-only снимок; **не будущий start SHA по умолчанию** |
| Старый Go reference | `9ef0792a90e1dd0f24ead40b72490ed844d2169e` | Только чтение |
| Старый FrankenASP reference | `52d51de33e3105501b298c1c214db55f05a2eb82` | Только чтение, независимо от текущего HEAD старого checkout |
| Новый frankenasp-v3 main | `e986ae18643e08e95f29a38aae1dd74762f6362e` | Подтверждены только README.md и LICENSE |

Remote HEAD получены через git ls-remote. Объекты upstream и начального C загружены для git show с `--no-tags --no-write-fetch-head`, без обновления веток. `merge-base --is-ancestor` fork main → upstream успешен; `rev-list --left-right --count` = **0 / 47**. В стратегии исторически указаны c88d6e3c… и 43 коммита; теперь сверху ещё четыре: 19825e93, 9d7094fe, 560970fc, 5cb2afbd. Это основание повторить проверку в 01, не разрешение выполнить FF сейчас.

Старый Go reference и upstream имеют merge-base 90b23260…. При сравнении этого common base с текущим upstream stock mic/boot/service files не изменились; изменились соседние consumers, включая calls и wake front end. Прямой reference→upstream diff содержит отсутствие экспериментальных ASP-файлов; это **не список предстоящих удалений**.

## 2. Подтверждённые точки интеграции

Пути ниже существуют в upstream 5cb2afbd…. Номера строк — ориентиры снимка, исполнитель ищет актуальные symbols после bootstrap.

| Реальный путь / symbol | Наблюдение | Владелец работы |
|---|---|---|
| `echod/internal/hardware/mic/mic.go`: Source, Run (~321), broadcast (~383), Close (~497) | Один capture owner; Run читает, overrun делает continue; broadcast держит s.mu, выполняет stock DSP, копирует mono, затем remember/spectrum/listeners; raw отдельно копируется | 03 native/mute; 06 IPC вне mutex, выбор и stable External output |
| `echod/internal/hardware/mic/athena.go`: pairedCapture.Read (~515), convert (~594) | Native micBuf/refBuf читаются до convert; старого retainNative/aspConfig нет. REF может отсутствовать; stock softFill/soft reference остаётся stock | 03 |
| `echod/internal/hardware/mic/capture_alsa.go`: openPlain; `mic.go`: captureDevice | Plain capture уже выдаёт interleaved S24_3LE; не нужна DSP-specific обёртка | 03 |
| `echod/internal/hardware/mic/capture_panel.go`: socketCapture.Read | Android app MIC S16 + software reference → stock S24 layout; сохранять текущий Panel compile/behavior, не создавать отдельную External platform subsystem | 03 regression, 07 builds |
| `echod/internal/layout/layout.go`: Idme/readIDME; `layout/device_cronos.go` | Существуют hardware identity/IDME sources; готовых ASP PlatformInfo/MicFactoryGain в upstream нет | 03 native PCM geometry only; stock IDME sources не входят в External |
| `echod/internal/hardware/mic/gain.go`: Rewire | Close capture → service.ErrRestart/reopen. Это не завершение External | 06 |
| `echod/internal/service/group.go`: Group.stop (~112) | wg.Wait до service.Close; cleanup только после group.Run слишком поздний | 06: cancel/close pipes до ожидания |
| `echod/internal/boot/boot.go`: Run/run; `cli/echod/run.go` | boot.run владеет долгоживущим inner context и update-triggered cancellation; signal идёт из CLI. Прямой минимальный pre-wait hook здесь или привязка к этому lifetime, не новый supervisor | 06, проверка signal/update/error paths |
| `echod/internal/config/microphone.go`, `feature/microphone/microphone.go`: build/Restore/Entities | ASP config/entities отсутствуют; добавить только Selected/Active и immediate persistence | 06 |
| `echod/internal/feature/phone/call.go`: callCanceller; `config/microphone.go`: CallCancelEngine | Новое относительно reference переключение stock AEC на время нескольких calls и восстановление актуальной настройки | 06/07 stock regression |
| `feature/detect/backends.go`, `feature/wakeword/wakeword.go`, `cli/echod/wake.go` | Upstream изменил выбор front end для конкретной wake model; не возвращать старые вызовы из reference | 07; не менять для ASP |
| `.github/workflows/check.yml`, `build.yml`, `echod/go.mod` | Реальная матрица tags: default, dot, dot,donut, spot, panel; Linux ARMv7; отдельный Android NDK panel build двух ABI | 01 фиксирует команды, 07 выполняет |

Новый пакет descriptor Go и новые C runtime filenames ещё не существуют. TASK задают **компоненты**, а не выдают предполагаемые имена за имеющиеся. Размещение компактного Go codec в новом `echod/internal/asp/` допустимо; содержание создаётся заново. C executor выбирает минимальную структуру и фиксирует реальные пути в handoff 04.

### Reference, разрешённый к чтению

- Go 9ef0792a…: `asp_secure_open_linux.go`, `asp_pinned_exec_linux.go` — trust boundaries/FD pinning; `asp.go`, `asp_control.go` — cancellation/identity и единственный reaper, **без** старых timers/state/wire.
- Go reference cfg2_metadata.go, asp_platform*, asp_factory_gain* — исторические примеры прежнего ошибочного metadata transport: **не переносить**. PCM-only host не читает IDME для External; FrankenASP launcher отдельно готовит vendor resources и IDME aliases. Не переносить CFG2 validator или Athena admission.
- C 52d51de3…: `tools/external-asp/core.c` — numeric mapping и 768→128 bridge; `factory_adapter.c/.h`, `factory.c` — vendor ABI/lifecycle; `factory_child.c` — изоляция stdout; `cfg2_runtime.c` — граница compatibility, не старый parser.
- C: `launcher/launcher.c`, `launcher/README.md`, `build_bionic.sh`, `BIONIC_FACTORY.md` — только необходимая native-launch/Bionic build семантика. Старый launcher был source/cross-build qualified, **не hardware-qualified**. Не переносить его прежние budgets и утверждения о готовности.
- Лицензия старого C — MIT. При любом минимальном reuse сохранить applicable notices и записать origin repo/SHA/path/фрагмент. Никаких vendor binaries/config assets в Git.

[Upstream snapshot](https://github.com/HuskerMinion/techo5/tree/5cb2afbd9eabc061e380f2b523ab2739e2c4747d/echod) · [Go reference](https://github.com/dmit2k/techo5/tree/9ef0792a90e1dd0f24ead40b72490ed844d2169e/echod/internal/hardware/mic) · [C reference](https://github.com/dmit2k/FrankenASP/tree/52d51de33e3105501b298c1c214db55f05a2eb82/tools/external-asp) · [New C baseline](https://github.com/dmit2k/frankenasp-v3/tree/e986ae18643e08e95f29a38aae1dd74762f6362e).

## 3. Этапы и границы

| ID / TASK | Исходные R | Новое в TECHO5 | Новое в frankenasp-v3 | Проверяемый результат | Сложность / контекст |
|---|---|---|---|---|---|
| [01 BASELINE_BOOTSTRAP](stages/01_BASELINE_BOOTSTRAP_TASK.md) | R1 + Git strategy | Только будущий отдельно разрешённый frozen baseline/bootstrap, stock baseline | Проверка независимой чистой базы/build prerequisites | Принятые T01/C01, stock seam/build inventory | Средняя / средний |
| [02 DESCRIPTOR_CODEC](stages/02_DESCRIPTOR_CODEC_TASK.md) | R2 | Новый descriptor writer и test decoding | Новый descriptor reader и независимые fixtures | Byte-exact Go/C conformance; runtime ещё не подключён | Средняя / средний |
| [03 NATIVE_PRIVACY](stages/03_NATIVE_PRIVACY_TASK.md) | R4, части R1/R8 | Native tap, **no External metadata assembly**, единый early SoftwareCut, stock ownership tests | Не требуется | Реальные native buffers/descriptor без DSP admission; Built-in сохранён | Высокая / средний |
| [04 HELPER_RUNTIME](stages/04_HELPER_RUNTIME_TASK.md) | C часть R3, R8 | Не требуется | Минимальный helper/DSP bridge/FIFO, wire loop, launch substrate, native tests и ARM32 build | C component PASS, 64-sample output padding, READY/fixed PCM; **ещё не Go/C gate** | Высокая / высокий |
| [05 PAIRED_WIRE_GATE](stages/05_PAIRED_WIRE_GATE_TASK.md) | Go часть R3, session части R5, secure launch | Secure launch + session IPC, deadlines, cancel, единственный Wait | Использование принятого 04; только необходимые corrections/tests | **Первый compiled Go/C wire PASS по паре T05/C05**; production Source ещё Built-in | Высокая / средний |
| [06 SOURCE_CONTROL_LIFECYCLE](stages/06_SOURCE_CONTROL_LIFECYCLE_TASK.md) | R4 output, R5, R6–R7 | Подключение Source, stable output, Selected/Active+HA, Rewire/XRUN, final cleanup | Обычно без изменений; пара проверяется повторно | Полный Source→helper→publication и управление, без промежуточной state machine | Высокая / высокий |
| [07 SOFTWARE_ACCEPTANCE](stages/07_SOFTWARE_ACCEPTANCE_TASK.md) | R8–R9 | Верификация принятой реализации | Верификация принятого helper | Финальная программная приёмка точной пары; отчёт, LOC, builds | Средняя / средний |

Почему не R1–R9: bootstrap необходим для новой родословной; codec можно проверить отдельно; native/privacy строится **до** включения runtime, чтобы первая активация не обходила mute; C ABI/build scope достаточно самостоятелен для отдельной сессии; 05 проверяет реальный IPC без временной пользовательской интеграции; 06 сразу строит полный минимальный control/lifecycle. Отдельный этап удаления diagnostics не нужен — они отсутствуют. Финальные R8/R9 объединены, чтобы не повторять одну и ту же общую приёмку.

## 4. Матрица зависимостей

Здесь «технические входы» — нужные артефакты. Исполнительский порядок **01→02→03→04→05→06→07**, поэтому все starts используют последний принятый handoff; более ранний технический dependency не разрешает старт от устаревшего integration HEAD.

| Этап | Технические входы | Обязательный предыдущий accepted handoff | Выход для следующего |
|---|---|---|---|
| 01 | Утверждение bootstrap/baseline отдельно | Нет; remote snapshot + strategy | T01=frozen-base dev-main, C01=initial и отдельный C dev-main от initial |
| 02 | 01 | 01 | T02/C02 codecs и fixtures |
| 03 | 01, 02 | 02 | T03 native/privacy; C03=C02 |
| 04 | 01, 02 | 03 | T04=T03; C04 helper/build/test commands |
| 05 | 02, 03, 04 | 04 | T05/C05 wire pair и совместный receipt |
| 06 | 03, 05 | 05 | T06/C06 полная Source/control/lifecycle pair |
| 07 | 02–06 | 06 | Та же финальная пара + software acceptance |

03 и 04 технически могут разрабатываться независимо после 02 в разных repositories, но **базовый план последовательный**. Параллелизм не разрешён этим пакетом: потребовал бы отдельного owner-approved плана зафиксированных bases и интеграции. Никакой параллельной записи в dev-main/main. Независимые read-only проверки внутри одного этапа допустимы.

## 5. Revisions, branches, worktrees

Для NN>01 открыть `stages/(NN-1)_HANDOFF.md` из опубликованной docs branch; проверить execution PASS, acceptance ACCEPTED с решением владельца и integration COMPLETE. Входы T_in/C_in = **output integration HEAD** этого handoff. Сверить локальные и remote integration refs на точное равенство; дополнительно проверить ancestry stage commits и исходного T01/C01. Если HEAD продвинулся — не использовать непроверенную вершину, получить актуализированный принятый handoff.

| ID | TECHO5 рабочее дерево / branch | C рабочее дерево / branch |
|---|---|---|
| 01 | `/home/gptlab/dev/techo5-v3-stage01` detached для baseline; будущая integration `/home/gptlab/dev/techo5-v3-dev-main` | `/home/gptlab/dev/frankenasp-v3-stage01` detached; репозиторий/основной checkout `/home/gptlab/dev/frankenasp-v3` |
| 02 | `/home/gptlab/dev/techo5-v3-stage02`; `experimental/external-asp-v3-dev-stage02` | `/home/gptlab/dev/frankenasp-v3-stage02`; `stage02` |
| 03 | `/home/gptlab/dev/techo5-v3-stage03`; `experimental/external-asp-v3-dev-stage03` | Нет пустой ветки; C_in только pin |
| 04 | Нет пустой ветки; T_in только pin | `/home/gptlab/dev/frankenasp-v3-stage04`; `stage04` |
| 05 | `/home/gptlab/dev/techo5-v3-stage05`; `experimental/external-asp-v3-dev-stage05` | Detached C_in для tests; `stage05` в отдельном C stage05 worktree только если меняется C code/tests |
| 06 | `/home/gptlab/dev/techo5-v3-stage06`; `experimental/external-asp-v3-dev-stage06` | Detached C_in; `stage06` только при доказанном scoped изменении IPC/tests, с новой приёмкой пары |
| 07 | `/home/gptlab/dev/techo5-v3-stage07` detached T06 | `/home/gptlab/dev/frankenasp-v3-stage07` detached C06 |

Если нужный путь/branch занят, STOP и сверить владельца; не overwrite/clean/reset. 01 и 07 не создают пустые stage01/stage07 refs. Это обоснованные административные этапы; code branches следуют общей нумерации, когда есть code changes.

Stage executor коммитит только свою меняющуюся code branch. Push stage refs допускается лишь в рамках отдельного поручения; не пушить integration/main. После owner acceptance integration agent отдельно выполняет разрешённое продвижение refs. Reports и решения идут отдельными documentation commits в design-staging, без merge документации в code. Если 07 находит дефект, вернуть его ответственному этапу, принять corrected pair и повторить затронутые проверки; 07 не чинит production незаметно.

## 6. Первая Go/C интеграция — gate 05

02 — conformance bytes, 04 — C component tests; ни один из них не является первой production wire-парой.

05 компилирует **новый Go host session code и новый C child из frankenasp-v3**, использует реальные OS pipes и production codec/wire/FIFO. Test-only DSP binding обеспечивает offline исполнение без vendor assets; не подменяет C runtime Python-скриптом и не вводит fake DSP в production. Отдельно собираются production TECHO5 ARM32 и C helper/необходимый launcher ARM32; отсутствие toolchain не разрешает заменить сборку моками.

Вход native/descriptor берётся из принятого 03 либо его exact fixtures; whole input write → 640-byte read, READY, partial I/O, timeout, EOF, stdout isolation и first output проверяются совместно. Полная Source activation начинается только в 06: 05 честно фиксирует это ограничение и не объявляет HA/lifecycle Rewire завершёнными.

Receipt пары:
- T_candidate/C_candidate — полные 40-символьные commit IDs (unchanged C = точный C04);
- обе сборки, toolchain identity, native/target binary SHA256 и fixture manifest digest;
- команды, exit codes, результаты conformance и compiled integration tests на **этих** commits;
- область mock/vendor и аппаратные ограничения явно разделены;
- execution PASS/FAIL; owner acceptance пары; затем фактические integration T_out/C_out и проверка обоих remote refs.

Нельзя принять «новый Go + старый FrankenASP», одиночный compile или один успешный fixture за gate. Не требуется двусторонний change ради пары: неизменённый проверенный C SHA полноценный член пары. При C correction сделать stage05 и перепроверить оба commit IDs. При обновлении одной стороны pair receipt устаревает до повторных relevant tests.

## 7. Software acceptance и отложенное hardware

07 проверяет §20–24 на итоговой T06/C06 (либо принятой corrected pair). Требуются conformance, compiled integration, race/security/native/privacy/retention/lifecycle/control/stock и все supported builds. Результаты предыдущих этапов можно переиспользовать только с доказанной применимостью к финальным SHA; финальный joint run обязателен.

Vendor runtime на реальном Echo, mounts/exec на устройстве, реальная калибровка/clock alignment и акустика: **НЕ ВЫПОЛНЕНЫ**. Это отдельное будущее owner instruction, не скрытый stage08 и не причина ложного software FAIL. Offline ABI mocks/crossbuild не являются hardware PASS.

## 8. PASS/FAIL и неперенос

Общий PASS каждого этапа: scope реализован, собственные проверки успешны, входы/выходы воспроизводимы, запрещённые компоненты отсутствуют, сохранён stock, отчёт заполнен. Невыполненный обязательный software test/build — FAIL/BLOCKED, с отделением baseline failure от новой регрессии. Отсутствие Echo tests фиксируется отдельно.

| Контроль | Владелец | Доказательство |
|---|---|---|
| Нет CFG2/TLV/envelopes, per-frame status/counters, compatibility | 02, 04, 05; final 07 | New-tree diff и negative wire tests |
| Нет preflight/first-PCM/NotReady/confirmation/delayed persistence | 04–06; final 07 | Первый полный output/Active; timeout/control tests |
| Нет ASP Debug/H-E/feedback/log queues | 06; final 07 | HA entity inventory и ASP log assertions |
| Нет host platform whitelist / synthetic production adapter | 03, 05; final 07 | Generic geometry tests и production/test boundary review |
| Нет inode cache / unsafe launch wrapper | 05; final 07 | Замена установки между активациями, symlink/rename tests |
| Нет retry/async worker/new manager/second privacy/rewire revalidation | 03, 05, 06; final 07 | Behavioral tests и focused code review |
| Нет C installer/OTA/release automation/vendor assets | 04; final 07 | New repository file inventory, provenance/license review |

В проверенном upstream этих ASP legacy subsystems нет. Planned removals = **0 по умолчанию**; будущие реальные удаления перечислять с причиной и source evidence. Не считать stock service restart, stock listener channels, WebRTC helper или Panel diagnostics запрещённой новой ASP-инфраструктурой.

## 9. Вопросы и точные границы неопределённости

Архитектурных блокеров декомпозиции не обнаружено. Будущие start SHA, новые имена модулей, paths к доступному toolchain/private runtime, перечень минимальных C launch fragments и точная реализация pre-wait hook определяются в 01/04/06 по фактическому коду. Это исполнительские решения в рамках §§4, 7, 11, 14, 22, не новая архитектура.

Если фиксированная исходная база содержит иные stock seams, 01 фиксирует diff и минимальный способ сохранить контракт. Если это невозможно — узкий STOP с file/SHA/scenario; V2 не редактируется. Отдельно не требовать native External subsystem для Panel или превращать stock software REF в выдуманный hardware REF.

## 10. Документальная проверка этого пакета

[REQUIREMENTS_TRACEABILITY.md](REQUIREMENTS_TRACEABILITY.md) назначает owners/test criteria для D01–D58, F1–F5, V01–V27, A1–A9. Каждый TASK имеет PROMPT. [HANDOFF_TEMPLATE.md](HANDOFF_TEMPLATE.md) разделяет execution, owner acceptance, integration и hardware status. Ни один будущий SHA не выдуман; 01 задаёт T01/C01, далее применяется accepted-handoff rule.

Сейчас выполнена только подготовка документации. FF, dev/stage branches, code/tests changes и hardware operations не входят в эту сессию.
