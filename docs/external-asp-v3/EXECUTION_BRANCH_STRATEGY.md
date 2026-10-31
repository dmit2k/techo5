# TECHO5 / External ASP V3 — утверждённая Git-стратегия исполнения

**Статус:** OWNER ACCEPTED 2026-10-31 для перезапуска. Предыдущая стратегия C main как integration отменена. Оба main защищены.  
**Характер:** правила репозиториев, веток, базовых ревизий, интеграции этапов и handoff. **Не заменяет и не изменяет архитектуру** из `SPEC_V2_APPROVED.md`.  
**Актуальный статус:** Git cleanup завершён, обе experimental dev-main созданы от frozen bases, Stage01 принят. Stage02 stage-ветки создаются только по отдельному заданию; FF/движение main запрещены.

## 1. Назначение репозиториев

| Репозиторий / ветка | Роль | Происхождение |
|---|---|---|
| `dmit2k/techo5:main` | Неприкосновенная чистая база V3: 62603a7854283a3a07473a17c831c6ef0fd6f8d4 | Никакого FF/reset/merge в Stage01–07 |
| `dmit2k/techo5:experimental/v3-design-staging` | Только документация, задачи, audit и handoff | Пересоздать от чистого TECHO5 main одним новым docs commit после разрешённой очистки; старую историю не наследовать и не merge в код |
| `dmit2k/techo5:experimental/external-asp-v3-dev-main` | Чистая интеграционная ветка новой Go-реализации V3 | Создать от TECHO5 main 62603a7854283a3a07473a17c831c6ef0fd6f8d4, зафиксировать точный SHA |
| `dmit2k/frankenasp-v3:main` | Защищённая чистая initial база: e986ae18643e08e95f29a38aae1dd74762f6362e | Контролируемо восстановить после согласования cleanup, далее НЕ продвигать Stage01–07 |
| `dmit2k/frankenasp-v3:experimental/v3-dev-main` | Только интеграция принятых C этапов | Создать от чистого initial C main после cleanup |
| `dmit2k/techo5:experimental/asp2-host-v1.6.12-test` | Старый проверенный Go External ASP — **reference only** | Сохранить неизменённым |
| `dmit2k/FrankenASP:main` | Старый рабочий helper, DSP и инструментарий — **reference only** | Сохранить неизменённым; V3 не создавать как его ветку |

**Имена регистрозависимы:** исторический `FrankenASP` и новый `frankenasp-v3` — разные репозитории.

## 2. Происхождение и точки старта (снимок на дату решения)

- `dmit2k/techo5:main`: `62603a7854283a3a07473a17c831c6ef0fd6f8d4` — фиксированная V3 база, не выполнять FF.
- Upstream HuskerMinion/techo5:main является только read-only reference; FF не допускается.
- Исторические данные о 43 commits и разрешении FF отменены для перезапуска. Оба main строго неизменны на протяжении Stage01–07.
- `dmit2k/techo5:experimental/asp2-host-v1.6.12-test`: `9ef0792a90e1dd0f24ead40b72490ed844d2169e` — справочный код, **не база V3**.
- `dmit2k/FrankenASP:main`: `52d51de33e3105501b298c1c214db55f05a2eb82` — справочный код, **не база V3**.
- `dmit2k/frankenasp-v3:main`: `e986ae18643e08e95f29a38aae1dd74762f6362e` — initial; после одобренного cleanup не менять.
- Принятая владельцем 2026-10-31 редакция SPEC_V2_APPROVED.md имеет Git blob `b324d2161e5ea316618f23bbe6d6aedf6c2df5c8`. Прежний blob eb678cac относится к историческому контракту; AUDIT_V2_FINAL не является независимым аудитом новой raw-IDME редакции.

**Фактический bootstrap:** cleanup завершён, Go и C dev-main уже существуют на T01=62603a78 и C01=e986ae18. Stage01 owner acceptance сохранён в `stages/01_HANDOFF.md`. Оба main неизменны. При расхождении refs — STOP без самовольного reset/force-push.

## 3. Clean implementation, не old-code refactor

V3 реализуется **с чистой основы**: Go-код начинает от frozen TECHO5 main 62603a7854283a3a07473a17c831c6ef0fd6f8d4; новый C helper — из отдельного `dmit2k/frankenasp-v3` без наследования истории старого FrankenASP. Старые проекты доступны для чтения как **reference implementations** (проверенная native geometry, DSP bridge, secure launch, factory calibration, тесты и т. п.).

