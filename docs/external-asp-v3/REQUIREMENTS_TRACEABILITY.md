# Матрица требований → исполнитель → проверка

Текущая трассировка включает **принятую владельцем PCM-only поправку от 2026-10-31** ([решение](PCM_ONLY_ARCHITECTURE_DECISION.md)). D01–D58 — номера [реестра](REGISTER_58_DECISIONS.md), F1–F5 — [уточнения](FINAL_FIVE_CLARIFICATIONS.md), V01–V27 — разделы [действующего ТЗ](SPEC_V2_APPROVED.md), A1–A9 — идентификаторы закрытых замечаний **исторического аудита V1** (его отдельный файл отсутствует в текущем clean-base docs tree). Предыдущие Stage02/03 tests/acceptance прошли по старому wire; **PCM-only correction tests/acceptance ещё PENDING**.

В столбце «Владелец» один этап отвечает за закрытие строки; вспомогательные implementation/verification stages указаны в критерии. Test IDs B/C/N/H/W/L/S определены в TASK 01/02/03/04/05/06/07. Итоговый 07 проверяет всю матрицу, не переносит на себя выполнение пропущенных этапных tests. Точные команды появятся в принятых handoffs после создания кода.

## 58 решений

| ID | Обязательство | Владелец | Критерий проверки |
|---|---|---|---|
| D01 | Необязательный External, полный Built-in | 06 | L01/L10, S07: без helper штатный поток и consumers работают |
| D02 | Единственная интеграция mic.Source | 06 | L07/L10, S09: один capture owner и mono publication |
| D03 | Исходный PCM до stock conversion | 03 | N01/N02: native bytes наблюдаются до convert |
| D04 | Фактический PCM-only capture descriptor | 02-R | C01–C05, 03-R N07: реальные PCM buffers/roles, no metadata |
| D05 | DSP adaptation у helper, vendor resources у FrankenASP launcher, PCM у host | 04 | H01/H02/H08 и 03-R N07: no host metadata/qualification |
| D06 | Нет generic platform admission | 03 | N02 и S09: корректные non-Athena/no-REF layouts не отвергаются host |
| D07 | Прямое использование подходящего capture PCM | 03 | N02/N04: plain buffer без DSP wrapper |
| D08 | Минимальный optional native tap | 03 | N01: paired callback до convert, прежний Read сохранён |
| D09 | Не переписывать capture backends | 03 | N01/N08: API/stock builds и Panel сохранены |
| D10 | Borrowed native input | 03 | N04; 05 W03: exchange завершён до следующего read |
| D11 | Не держать publication mutex на IPC | 06 | L04: подписки/управление доступны при blocked I/O |
| D12 | Все Built-in DSP стадии сохранены | 06 | L10/S07: mixer/AEC/denoise/leveling даже при External |
| D13 | Не отключать Built-in ради CPU | 06 | L10/S09: нет External skip stock DSP |
| D14 | Sensors на stock данных | 03 | N05 и 06 L10: noise/facing/standout/levels |
| D15 | Один выбранный mono поток | 06 | L07: history/MWW/STT/spectrum/listeners получают тот же frame |
| D16 | ListenRaw/calls/WebRTC сохранены | 06 | L10/S07: raw path, SIP/intercom, новый callCanceller |
| D17 | Единая software mute policy host | 03 | N03/N06: single value, helper без privacy state |
| D18 | Native MIC заглушён до helper | 03 | N03; 06 L08: nonzero source не проходит mute boundary |
| D19 | Одна проверка native+stock | 03 | N03/N06: early zero и stock clear по тому же value |
| D20 | Software MIC+REF, hardware REF семантика | 03 | N03/N06: software оба; аппаратное поведение не подменено |
| D21 | Нет второго output mute check | 03 | N06; 06 L08/S09: нет output masking/epochs |
| D22 | Нет mute resets DSP/FIR/history | 03 | N06; 04 H03/06 L08: хвост допустим, padding не повторяется |
| D23 | Spot ALSA codec mute вне scope | 01 | B02 и S09: нет нового Spot hardware control |
| D24 | Panel без отдельной External subsystem | 03 | N08; S08 Android builds, stock Panel сохранён |
| D25 | Sync один outstanding block | 05 | W01/W03: полный последовательный write/read без worker/queue |
| D26 | Общий PCM deadline 5s | 05 | W04: budget общий, не idle и не на каждую syscall |
| D27 | Startup deadline отдельно до READY | 05 | W05: timeout до READY, живой child после >10s; L01 stock во время старта |
| D28 | Descriptor→READY→fixed PCM | 05 | W01/W02/W08 compiled pair, C01 codec |
| D29 | Не переносить CFG2/TLV/old envelopes | 02 | C06; H08/W10/S09: rejection старого wire и absence в production |
| D30 | Нет per-frame metadata/framing | 05 | W02/W03/W10: только input bytes и 640 output |
| D31 | Нет wire status/messages | 04 | H04/H08, W10: stderr/EOF вместо status |
| D32 | Нет NotReady/State/Event responses | 04 | H03/H04: первый output полный; W02 |
| D33 | Stderr отдельно от PCM | 04 | H04; W08: vendor stdout chatter не попадает в wire |
| D34 | 64-sample output padding один раз | 04 | H03; L05/L08: 8ms quantum, без input/repeated padding |
| D35 | Session не принадлежит capture Run/Close | 06 | L05/L06: process lifetime длиннее capture lifetime |
| D36 | Не вводить closeASP в Run/Close | 06 | L05/S09: в чистой базе удалять нечего, hook не добавлен там |
| D37 | Rewire сохраняет helper/session | 06 | L05: same PID, один READY/descriptor, FIFO сохранён |
| D38 | Нет descriptor revalidation после reopen | 06 | L05/S09: нет generation/hash/renegotiation |
| D39 | XRUN не завершает helper | 06 | L05: stock continue и same session |
| D40 | Нет auto retry/recovery state machine | 06 | L02: после failure только manual Built-in→External |
| D41 | Настоящие termination causes и один Wait | 05 | W06; 06 L06: explicit stop/failure/shutdown, reaped child |
| D42 | Cleanup в existing long-lived teardown | 06 | L06: cancel/pipes до group wait, не новый supervisor |
| D43 | Selected/Active, immediate persistence | 06 | L01/L02/L09: fallback сохраняет Selected |
| D44 | Нет preflight, первый PCM активирует | 06 | L01: Active по первому successful exchange, W10 absence timers |
| D45 | HA option без executable на startup | 06 | L09: поздняя установка и manual activation без restart |
| D46 | Нет numeric H/E/status messages | 06 | L09/S09: ASP logs без codes, H08 wire absence |
| D47 | Ровно три ASP host event types | 06 | L09: launch/fallback reason/termination, stock logs сохранены |
| D48 | Не переносить HA ASP Debug | 06 | L09/S09: entity inventory без Debug/замены |
| D49 | Не переносить feedback/log subsystem | 06 | L09/S09: нет queues/suppression/counters/router/formatters |
| D50 | Secure launch из /data | 05 | W07/W08: traversal/mode/type/owner/O_NOFOLLOW/FD pinning |
| D51 | Нет lifetime inode cache | 05 | W07; L09: replacement виден новой manual session |
| D52 | Минимальные launch/validation abstractions | 05 | W07/W10/S09: trust boundaries сохранены, лишние wrappers не созданы |
| D53 | Нет production synthetic DSP adapter | 04 | H06/H08; C06/W10/S09: mocks только tests, generic host без fake DSP |
| D54 | Не переносить legacy-only tests | 07 | S09: нет status/preflight/old-wire/25ms/AST matrices; реальные regressions сохранены |
| D55 | Реальные contract regressions | 07 | S01–S08 агрегируют C/N/H/W/L tests без отсрочки их реализации |
| D56 | Stock/build compatibility | 07 | S07/S08: Built-in/consumers/tags/Android; N06 mute exceptions |
| D57 | Согласованная Go/C реализация | 05 | W01–W10 exact T/C pair, S02 финальная пара; no old compatibility |
| D58 | Минимальная architecture, LOC не квота | 07 | S09/S10: diff от T01/C01, no new managers/queues/registries |

