# TECHO5 External ASP — реестр 58 архитектурных решений

**Дата согласования:** 30 октября 2026 года  
**Основание:** три раунда архитектурного аудита и согласованные решения в TECHO5/FrankenASP-чате.  
**Статус:** редакционно исправленный для V3 restart реестр, новая поправка ожидает подтверждения владельца; ранее утверждённая редакция сохранена в истории; пять приоритетных поправок хранятся отдельно в `FINAL_FIVE_CLARIFICATIONS.md`.

> **Происхождение:** текстовая редакционная копия утверждённого в чате 58-пунктного реестра. Нумерация и смысл решений сохранены. Для дословной сверки с оригинальной формулировкой использовать сообщение в исходном чате, начинающееся «TECHO5 External ASP — финальный реестр архитектурных решений для утверждения». Это не новая архитектура и не замена пяти поправок.
>
> А1 — первый аудит; А2 — углублённый аудит; А3 — финальные уточнения аудитора; Д — решения обсуждения. Приоритет: **58 пунктов + пять финальных поправок**, затем утверждённое ТЗ.

## I. Архитектура и ответственность

1. **External ASP — необязательное расширение.** Без helper TECHO5 полностью использует Built-in, не заводя вторую аудиосистему. Почему: цель — небольшой upstream PR. [А1–А3]
2. **Точка интеграции — `mic.Source`.** Не создавать отдельных capture owners, supervisor или путей публикации. Почему: все потребители уже используют общий источник. [А1, А2, Д]
3. **Host предоставляет исходный MIC/REF PCM byte-for-byte.** Где stock делает конверсию, External получает данные до неё. Оба аппаратных stereo REF L/R передаются в исходном порядке, формате, частоте и interleaving, без scaling, reordering, downmix и синтезирования REF. Почему: внешнему DSP нужна исходная акустическая информация. [А1–А3]
4. **Descriptor описывает фактический capture.** Частоты, форматы, bit depth/alignment, каналы/порядок, размеры блоков и только доступные исходные bytes hardware/IDME metadata, без host-side normalisation, confidence, Q30 или MIC-slot calibration mapping. Не навязывать DSP-требования. [А2, А3, Д]
5. **DSP-специфическая адаптация — только в helper.** Mapping, REF stereo L/R selection/reordering, scaling, resampling, quantum/FIFO, factory IDME parsing, Q30/gain/masks, MIC-slot interpretation и qualification не выполняются в generic TECHO5. *Границы этого пункта уточнены финальной поправкой №1*: штатный hardware capture/metadata/Built-in остаются в TECHO5. [Д]
6. **Удалить platform admission в generic host.** Никаких Athena-only whitelist, MIC4/REF2/96-kHz ограничений ради FrankenASP. [А1–А3, Д]
7. **Использовать существующий PCM напрямую, если capture не конвертирует его.** Не плодить платформенные обёртки. [А3, Д]
8. **Минимальный универсальный native tap при скрытой stock-конверсии.** Допустим необязательный `NativeFormat()` / `ReadWithNative()` до `convert()`, но имена не обязательны. [А3]
9. **Не переписывать capture backends.** Сохранить обычный `Read()` и существующий stock conversion; не создавать External frontend по модели Echo. [А3, Д]
10. **Borrowed native input.** Использовать входные буферы синхронно, до следующего capture read, без новых очередей и копий. Отдельная гарантия опубликованного output — в поправке №4. [А2–А3]
11. **Не держать publication mutex на время blocking helper IPC.** Сохранять короткие критические секции публикации. [А1–А2]

## II. Stock аудиотракт