- Старые файлы **не переносятся целыми каталогами или массовым cherry-pick**. Переносить только реально необходимые доказанно работающие фрагменты/модули, сохраняя применимые лицензионные условия и историю происхождения кода.
- Не вводить в V3 CFG2/TLV, status/feedback framework, H/E, preflight, synthetic production DSP adapter, platform whitelist, inode cache, лишние managers/queues только для их последующего удаления.
- Пункты ТЗ «удалить старый механизм» интерпретировать для **нового чистого дерева** как «не создавать/не переносить». Если соответствующий механизм действительно обнаружится в **актуальном upstream**, определить его происхождение и применимость, не удалять чужой stock-функционал без подтверждённой необходимости.
- `frankenasp-v3` содержит минимальный необходимый runtime helper, DSP adaptation/bridge, build и conformance/tests. Старые installer, OTA, release automation и vendor-proprietary assets **не импортировать автоматически**. Vendor libraries/firmware не коммитить.
- Утверждённый функциональный контракт Go/C, mono PCM, secure launch, SoftwareCut, ownership, Selected/Active и lifecycle остаётся обязательным. **Clean implementation — способ исполнения, а не изменение V2.**

## 4. Stage-ветки и worktrees

**Общие номера этапов** для обоих проектов; одинаковый ID означает один логический этап, даже если одна сторона не меняется.

| TECHO5 | FrankenASP V3 |
|---|---|
| Интеграция: `experimental/external-asp-v3-dev-main` | Интеграция: `experimental/v3-dev-main` |
| Этап: `experimental/external-asp-v3-dev-stage01`, `...stage02`, ... | Этап: `experimental/v3-stage01`, `experimental/v3-stage02`, ... |

Правила:
1. После утверждения этапного задания создать **новую** stage-ветку от последнего **принятого интеграционного HEAD** соответствующего репозитория, в **отдельном worktree**. Имена должны совпадать с номером задания/плана.
2. Если этап затрагивает только одну сторону — на второй **не создавать пустую stage-ветку**; в handoff фиксировать неизменённый интеграционный SHA.
3. Каждый исполнитель работает только в своей stage-ветке/своём worktree; не коммитит напрямую в интеграционные ветки, документальную ветку или `main` TECHO5.
4. Выполнить изменённые tests на этапе. Отчёт содержит реальный Git SHA, команды и результаты тестов, ожидаемые удаления/непоявление запрещённых механизмов.
5. После **явной приёмки** stage интегрируется в dev-main (TECHO5) или `experimental/v3-dev-main` (frankenasp-v3); по возможности **fast-forward**, без squash/cherry-pick, скрывающих принятую ревизию. Конкретный способ интеграции и полномочия integration agent определяются отдельным поручением; исполнитель **не делает merge автоматически**.
6. Для первой интеграции Go↔C descriptor/READY/fixed PCM обе стороны должны быть собраны и проверены как **пара**. По отдельности одна новая сторона со старой не является рабочей интеграцией. В handoff обязательно указать **пару commit IDs и совместные тесты**; только затем принимать оба интеграционных HEAD.
7. Последующий этап стартует от **принятых интеграционных** SHA, а не от незавершённых stage-веток. Параллельность разрешать лишь при явной независимости файлов/контрактов и с отдельным планом интеграции.
8. Документы, stage reports и обновлённый план записываются **только** в `experimental/v3-design-staging` TECHO5; её нельзя merge/rebase в чистую реализацию.