## Пять приоритетных уточнений

| ID | Обязательство | Владелец | Критерий |
|---|---|---|---|
| F1 | Host capture/PCM/Built-in; FrankenASP launcher vendor resource aliases; Calibrator DSP calibration | 02-R/03-R/04 | C03/N07/H02/H08: no metadata wire, no host External factory reader, launcher owns factory view |
| F2 | Фактический input representation и единый mono S16LE/16k/320 output | 02 | C01–C05; 03 N01/N02 actual buffers, 05 W02 fixed output |
| F3 | Cancellation/identity устаревшего launch/PCM, без generation/queues | 06 | L03/L04 все 8 races, W06 old cleanup не трогает new session |
| F4 | Borrowed input отдельно от stable output | 06 | N04 input; L07 retention во время reads/exchanges/fallback/toggles |
| F5 | Разрешённые mute transition differences без компенсации | 03 | N03/N06 stock resampler/meters на нулях, no reset/second check; S07 stock вне mute |

## Все 27 разделов V2

| ID | Раздел | Владелец | Критерий |
|---|---|---|---|
| V01 | Цель/минимальность | 07 | S09/S10 и upstream diff |
| V02 | Scope и исключения | 01 | B02/B05; H08/S09 no installer/OTA/assets/extra platform |
| V03 | mic.Source/stock/output | 06 | L07/L10/S07 |
| V04 | Native PCM-only boundary и resource ownership | 03-R/04 | N01–N08; H01/H02/H08 helper-owned resources |
| V05 | Descriptor/transport | 02 | C01–C07; 05 W01/W08 подтверждает runtime и FD roles |
| V06 | PCM/deadlines | 05 | W02–W05: 640 bytes, 5s total, startup10s отдельно |
| V07 | FrankenASP runtime/FIFO | 04 | H01–H08; W02 compiled pair |
| V08 | SoftwareCut | 03 | N03/N06; L08 |
| V09 | Published output ownership | 06 | L07 External/fallback/toggles, N05 stock baseline |
| V10 | Selected/Active | 06 | L01/L02/L09 |
| V11 | Lifecycle/cleanup | 06 | L05/L06; W06 sole Wait |
| V12 | Cancellation/session identity | 06 | L03/L04 и W06 |
| V13 | Failure/fallback/manual retry | 06 | L02/L09 |
| V14 | Secure launch | 05 | W07/W08; H07 launch build |
| V15 | Home Assistant | 06 | L09 |
| V16 | Logging/no diagnostics | 06 | L09/H08/W10/S09 |
| V17 | Production cleanup/no replacement complexity | 07 | S09 + absence owners по таблице ниже |
| V18 | Concrete change areas | 06 | L01–L11; seam ownership 03/04/05 в плане |
| V19 | Execution dependencies/paired gate | 05 | W01–W10 + accepted handoff chain 01–07 |
| V20 | Mandatory tests | 07 | S01–S08 и таблицы C/N/H/W/L |
| V21 | Test infrastructure | 02 | C01/C06; H06/W10/S09: independent fixtures/test-only fakes |
| V22 | Builds/software vs hardware | 07 | S07/S08/S10, hardware отдельное НЕ ВЫПОЛНЕНЫ |
| V23 | Metrics | 07 | S10: LOC/files от clean T01/C01, reference отдельно |
| V24 | Acceptance checklist | 07 | S01–S10, все D/F/A строки закрыты |
| V25 | Git/results | 01 | B01 и handoff каждого этапа, S10 pair/refs/clean |
| V26 | Minimal final implementation | 07 | S09, no hidden frameworks |
| V27 | Executor STOP discipline | 01 | B01/B02; каждый TASK STOP с code/SHA/scenario, без изменения V2 |