12. **Полностью сохранить Built-in DSP.** Mixer, AEC, denoise, leveling и другие штатные стадии продолжают обрабатывать поток. [А1–А3]
13. **Не отключать части Built-in ради CPU в первом PR.** Без отдельных измерений такая оптимизация не нужна. [А1–А2]
14. **Штатные сенсоры остаются на штатных данных.** Noise, facing, standout, levels не переводятся на External output. [А1–А2]
15. **Один выбор и одна публикация mono PCM.** History, MWW, STT, spectrum, подписчики получают единый выбранный поток; выход helper — согласованный штатный формат. [А1–А3]
16. **Не менять `ListenRaw()` и телефонный тракт ради External.** SIP/intercom используют общий mono path; WebRTC AEC switching остаётся Built-in. [А2]

## III. Software Mute

17. **Одна software-mute policy в TECHO5.** Не создавать в helper privacy supervisor/state machine. [А1–А3, Д]
18. **Исправить обход SoftwareCut native потоком.** До передачи внешнему DSP native MIC должен быть заглушён при software mute: сейчас `processASP(native)` раньше stock `clear(raw)`. [А1–А3]
19. **Одна проверка SoftwareCut для native и stock.** По одному значению заглушать native MIC/REF до conversion и сохранять stock `clear(raw)`; это не две privacy policies. [А3, Д]
20. **Сохранить MIC/REF mute semantics.** Software mute stock глушит MIC+REF; hardware mute может оставлять REF по аппаратным правилам. [А2–А3, Д]
21. **Не добавлять второй mute check после External DSP.** Никаких output masking/epochs/проверок перед history/STT/MWW. [Д, А3]
22. **Не сбрасывать DSP/FIR/history при mute/unmute.** Короткий фильтровый хвост допустим; постоянная передача живого MIC — нет. [А2–А3, Д]
23. **ALSA codec mute для Spot вне текущего рефакторинга.** Не входит в требования первого PR без аппаратных доказательств. [Д, А3]
24. **Panel не требует специальной External реализации.** Сохранить stock сборку/работу Android Panel, без новой платформенной подсистемы. [Д, А3]

## IV. IPC и протокол

25. **Синхронный PCM exchange.** Один outstanding block/response; никаких async worker, audio queue, stale age tracking. [А2, Д]
26. **5-секундный аварийный deadline PCM exchange.** Один дедлайн на фактический write/read, не idle timeout между блоками. [А2–А3, Д]
27. **Отдельный конечный deadline инициализации до READY.** Удалить конфликтующие HELLO 500ms и first-PCM 1s; не ограничивать им жизнь готового процесса. [А2–А3]
28. **Минимальный wire: Descriptor → READY → fixed MIC/REF input → fixed PCM output.** Не строить новый messaging framework. [А1–А3, Д]
29. **Удалить универсальный CFG2/TLV и старые HELLO/ACK2/INP2/RSP2 envelopes.** Сохранить лишь действительно нужные поля компактного descriptor. [А1–А3]
30. **Без per-frame framing/metadata.** Нет sequence/origin/counters/magic/checksum/record headers/status; размеры и порядок известны из синхронного IPC. [Д, А3]
31. **Полностью удалить wire status codes/messages.** H/E-коды, origin/status и diagnostic events не отправляются по PCM IPC. [А1, Д]
32. **Удалить `NotReady`, `State`, `Event` ответы.** Ошибки по EOF, timeout, IPC/process exit; в норме полный PCM блок. [А1–А3]
33. **Диагностический stderr не смешивается с PCM.** Выделенный wire stream остаётся чистым, новый diagnostic channel не требуется. [А3]
34. **FrankenASP: однократное 64-sample output FIFO padding.** В новой DSP-сессии, НЕ в MIC/REF input; после каждого 20ms exchange — 320 mono samples, без повторов после Rewire/mute. [А1–А3]

## V. Жизненный цикл и fallback