**Прямая GitHub-публикация с validation host (проверена 2026-10-31):** на `validation host` настроены два различных SSH deploy key с write permission: `${HOME}/.ssh/techo5_v3_write` только для `dmit2k/techo5` и `${HOME}/.ssh/frankenasp_v3_write` только для `dmit2k/frankenasp-v3`. Приватные ключи остаются только на validation host; Git credentials/token с других машин не копировать. В локальных Git-конфигурациях обоих репозиториев `origin` fetch остаётся HTTPS, `origin` push — `git@github.com:...` с соответствующим `core.sshCommand` (установлены `IdentitiesOnly`, `BatchMode`, проверка known_hosts). **Все stage и docs pushes выполнять напрямую на validation host**, а не через build host или GitHub API с пересозданием Git commits. Проверенная процедура: проверить exact local commit SHA, remote target/ancestry/права, `git push origin HEAD:refs/heads/<exact-approved-stage-branch>` без force, затем `git ls-remote`/GitHub API и сравнение SHA, отсутствие движения main/dev-main, чистоту worktree. Прямой no-op push уже проверен на точных опубликованных Stage02 refs Go и C (оба `[up to date]`). Право write deploy key распространяется на репозиторий, а не только на stage-ветку: ограничения на main и dev-main продолжают действовать как обязательные правила исполнения. Если direct authentication нарушена — STOP и устранить её, без обходных публикаций или смены tested SHA.

## PCM-only correction gates (2026-10-31, higher priority for future execution)

Владелец утвердил перенос factory resource preparation из External PCM descriptor обратно в **FrankenASP launcher**. Авторитет: [PCM_ONLY_ARCHITECTURE_DECISION.md](PCM_ONLY_ARCHITECTURE_DECISION.md). Исторические Stage01/02/03 HANDOFF не переписывать; последние принятые по прошлому контракту dev-main T03=`4fb329cf789b156bbeb8d13da20b8c4bda70f108`, C02=`0f39216a5240306344af7c2457716f4eaa466f40`. **Они не считаются принятыми под PCM-only.**

Далее строго: (1) новая Go+C Stage02-R correction branches от T03/C02 → независимая приёмка → оба dev-main FF → `02_PCM_ONLY_CORRECTION_HANDOFF.md`; (2) новая Go Stage03-R branch от correction pair → приёмка → Go dev-main FF → `03_PCM_ONLY_CORRECTION_HANDOFF.md`; (3) только затем Stage04 с restore factory launcher responsibility. Исполнители меняют только свои новые stage refs; acceptance/integration coordinator отдельно. Оба `main` неизменны. До соответствующего `ACCEPTED + COMPLETE` следующий worker **не стартует**. Не удалять старые stage refs и не использовать force/rebase для коррекции.

## 5. Как адаптировать утверждённое ТЗ к новому пути

Первоначальный §19 V2 обсуждает рефакторинг старой экспериметальной реализации. **Его требования и зависимости остаются нормативными по смыслу**, но их исполняют через построение минимальной новой системы:

- Baseline начинается с **проверки изменений текущего upstream stock audio path** относительно прежней проверенной версии. Не предполагать сохранение всех точек интеграции после 43 upstream-коммитов.
- Codec/fixtures можно подготовить и проверить до подключения нового runtime; в чистом TECHO5 **нет необходимости сохранять старый ASP wire**.
- Go↔C integration gate — **первая согласованная активация V3**, а не необходимость переносить и затем удалять старый CFG2. Не создавать old/new compatibility.
- В штатном upstream может не быть ASP statuses/feedback/HA Debug/preflight; критерий — **их отсутствие в новой V3** плюс корректные Selected/Active, а не надуманный commit с удалениями несуществующих файлов.
- Lifecycle/Rewire/SoftwareCut, output ownership, secure launch и stock regression необходимо построить правильно с самого начала и проверить в назначенных этапах.
- Разрешённое расхождение: порядок/группировка технических задач и обязательство «не вводить» вместо «удалить». **Запрещённое расхождение:** изменение утверждённого внешнего поведения, IPC, hardware metadata, DSP semantics и инвариантов.
- Результат каждого этапа: работающий в оговорённом объёме и проверенный набор, без искусственных временных менеджеров/подсистем.

## 6. Что агенту декомпозиции требуется подготовить

Использовать `TASK_DECOMPOSE_APPROVED_V2.md` **в актуальной редакции**, совместно с данным документом. План должен назначать новые Git bases, форматы stage-веток и worktrees, минимально необходимую реализацию чистых компонентов, матрицу требования → этап/проверка, отдельный Go/C paired gate, handoff/промпты и статус software/hardware acceptance.

**После принятого Stage01:** Stage02–07 stage refs и код создаются исключительно по отдельным исполнительским заданиям; интеграционные refs меняются только после приёмки и отдельного поручения. Оба main, FF, Echo и hardware tests остаются вне полномочий.