## Девять закрытых замечаний — обязательная сохранность исправлений

| ID | Исправление | Владелец | Проверка |
|---|---|---|---|
| A1 | FD3/FD4 не PCM | 05 | W08 + H04: pinned exec/assets, 0/1 pipes, C wire isolation |
| A2 | Полный PCM-only binary descriptor | 02-R | C01/C02/C04: **16/28/8**, no metadata, valid bits/shift, channel coverage/offsets/last sample |
| A3 | Calibration вне PCM IPC | 02-R/03-R/04 | C03/N07/H02/H08: no IDME bytes/fields; launcher provides permitted vendor resources and vendor Calibrator owns gain |
| A4 | Buffer count не 1–2, bounded PCM-only descriptor отдельно от PCM input | 02-R | C02/C04/C05: >2 buffers, stereo REF L/R, max header+records 9196, overflow reject |
| A5 | Parsing не обещает DSP compatibility | 04 | H01/H02 и W02: supported positive, valid unsupported reject до READY; no new resampler |
| A6 | Startup budget не убивает READY child | 05 | W05 actual >10s survival, L01 Built-in во время запуска |
| A7 | Cancel/pipes до Run wait; единственный cmd.Wait | 06 | L06 signal/update/error path, W06 process lifecycle |
| A8 | Codec отдельно, первая интеграция как пара без old wire/timers | 05 | 02 C checks; 04 H checks; W01–W10 exact T/C receipt, 06 без first-PCM/preflight; local tests в каждом stage |
| A9 | Software/hardware acceptance разделены | 07 | S10 software verdict; hardware/vendor Echo/acoustic НЕ ВЫПОЛНЕНЫ, explicit separate instruction |