35. **Helper не зависит от отдельных capture Run/Close.** Он принадлежит External-сессии и жизненному циклу echod. [А3, Д]
36. **Удалить наш `defer s.closeASP()` из `Source.Run()` и `s.closeASP()` из `Source.Close()`.** Stock эти вызовы не делает. [Д]
37. **При `Rewire()` оставить тот же helper и session.** Временно нет PCM, после reopen обмен возобновляется, без нового handshake/reset. [Д, А3]
38. **Не проверять неизменность descriptor после штатного reopen.** Никаких generations и гипотетического hot-swap admission. [Д]
39. **ALSA XRUN/overrun не завершает External.** Соблюдать existing stock `continue`, не перезапускать DSP. [Д, А3]
40. **Нет auto retry или recovery state machine.** После реального отказа пользователь может выбрать Built-in → External. [Д, А3]
41. **Завершать helper по настоящим причинам:** explicit Built-in, реальный launch/IPC failure или echod shutdown. Закрывать pipes/process, выполнять единственный Wait. [А2–А3, Д]
42. **Cleanup в существующем долгоживущем lifecycle, не в capture Close.** Новый supervisor запрещён, конкретное место выбирается минимально. [Д, А3]

## VI. HA и диагностика

43. **Только Selected и Active.** Selected сохраняется сразу, fallback меняет Active, не Selected; убрать Confirmed/keep/delayed persistence. [А1–А3]
44. **Удалить 3-frame preflight и confirmation state machine.** Первый валидный PCM может активировать External. [А1–А3]
45. **HA External option не зависит от наличия executable на startup.** Поздняя установка и ручная активация без echod restart. [А2–А3, Д]
46. **Полностью удалить собственные numeric status codes/messages, включая H:/E:.** Ранее утверждённый исторический диагностический формат отменён этим решением. [А1, Д]
47. **Ровно три типа host log:** launch; fallback с причиной; session/process termination. Отдельный stderr helper допустим. [Д]
48. **Удалить `ASP Debug` из HA со всей обвязкой.** Нет нового diagnostic entity/health/status attributes. [Д]
49. **Удалить целиком feedback/logging subsystem:** `asp_feedback`, `asp_log`, custom queues, suppression, counters, formatters, диагностические поля. [А1–А2, Д]

## VII. Запуск и production cleanup

50. **Сохранить secure helper launch из `/data`.** Path traversal, owner/mode/type validation, O_NOFOLLOW, FD pinning на действительных trust boundaries. [А1–А3]
51. **Удалить lifetime inode identity cache** (`discoveredASP = sync.OnceValue(...)`). Новая manual activation видит обновлённый executable. [А1–А3, Д]
52. **Уменьшить дублирующую validation и лишние launch/test abstractions.** Сохранять проверки на реальных границах доверия. [А1–А2]
53. **Удалить synthetic DSP adapter из production TECHO5.** Тестовые fake DSP остаются только в test code. [А1–А2]

## VIII. Проверки и завершение

54. **Удалить тесты исключённых механизмов:** старые status/feedback/preflight/confirmation/wire/25ms/AST matrices. [А1–А2]
55. **Сохранить регрессии реальных контрактов:** Go↔C codec, sync PCM, native input, SoftwareCut, output ownership, FIFO/padding, fallback, Rewire/XRUN, cancellation/session identity, HA, secure launch. [А1–А3]
56. **Сохранить stock и supported build compatibility:** Built-in без helper, subscriptions/sensors/calls, platform build tags, Android Panel; mute transitions могут отличаться от stock побитово. [А2–А3]
57. **Согласованная Go/C миграция.** Новый wire применяется обеими сторонами, без совместимости со старым неопубликованным протоколом. [А1–А3]
58. **Оценивать реальное сокращение архитектурной сложности, а не переименование.** Никаких новых queues/managers/validators/status frameworks взамен удалённых. Оценка 1100–1700 Go production LOC — ориентир, не жёсткая квота. [А1–А2, Д]

## Нормативные уточнения

Применять все пять приоритетных положений из [FINAL_FIVE_CLARIFICATIONS.md](FINAL_FIVE_CLARIFICATIONS.md). Последующие девять замечаний к V1 относятся к **конкретизации ТЗ**, не отменяют архитектурных решений.

**Не входит в этот реестр:** installer, OTA extraction, release automation, vendor runtime assets, операции с Echo, Git branch policies (они задаются отдельно как условия реализации).
