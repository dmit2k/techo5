# TECHO5 / FrankenASP External ASP — финальный аудит SPEC V2

**Дата:** 30 октября 2026 года (Europe/Riga).

**Итоговый вердикт:** **APPROVED**.

**Объект:** SPEC V2 Candidate; проверка готовности ТЗ к декомпозиции, без повторного выбора архитектуры.

**Проверенная ревизия документации:** `89c1f948e3be4bd2f27a3488c22918f76558cc63`, ветка `experimental/v3-design-staging`.

**Исходники для проверки контрактов:** TECHO5 `9ef0792a90e1dd0f24ead40b72490ed844d2169e`; FrankenASP `52d51de33e3105501b298c1c214db55f05a2eb82`.

## Основание и границы проверки

Нормативная основа — утверждённые 58 решений и пять приоритетных уточнений. Девять замечаний к V1 проверены по фактическому тексту V2 и исходникам, а не только по changelog.

Полностью рассмотрены следующие документы указанной ревизии:

1. [SPEC_V2_CANDIDATE.md](https://github.com/dmit2k/techo5/blob/89c1f948e3be4bd2f27a3488c22918f76558cc63/docs/external-asp-v3/SPEC_V2_CANDIDATE.md).
2. [CHANGELOG_V1_V2.md](https://github.com/dmit2k/techo5/blob/89c1f948e3be4bd2f27a3488c22918f76558cc63/docs/external-asp-v3/CHANGELOG_V1_V2.md).
3. [AUDIT_V1_REQUIRED_CORRECTIONS.md](https://github.com/dmit2k/techo5/blob/89c1f948e3be4bd2f27a3488c22918f76558cc63/docs/external-asp-v3/AUDIT_V1_REQUIRED_CORRECTIONS.md).
4. [REGISTER_58_DECISIONS.md](https://github.com/dmit2k/techo5/blob/89c1f948e3be4bd2f27a3488c22918f76558cc63/docs/external-asp-v3/REGISTER_58_DECISIONS.md).
5. [FINAL_FIVE_CLARIFICATIONS.md](https://github.com/dmit2k/techo5/blob/89c1f948e3be4bd2f27a3488c22918f76558cc63/docs/external-asp-v3/FINAL_FIVE_CLARIFICATIONS.md).
6. [SPEC_DRAFT_V1_2026-10-30.md](https://github.com/dmit2k/techo5/blob/89c1f948e3be4bd2f27a3488c22918f76558cc63/docs/external-asp-v3/SPEC_DRAFT_V1_2026-10-30.md).

Архивная V1 побайтово совпадает с полностью прочитанной при предыдущем аудите приложенной V1: Git blob `eda5825f92ea0cffc672361b07525bce723b5ce7`. Проверенная V2 имеет blob `eb678cacfbcc44927c65e402a9fa821fc0b56e9d`. Дополнительно сопоставлены изменения V1 → V2.

Перед подготовкой заключения проверен существующий worktree `/home/gptlab/dev/techo5-v3-design-staging` на anfisa-trainer: HEAD и ветка совпадали с указанными, рабочий каталог был чистым, документы не отличались от закреплённой ревизии, файла `AUDIT_V2_FINAL.md` не существовало.

Это статическая проверка документации и исходников. Сборки, тесты реализации, запуск helper, аппаратные и акустические испытания не выполнялись. Разрешённое изменение в рамках аудита — только этот документ и его публикация в заданной ветке.

## A. Проверка девяти замечаний

| № | Статус | Разделы V2 | Техническое заключение |
|---|---|---|---|
| 1 | **FIXED** | §5.2 | FD3/FD4 явно сохранены для executable/assets; PCM использует stdin/stdout pipes с изолированным wire FD внутри FrankenASP. Требования переиспользовать secure-launch FD для PCM нет. |
| 2 | **FIXED** | §5.3.1–5.3.3, §5.3.5–5.3.6, §20.1 | Размеры и смещения записей согласованы; заданы коды PCM, signed valid bits и LSB shift, покрытие channel records, независимая нумерация MIC/REF, последовательные buffer offsets и проверка последнего sample. Контракт достаточен для независимых Go/C реализаций. |
| 3 | **FIXED** | §4.4, §5.3.4, §7.1 | PlatformInfo и MicFactoryGain имеют конкретные сохраняемые value encodings без TLV envelope. Сохранены detection, encoding, availability, provenance, applied-to-PCM, present mask и связь коэффициентов с captured MIC slots. Непрояснённый hardware payload не введён. |
| 4 | **FIXED** | §5.3.2, §5.3.5–5.3.6 | Список buffers не ограничен двумя. Предел 8192 относится к descriptor, достаточен для сохраняемых metadata и не требует максимальной предварительной аллокации. Он не является ограничением размера PCM-блока и не обязывает helper поддерживать каждую валидную конфигурацию. |
| 5 | **FIXED** | §7.1–7.2, §20.1, §24 | Представимость, корректный разбор и DSP compatibility разделены. FrankenASP сохраняет рабочие конфигурации и вправе отказать до READY. Новые resamplers и расширение vendor compatibility не требуются; положительные и отрицательные проверки разделены. |
| 6 | **FIXED** | §6.6, §20.2 | Startup 10 секунд ограничивает только инициализацию до READY; после READY снимается без отмены session context. Риск использования истекающего контекста в exec.CommandContext назван явно. Built-in продолжает работать при старте; предусмотрена регрессия работоспособности после исходных десяти секунд. |
| 7 | **FIXED** | §11.2, §11.6, §12, §20.6–20.7 | Cancel и закрытие pipes выполняются до ожидания Run; cmd.Wait имеет одного владельца. Shutdown ожидает reaper без второго Wait. Cleanup не возвращён в capture-only Close; поздние завершения ограничены session identity. |
| 8 | **FIXED** | §19, R2–R9 | R2 готовит codec без runtime switch. R3 принимает согласованную Go/C пару и сразу удаляет wire-зависимый control plane, preflight и конфликтующие таймеры. R5 не наследует эти таймеры. R6/R7 завершают чистку, затронутые тесты обновляются на каждом этапе. |
| 9 | **FIXED** | §19/R9, §22, §24 | Программная приёмка отделена от vendor runtime на Echo и акустики. Отложенная аппаратная проверка явно имеет статус «НЕ ВЫПОЛНЕНА» и не блокирует закрытие программного рефакторинга. Неявного разрешения на запуск на Echo нет. |

**Результат:** 9 FIXED; 0 PARTIALLY FIXED; 0 NOT FIXED.

### A.1. Проверка бинарного контракта

Размеры получаются непосредственно из полей: header — 24 байта; buffer record — 28; channel record — 8. При обозначениях B = число buffers, C = число channels, P/H/G = длины platform/hardware/gain sections общий размер описанного descriptor составляет:

`D = 24 + 28×B + 8×C + P + H + G`.

Для v1 H = 0. Поля количества задают представимые значения, а общий предел 8192 ограничивает допустимые сочетания; это совместные ограничения. Например, 255 buffers и 255 channels требуют 9204 байта ещё без metadata и потому не проходят общий предел. Это не противоречит диапазону uint8 и не возвращает ограничение «не больше двух buffers».

Последовательные offsets, сумма Length, непрерывное покрытие channel records и проверка последнего sample устраняют неоднозначности V1. Различение packed S24_3LE и S32LE с 24 значащими битами задано явно. Нулевой REF выражается отсутствием REF records. При наличии gain его индексы относятся к captured MIC Role index, а не к vendor logical mapping.

Для текущей формы Athena — два buffers, четыре MIC, два REF, platform codename «athena» и четыре Q30 coefficients — descriptor занимает **186 байт**:

`24 + 2×28 + 6×8 + (4+6) + (16+4×8) = 186`.

Даже пример с прежними codec-пределами 16 MIC и 8 REF, двумя buffers, максимальным codename и полной calibration занимает 452 байта. Поэтому 8192 достаточно для действительно сохраняемых данных; оснований требовать больший предел не обнаружено. Предел 16 slots в gain ограничивает только существующее кодирование calibration, а не общий PCM-интерфейс.

PCM payload считается отдельно. Для текущих MIC4/S32 и REF2/S16 при 96 kHz за 20 мс это `30720 + 7680 = 38400` байт — больше 8192, но допустимо, поскольку 8192 ограничивает descriptor. Размер входа определяется Total input block size и фактическими buffers, а совместимость и необходимые ресурсы конкретной DSP-конфигурации проверяет helper до READY. Максимальная аллокация по wire-пределу не требуется.

Выход остаётся ровно 640 байт mono S16LE. Пять секунд ограничивают один полный write/read exchange, а не паузу capture между обменами. Дополнительное framing для принятого синхронного контракта не требуется.

### A.2. Сверка metadata с исходниками

Описанные в §5.3.4 байты PlatformInfo совпадают с [Go marshal/parse](https://github.com/dmit2k/techo5/blob/9ef0792a90e1dd0f24ead40b72490ed844d2169e/echod/internal/asp/cfg2_metadata.go#L55) и [C decode_platform](https://github.com/dmit2k/FrankenASP/blob/52d51de33e3105501b298c1c214db55f05a2eb82/tools/external-asp/cfg2.c#L63): однобайтная schema, detection, длина codename, reserved byte и ASCII codename.

Prefix и коэффициенты MicFactoryGain совпадают с [Go marshal/parse](https://github.com/dmit2k/techo5/blob/9ef0792a90e1dd0f24ead40b72490ed844d2169e/echod/internal/asp/cfg2_metadata.go#L157) и [C decode_gain](https://github.com/dmit2k/FrankenASP/blob/52d51de33e3105501b298c1c214db55f05a2eb82/tools/external-asp/cfg2.c#L87). Сохраняются различия absent/NONE/COMPLETE/PARTIAL, нулевые masked-out slots, ограничения положительных Q30 значений диапазоном signed int32 и applied_to_pcm=0. Порядок коэффициентов привязан к captured MIC slots.

Сохранить эти два конкретных value encodings и удалить TLV оболочку технически корректно. Это не требует сохранения старого протокола, registry или старых названий методов. Проверка совместимости с vendor DSP остаётся ответственностью helper; схема metadata не даёт generic host права возвращать Athena-only admission.

Право FrankenASP отказать подтверждается и существующей реализацией: [cfg2_runtime_init](https://github.com/dmit2k/FrankenASP/blob/52d51de33e3105501b298c1c214db55f05a2eb82/tools/external-asp/cfg2_runtime.c#L34) уже отделяет декодирование от принятия конкретной DSP-конфигурации. V2 сохраняет эту границу, не превращая перечень wire formats в обещание обработки всех вариантов.

### A.3. Сверка process lifecycle

Назначение FD3/FD4 подтверждено [prepareSecureASPPinned](https://github.com/dmit2k/techo5/blob/9ef0792a90e1dd0f24ead40b72490ed844d2169e/echod/internal/hardware/mic/asp_pinned_exec_linux.go#L48), подключение pipes и единственный reaper — [asp.go](https://github.com/dmit2k/techo5/blob/9ef0792a90e1dd0f24ead40b72490ed844d2169e/echod/internal/hardware/mic/asp.go#L110), изоляция wire от stdout логов — [factory_child.c](https://github.com/dmit2k/FrankenASP/blob/52d51de33e3105501b298c1c214db55f05a2eb82/tools/external-asp/factory_child.c#L60).

Требование отменять IPC до ожидания Run учитывает существующий [service.Group.stop](https://github.com/dmit2k/techo5/blob/9ef0792a90e1dd0f24ead40b72490ed844d2169e/echod/internal/service/group.go#L112), который ждёт running services перед Close. В V2 не осталось требования ждать блокирующий exchange прежде его отмены.

Startup 10 секунд — явно обозначенный проектный default, не измеренная характеристика устройства. Его численное значение не создаёт противоречия с реестром. Разделение startup timeout и session cancellation устраняет доказанный риск преждевременного убийства child через exec.CommandContext.

## B. Дополнительные обнаруженные проблемы

**Дополнительных технических ошибок или противоречий, требующих исправления V2 до декомпозиции, не обнаружено. Обязательных редакционных правок нет.**

Проверка полного текста V2 не выявила возвращения ранее отвергнутых механизмов:

- Native access, hardware configuration и stock conversion сохраняют согласованные границы ответственности; Built-in и сенсоры остаются на штатной обработке.
- SoftwareCut использует одно решение для native и stock. Нормативный пункт 19 реестра сохраняет раннюю границу до conversion и stock clear по тому же значению; §8 V2 не вводит вторую policy или output gate.
- Borrowed native input не распространяется на опубликованный output: §9 отдельно сохраняет неизменность кадра для подписчиков.
- Session identity защищает отменённые launch/response; эта проверка не является capture generation или per-frame timeline tracking.
- Rewire/XRUN сохраняют session, DSP/FIFO и pipes. Ни descriptor renegotiation, ни повторный startup padding не требуются.
- Удаление statuses, H/E codes, feedback/log queues и ASP Debug задано полностью. Три вида host log и отдельный stderr не образуют новый diagnostic protocol.
- Inode cache удаляется при сохранении безопасного запуска из /data и FD-pinning.
- Аппаратная приёмка остаётся отдельной и не подменяется прохождением программных fixtures.

Общие слова о «resampling» в перечне ответственности helper и checklist читаются вместе с конкретным ограничением §7.1–7.2: новые resamplers и расширение поддерживаемых DSP-конфигураций исключены. Требования поддержать все описываемые descriptor случаи нет.

Этот вывод относится к достаточности ТЗ. Он не означает, что будущая реализация уже собрана, прошла conformance или проверена на устройстве.

## C. Готовность к декомпозиции

### Wire contract

Контракт достаточно определён для раздельной реализации Go writer и C reader с независимыми fixtures: известны байты, размеры, порядок секций, layout, metadata, READY и фиксированный output. Все необходимые вопросы из замечаний V1 закрыты без нового compatibility или validation framework.

### Границы и зависимости этапов

| Этап | Основание для самостоятельной приёмки | Зависимость |
|---|---|---|
| R1 | Зафиксированные baseline, ветки, исходные проверки и diff | Начальный этап |
| R2 | Go/C codec и независимые conformance fixtures; рабочий handshake не переключён | R1 и зафиксированный §5.3 |
| R3 | Компилируемая согласованная Go/C пара с новым runtime wire, padding и удалёнными непосредственными wire/control dependencies | R2; приёмка обеих сторон совместная |
| R4 | Native seam, SoftwareCut и ownership с локальными регрессиями | Принятый runtime-контракт R3 |
| R5 | Сохранение процесса при Rewire/XRUN и корректные cancel/shutdown/Wait | Удаление старых таймеров в R3; согласованный capture path |
| R6–R7 | Завершённое упрощение Selected/Active и полное удаление diagnostics; допускается объединённая приёмка | Wire-зависимый state уже удалён R3 |
| R8 | Общая regression/build/race/security/stock проверка, остаточный cleanup и diff | Завершённые изменения R2–R7 |
| R9 | Программная приёмка окончательной пары commit IDs и отчёт | R8; аппаратная приёмка не включается автоматически |

Такие этапы можно выполнять в отдельных агентских сессиях, передавая каждой зафиксированную входную ревизию и результат предшествующего этапа. Это не означает, что все этапы независимы по данным или могут одновременно менять один исходный baseline.

R3 — обязательный совместный интеграционный рубеж двух репозиториев. По отдельности новая Go-сторона и старая C-сторона, либо наоборот, рабочей парой не считаются. Пара commit IDs решает задачу приёмки без временной old/new wire compatibility.

Краткий перечень R6–R7 не отменяет обязательства полного ТЗ. При последующей декомпозиции нужно распределить также явно сохранённые требования §4.5/§14/§17: удаление platform admission, inode cache и synthetic adapter. Это назначение ответственности существующим заданиям, а не необходимость нового раздела архитектуры или изменения V2.

Окончательные исполнительские задания, стартовые промпты и новые ветки этим аудитом не создаются.

## D. Итоговый вердикт

**APPROVED**

SPEC V2 Candidate ревизии `89c1f948e3be4bd2f27a3488c22918f76558cc63` готова к утверждению и последующей декомпозиции.

Все девять обязательных замечаний к V1 закрыты. Дополнительных технических блокеров и обязательных редакционных исправлений не выявлено. Принятые 58 решений и пять уточнений сохранены; расширения DSP functionality, новой диагностической подсистемы, recovery framework или несовместимого порядка жизненного цикла ТЗ не требует.

Следующий этап — отдельная сессия декомпозиции по §19 с самостоятельными критериями приёмки и фиксированными зависимостями. Утверждение ТЗ не является подтверждением выполнения будущих тестов и не разрешает аппаратные операции на Echo.