## Контроль отсутствия запрещённых механизмов

Проверяется **новое дерево** относительно T01/C01. Физическое удаление требуется лишь при доказанном наличии относящегося к ASP механизма в принятой базе; baseline review 01 не нашёл старую ASP систему в чистом upstream. Не удалять unrelated stock service restart, listener channels, WebRTC helper или Panel diagnostics.

| Механизм | Этап первого контроля | Финальная проверка / evidence |
|---|---|---|
| CFG2/TLV, HELLO/ACK2/INP2/RSP2, compatibility readers/registry | 02 | C06/H08/W10/S09; production diff и negative decode |
| Per-frame headers/sequence/origin/checksum/timestamps/status | 04 | H04/W02/W03/S02; точные pipe bytes |
| NotReady/State/Event, first-PCM/500ms/25ms конфликтующие timers | 04 | H03/W04/W05/L01/S02 |
| Three-frame preflight, confirmation/Confirmed/Pending/Keep, delayed persistence/config copies | 06 | L01/L09/S05; только Selected/Active |
| H/E/status messages, feedback/log queue/suppression/repetition counters/routers/formatters | 06 | L09/H08/S09; три новых ASP event types |
| ASP Debug + callbacks/entities/tests, diagnostic replacements | 06 | L09/S05/S09; HA inventory |
| Generic host platform whitelist/MIC4-REF2-96k admission | 03 | N02/S09; helper compatibility отдельно |
| Lifetime inode identity cache, duplicate validators/wrappers/test-only production entrypoints | 05 | W07/W10/S09; manual replacement и trust boundary review |
| Production synthetic DSP adapter | 02 | C06/H06/W10/S09; test-only files/build bindings |
| Async PCM worker/queue, auto retry/recovery/new lifecycle supervisor | 05 | W03/L02/L05/L06/S09; existing stock supervision сохранён |
| Generation/age tracking, descriptor revalidation/renegotiation Rewire | 06 | L03/L05/S09 |
| Второй privacy/output gate, epochs/masking, mute FIR/DSP/history reset | 03 | N06/L08/S03 |
| Новые resamplers/metadata providers/универсальный DSP platform dispatch | 03 | N07/H01/H02/S09 |
| Installer/OTA/download/release automation/vendor proprietary assets | 04 | H08/S09, new C file inventory/provenance |
| Legacy-only tests (status/preflight/old deadline/AST matrices) | 02 | Все stage diffs; 07 S09, без удаления unrelated stock tests |

## Практическая проверка полноты

Всего **58 D + 5 F + 27 V + 9 A = 99 адресуемых требований**. Для каждого есть owner и наблюдаемый критерий; отсутствие hardware evidence не скрывается software PASS.

Code executor заполняет только относящиеся к своему stage строки evidence в handoff. 07 прикладывает общий verdict по 99 IDs и exact final T/C pair. «Проверено grep» само по себе недостаточно для concurrency, secure launch, privacy или wire — нужны назначенные behavioral tests.

[План и зависимости](DECOMPOSITION_PLAN.md) · [Межэтапный контракт](MASTER_CONTRACT.md) · [Формат передачи](HANDOFF_TEMPLATE.md).
