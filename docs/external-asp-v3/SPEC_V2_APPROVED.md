# TECHO5 / FrankenASP — техническое задание на финальный рефакторинг External ASP

**Статус:** V3 RESTART CORRECTED DRAFT — ожидает подтверждения владельцем; старый AUDIT_V2_FINAL не утверждает raw-IDME поправку; ранее V2 CANDIDATE — девять замечаний аудитора V1 внесены; окончательное утверждение ещё требуется  
**Дата:** 30 октября 2026 года  
**Основание:** 58 архитектурных решений, пять финальных уточнений и девять замечаний аудитора к V1  
**Область:** экспериментальная ветка TECHO5 с External ASP и сопряжённый FrankenASP  
**Режим текущего этапа:** подготовка ТЗ, без изменения исходного кода

---

# 1. Цель работы

Выполнить радикальное упрощение экспериментальной интеграции External ASP в TECHO5 при сохранении всех согласованных функциональных возможностей.

Основная цель — не косметическое сокращение LOC, а **устранение избыточных архитектурных механизмов**, возникших при экспериментальной разработке.

В результате должен остаться минимальный универсальный интерфейс между существующим аудиотрактом TECHO5 и внешним DSP-процессом.

Рефакторинг должен сделать изменение пригодным для upstream PR без превращения TECHO5 в универсальный DSP framework.

## 1.1. Приоритеты

В порядке убывания:

1. Сохранение штатной функциональности Built-in и необходимых гарантий External ASP.
2. Удаление избыточных подсистем и control-plane complexity.
3. Минимизация архитектурных изменений существующего TECHO5.
4. Простота сопровождения и проверки кода.
5. Сокращение production LOC.
6. Производительность при отсутствии лишних копирований, очередей и потоков исполнения.

Сокращение LOC не является оправданием удаления необходимой синхронизации, защиты запуска или гарантии стабильности опубликованного PCM.

## 1.2. Критерий архитектурного результата

Целевая схема:

**TECHO5**

- Выполняет штатный hardware capture.
- Предоставляет фактический native PCM и descriptor.
- Сохраняет штатную обработку Built-in.
- Управляет пользовательским выбором ASP.
- Запускает и завершает внешний helper.
- Выполняет синхронный IPC.
- Выбирает итоговый mono PCM.
- Публикует его существующим подписчикам.

**External ASP helper**

- Проверяет совместимость capture с собственным DSP.
- Выполняет mapping MIC/REF.
- Преобразует частоту и PCM formats.
- Согласует размеры квантов.
- Управляет собственными FIFO.
- Использует доступные metadata и calibration.
- Выполняет DSP.
- Возвращает строго определённый mono PCM.

**FrankenASP** — конкретная реализация External ASP helper для соответствующего vendor DSP.

Ни одна DSP-специфическая особенность FrankenASP не должна становиться обязательным ограничением generic host TECHO5.

---

# 2. Границы рефакторинга

## 2.1. Включено

- Generic External ASP integration в TECHO5.
- Native PCM tap.
- Descriptor и IPC между Go host и C helper.
- Синхронная обработка PCM.
- SoftwareCut.
- Управление выбором и активностью ASP.
- Fallback.
- Lifecycle и cleanup.
- Secure launch.
- Удаление старой диагностической и протокольной инфраструктуры.
- Согласованные изменения FrankenASP.
- Замена и сокращение тестов.
- Проверка совместимости со штатным TECHO5.

## 2.2. Исключено

- Изменения installer и OTA extraction.
- Изменения vendor runtime assets.
- Release automation.
- Автоматическое восстановление DSP после отказа.
- Асинхронная аудиоочередь.
- Изменение штатных DSP-стадий.
- Оптимизация CPU путём отключения Built-in.
- Переработка Android Panel.
- Новая External-реализация для Spot.
- Spot ALSA codec mute.
- Новая система телеметрии или DSP diagnostics.
- Совместимость с прежним неопубликованным IPC-протоколом.

Работа ограничена рефакторингом существующей экспериментальной интеграции.

---

# 3. Требуемая архитектура TECHO5

## 3.1. Единственная точка интеграции

Использовать существующий `mic.Source`.

Не создавать:

- Дополнительный capture owner.
- Отдельный audio supervisor.
- Независимый External audio pipeline.
- Вторую систему mono subscriptions.
- Отдельные источники для MWW, STT или history.

Сохранить существующий механизм доставки PCM.

## 3.2. Единая точка выбора output

На каждом обрабатываемом блоке возможны два кандидата:

- Stock Built-in mono PCM.
- Mono PCM, возвращённый External helper.

Если External выбран, текущая сессия действительна и helper вернул корректный полный блок, используется его результат.

Во всех остальных случаях используется Built-in.

Выбор производится до общей точки публикации.

History, MWW, STT, spectrum и остальные mono-подписчики должны видеть один согласованный поток.

## 3.3. Stock DSP

Полностью сохранить существующие стадии:

- Mixer.
- AEC.
- Denoise.
- Leveling.
- Другие штатные преобразования, присутствующие в соответствующей платформенной реализации.

Существующий stock conversion продолжает выполняться даже при активном External, если он необходим для Built-in, fallback и штатных сенсоров.

Не заменять stock DSP внешним процессом.

## 3.4. Сенсоры

Сохранить существующие источники данных для:

- Noise.
- Facing.
- Standout.
- Level measurements.
- Остальных stock audio sensors.

Не переводить их на обработанный External mono PCM без отдельного решения.

## 3.5. Raw subscribers и телефонный тракт

Сохранить существующее поведение `ListenRaw()`.

Не вводить специальную обработку External ASP для SIP/intercom.

Существующий WebRTC AEC switching остаётся частью Built-in.

---

# 4. Native PCM и граница ответственности

## 4.1. Требование

TECHO5 должен передавать helper исходный capture PCM без DSP-specific преобразований (исключение — согласованный SoftwareCut).

Host предоставляет фактическое представление данных, а не подготавливает их под конкретный DSP.

В частности, host не должен:

- Переставлять MIC/REF ради FrankenASP.
- Приводить вход к MIC4/REF2.
- Ресемплировать native input к 96 kHz или 16 kHz ради helper.
- Преобразовывать S24_3LE в S32 только ради helper.
- Разделять interleaved PCM на MIC и REF buffers.
- Добавлять DSP-specific FIFO.
- Согласовывать vendor 8-ms quantum.

Все перечисленные функции при необходимости реализуются внутри helper.

## 4.2. Native tap

Для backends, уже возвращающих подходящий native PCM, использовать имеющиеся buffers напрямую.

Для backend с внутренней конверсией, включая `pairedCapture.Read()`, добавить минимальную возможность получить native representation до штатного `convert()`.

Предпочтительный вариант — необязательный callback/API уровня:

- `NativeFormat()`
- `ReadWithNative(...)`

Конкретные имена не обязательны.

Функциональные требования:

1. Native PCM доступен до stock conversion.
2. Существующий `Read()` сохраняет прежний контракт.
3. Built-in не зависит от наличия External ASP.
4. Никакая специальная DSP-логика не переносится в capture backend.
5. Для платформ без необходимости native tap дополнительная обвязка не создаётся.

Не требуется переписывать все capture backends.

## 4.3. Входные buffers

Допустимо передавать native PCM как borrowed buffers.

Срок их действительности — до следующего capture read.

Синхронный IPC обязан завершить использование данных до освобождения или повторного использования соответствующих buffers.

Не добавлять дополнительные копии или собственные buffer pools без необходимости.

## 4.4. Metadata

TECHO5 сохраняет ответственность за:

- Hardware capture configuration.
- Получение фактических capture properties.
- Чтение исходных, ненормализованных заводских IDME/metadata bytes без их семантического преобразования для External.
- Передачу реально доступных hardware/IDME metadata byte-for-byte; без Go parseIdme/layout.Idme-derived values, Q30, masks, confidence, availability, MIC-slot gain association или host platform qualification.
- Stock conversion и Built-in processing.

Helper отвечает за интерпретацию этих данных применительно к своему DSP.

Недоступные metadata обозначаются отсутствующими.

Не подменять отсутствующие metadata вымышленными platform defaults.

## 4.5. Platform admission

Полностью удалить generic host restrictions вида:

- `layout.Board == "athena"`.
- Только MIC4/REF2.
- Только 96 kHz.
- Только заранее перечисленные модели Echo.

Host может проверять корректность собственного descriptor и границы buffers, но не должен решать, подходит ли устройство конкретному vendor DSP.

Оценка совместимости — задача helper.

---

# 5. Новый бинарный IPC-контракт

## 5.1. Общий принцип

Заменить текущие CFG2/TLV и message envelopes одним компактным бинарным протоколом.

Последовательность:

1. Host запускает helper.
2. Host передаёт descriptor.
3. Helper проверяет descriptor и инициализирует DSP.
4. Helper возвращает `READY`.
5. Host передаёт native PCM block.
6. Helper возвращает mono PCM block.
7. Пункты 5–6 повторяются синхронно до завершения сессии.

В каждый момент допускается только один незавершённый PCM exchange.

### 5.2. Транспорт

**FD3 и FD4 не являются PCM pipes.** Сохранить существующий secure launch: FD3 закрепляет executable, FD4 закрепляет каталог assets. Не переиспользовать их для аудио.

PCM IPC — прежние stdin/stdout pipes дочернего процесса. В FrankenASP PCM идёт через выделенный wire FD, тогда как обычный stdout runtime перенаправляется в stderr. Сохранить эту изоляцию; stdout логов не должен загрязнять PCM. Новый diagnostic channel не нужен.

## 5.3. Descriptor v1

**Предлагаемое обязательное представление:** компактный little-endian binary descriptor, один раз на сессию.

Не использовать сериализацию Go/C structs посредством прямого копирования памяти.

Все multi-byte integer fields передаются явно в little-endian.

Descriptor состоит из:

1. Fixed header.
2. Buffer records.
3. Channel records.
4. Необязательного platform identifier.
5. Необязательных существующих hardware metadata.
6. Необязательных factory calibration bytes.

Здесь нет TLV registry, набора расширяемых message types или negotiation framework.

### 5.3.1. Fixed header — 24 bytes

| Offset | Field | Type |
|---|---|---|
| 0 | Magic `ASP1` | 4 bytes |
| 4 | Version = 1 | uint16 |
| 6 | Buffer count | uint8 |
| 7 | Channel count | uint8 |
| 8 | Total descriptor size | uint32 |
| 12 | Total input block size | uint32 |
| 16 | Platform info size | uint16 |
| 18 | Hardware metadata size | uint16 |
| 20 | Calibration size | uint32 |

Magic используется только при инициализации descriptor, никогда в PCM frames. Значения размеров, записей и max limits полностью определены в §5.3.2–5.3.6.

##### 5.3.2. Buffer record — 28 bytes

Каждый record описывает один фактически передаваемый физический PCM buffer.

| Offset | Field | Type | Семантика |
|---|---|---|---|
| 0 | Offset | uint32 | Смещение в логическом input block |
| 4 | Length | uint32 | Число реально отправляемых байтов buffer |
| 8 | Sample rate | uint32 | Частота исходного PCM |
| 12 | Frame count | uint32 | Число sample frames |
| 16 | Frame stride | uint32 | Расстояние между соседними sample frames (bytes) |
| 20 | PCM format | uint8 | 1=S16LE; 2=packed S24_3LE; 3=S32LE |
| 21 | Valid bits | uint8 | Значащие биты, **включая бит знака** |
| 22 | LSB shift | uint8 | Число младших незначащих бит до первого значащего |
| 23 | Channels | uint8 | Число каналов данного buffer |
| 24 | First channel | uint16 | Индекс первой channel record |
| 26 | Reserved | uint16 | 0 |

Signed PCM — two's complement, little-endian. Ширина хранилища 16/24/32 соответственно. `Valid bits` — 2…storage bits. `LSB shift` — 0…(storage bits − Valid bits): значение занимает биты [LSB shift, LSB shift + Valid bits − 1]. Например, S32LE с 24 valid bits имеет shift 0 при младшем выравнивании или shift 8 при старшем; S24_3LE/24 valid bits — shift 0. Никаких DSP-specific преобразований в host.

Для принятой **последовательной** записи buffers: `Offset[0]=0`; `Offset[i] = Offset[i−1] + Length[i−1]`; `Total input block size = Σ Length[i]`. Иначе descriptor ссылался бы на байты, которых нет в pipe. Все суммы проверяются на переполнение.

Для каждого канала проверять не только `Frame byte offset + sample width ≤ Frame stride`, но и **`(Frame count−1) × Frame stride + Frame byte offset + sample width ≤ Length`** без переполнения. Это гарантирует размещение последнего sample последнего frame в реально отправляемых bytes.

### 5.3.3. Channel record — 8 bytes

| Offset | Field | Type | Семантика |
|---|---|---|---|
| 0 | Buffer index | uint8 | Индекс buffer: 0…Buffer count−1 |
| 1 | Role | uint8 | 1=MIC; 2=REF |
| 2 | Role index | uint16 | captured slot index внутри роли |
| 4 | Frame byte offset | uint32 | Смещение sample в пределах frame |

`First channel[0]=0`; `First channel[i] = First channel[i−1] + Channels[i−1]`. Диапазон channel records каждого buffer непрерывен, каждая его запись имеет `Buffer index=i`. Сумма `Channels[i]` равна `Channel count`, все записи покрыты ровно один раз.

`Role index` уникален и непрерывен **от 0 до N−1 внутри каждого Role** по всем buffers; индексы MIC независимы от REF. Это captured channel slots, не предполагаемая физическая геометрия. Как минимум один MIC обязателен; отсутствие REF выражается отсутствием REF channel records. Factory gain соотносится именно с captured MIC Role index.

### 5.3.4. Переменные секции — raw hardware/IDME data

Здесь сохраняется только семантика универсального интерфейса: по возможности передать helper исходные, неизменённые bytes заводских источников, а не результаты их декодирования в Go. Подготовка DSP-ориентированных коэффициентов и квалификация совместимости в TECHO5 запрещены.

- Если доступны идентификаторы аппаратной платформы или заводские IDME/metadata, передавать оригинальные bytes ровно в том виде, в каком их выдаёт исходный источник, сохраняя происхождение/имя источника лишь настолько, насколько необходимо для однозначного различения raw fields. Не использовать нормализованные результаты layout.Idme()/parseIdme() вместо raw values.
- Нельзя в TECHO5 вычислять Q30 gains, MIC gain masks, availability/confidence, factory completeness, captured MIC slot association, производные платформенные identity или выполнять qualification для FrankenASP.
- Недоступные источники отсутствуют. Нельзя заполнять пробелы синтетическими заводскими данными или выводить «raw» из уже разобранного значения.
- Источник чтения и его raw bytes должен быть проверяемым; если исходные bytes недоступны через текущий stock API, stage исполнитель выбирает минимальный способ безопасно их прочесть без изменения штатной семантики Built-in.
- FrankenASP самостоятельно распознаёт источники, разбирает IDME/calibration, сопоставляет MIC/REF и определяет пригодность для vendor DSP. Неподдерживаемый raw value — отказ до READY.

Поля длин descriptor позволяют передать raw metadata, но **конкретное минимальное Go/C кодирование этих сырых значений и startup handshake проектирует исполнитель Stage02**, совместно с C decoder. Этот документ не вводит новый schema, кодировки, поля gain/identity, протокол negotiation или дополнительные управляющие сообщения. Если существующая таблица header требует корректировки ради raw byte contract, согласовать её в Stage02 в рамках минимального решения, с независимыми fixtures и обратной ссылкой на это требование.

### 5.3.5. Формирование input block

Host последовательно пишет **список из Buffer count buffers** в порядке records. Для одного interleaved MIC+REF buffer — исходные bytes; для раздельных MIC/REF — bytes каждого вслед за предыдущим. Не собирать их в новый объединённый buffer без необходимости. Между ними нет headers и пропусков. Смещения обязаны соответствовать §5.3.2.

### 5.3.6. Ограничения descriptor

- Version и reserved fields должны быть определены единым проверенным Go/C codec Stage02; корректные counts, sizes, роли и channel ranges.
- `Buffer count` представлен uint8: 1…255, **не ограничивать универсальный формат двумя buffers**. Реальные начальные backends могут предоставлять один или два; количество не является DSP-admission rule host.
- `Channel count` представлен uint8, как минимум 1; сумма Channels всех buffers равна count.
- Проверять offsets, channel layouts, границы **последнего sample**, packed PCM storage, `Valid bits` и `LSB shift`, переполнения и lengths перед доступом к буферам.
- Для минимального v1 сохранить конечный обоснованный предел descriptor; прежние оценки по нормализованным platform info (36 bytes) и factory gain (144 bytes) **больше не доказывают достаточность 8192 bytes для raw IDME**. Исполнитель Stage02 проверяет реальные raw sources и выбирает bounded limit с независимыми fixtures. Этот лимит не является пределом PCM input и не обязывает C helper выделять максимум заранее.
- Этот wire-limit ограничивает общий размер представления, а **не обязывает FrankenASP принимать все валидные комбинации**. Поддержку DSP geometry, sample rate, formats и calibration решает helper; неподдерживаемый descriptor может быть корректно отклонён до READY.

Все проверки ограничены необходимой корректностью bytes и памяти; новый generic validation framework не создавать.

### 5.3.7. Обмен инициализации

После успешной проверки конфигурации и инициализации helper однократно сообщает о готовности к обмену PCM. Конкретный **минимальный Go/C формат startup readiness**, границу сообщения и точные bytes совместно определяет исполнитель Stage02. Здесь фиксируется только его смысл: host не начинает exchange до подтверждённой готовности.

Не вводить дополнительные статусы, negotiation, READY payload framework или повторные handshakes.

Если helper не может работать:

- Он может написать понятную причину в stderr.
- Завершает процесс или закрывает соответствующий IPC channel.
- Host фиксирует неуспешный запуск и остаётся в Built-in.

Не вводить отдельные ответы `INCOMPATIBLE`, `NOT_READY`, `ERROR`, `STATE` и подобные.

### 5.3.8. Неподвижность descriptor

Descriptor относится к одной External-сессии.

Он не переотправляется после обычного `Rewire()`.

Не добавлять generation, descriptor hash, compatibility renegotiation или проверку неизменности capture geometry после штатного reopen.

Новая ручная активация создаёт новую сессию и новый descriptor.

---

# 6. PCM exchange

## 6.1. Input

Каждый PCM input block имеет точный размер, установленный descriptor при создании сессии.

Размер не изменяется в пределах сессии.

Host не обязан приводить native PCM к mono, 16 kHz или другой DSP-specific geometry.

Для согласованного 20-ms host exchange источник предоставляет соответствующий capture block. Если конкретный backend не обеспечивает совместимую временную геометрию, её поддержка оценивается helper; в generic host не вводится отдельный преобразователь квантов.

## 6.2. Output

**Единый обязательный контракт External ASP:**

- Mono.
- Signed 16-bit PCM.
- Little-endian.
- 16 000 samples/s.
- 320 samples.
- 640 bytes.
- Продолжительность — 20 ms.

Каждый успешный exchange возвращает ровно один такой блок.

Данный контракт одинаков для всех External helpers.

## 6.3. Удалить per-frame framing

В PCM block отсутствуют:

- Header.
- Sequence number.
- Sample origin.
- Block counter.
- Magic.
- Checksum.
- Timestamp.
- Status.
- Event.
- Размер, передаваемый на каждый frame.

Host знает размеры из descriptor и констант output contract.

## 6.4. Синхронный exchange

На один блок выполняется:

1. Полная запись входного PCM.
2. Полное чтение 640 bytes результата.
3. Проверка успешного завершения.
4. Проверка принадлежности ответа текущей сессии.
5. Использование результата при допустимом текущем режиме.

Использовать эквивалент `writeFull` и `readFull`.

Partial writes/reads допустимы как промежуточное состояние системного вызова, но не как успешный exchange.

EOF, broken pipe, timeout и неполный блок означают ошибку.

## 6.5. Deadline

PCM exchange имеет один общий deadline:

**5 секунд.**

Deadline покрывает полный write/read exchange, а не каждую операцию отдельно.

Это аварийное ограничение, а не целевая latency.

После deadline host выполняет fallback.

Не создавать:

- 25-ms deadline.
- Несколько повторных попыток.
- Retry queue.
- Отдельный exchange worker.
- Frame age tracking.

## 6.6. Initialization deadline

Ввести один конечный deadline на весь запуск до READY.

Для ТЗ принять `10 seconds` как единый **только startup** deadline (проектное значение без подтверждения измерениями). Он покрывает подготовку запуска, передачу descriptor и получение READY, но не жизнь успешно запущенного процесса.

После READY **снять/завершить startup deadline**, не отменяя общий context текущей External-сессии. Нельзя применять истекающий через 10 секунд context к `exec.CommandContext` так, чтобы после READY он уничтожил работающий helper. Реальная отмена сессии и 5-секундный PCM deadline остаются независимыми.

Инициализация helper остаётся **вне capture loop**, поэтому Built-in продолжает обработку во время ожидания READY. Удалить HELLO 500 ms и first-PCM 1 s.

Обязательная регрессия: успешная сессия продолжает работать более 10 секунд с момента начала запуска.

---

# 7. FrankenASP — обязательный рефакторинг

## 7.1. Новый descriptor reader

Удалить поддержку старого CFG2/TLV протокола.

Реализовать чтение нового descriptor v1. **Парсинг и DSP compatibility — разные требования:** корректно распознавать wire представления, но не обещать DSP processing всех допустимых конфигураций.

Descriptor может описывать:

- Single interleaved MIC/REF buffer.
- Separate MIC/REF buffers.
- Случай без REF.
- S16LE.
- S24_3LE.
- S32LE.
- Valid bits и alignment.
- Фактический channel order.
- Sample rates.
- Frame geometry.
- Доступные raw hardware/IDME bytes без host-side gain/platform normalization.

Проверять совместимость с конкретным DSP внутри FrankenASP; переданные исходные hardware/IDME bytes интерпретирует исключительно helper. **Сохранять уже подтверждённые рабочие конфигурации; неподдерживаемые частоты, layouts, форматы, отсутствие REF и временная геометрия отвергаются до READY.** Разработка новых resamplers и расширение vendor DSP compatibility не являются задачей рефакторинга.

## 7.2. DSP adaptation

В FrankenASP остаются или переносятся:

- MIC/REF mapping.
- Channel reordering.
- DSP-specific sample format conversion.
- Существующий DSP-specific resampling для уже поддерживаемых конфигураций (без обязательной разработки новых resamplers).
- Существующее DSP input quantum conversion.
- Vendor-specific FIFO.
- Calibration interpretation.
- Vendor runtime qualification.

Не предполагать, что TECHO5 подготовил данные под MIC4/REF2/96 kHz.

## 7.3. Ровно один output на каждый exchange

После READY FrankenASP обязан возвращать полный 640-byte PCM block.

Не передавать промежуточные статусы, включая `NotReady`.

## 7.4. Initial FIFO padding

При создании новой vendor DSP-сессии:

1. Инициализировать обычный output FIFO.
2. Один раз добавить в него 64 нулевых samples.
3. Выполнять штатные vendor DSP calls.
4. Возвращать по 320 samples на каждый host exchange.

Padding применяется **только к output FIFO**.

Не добавлять 64 samples во входные MIC/REF streams.

Не повторять padding:

- После mute.
- После unmute.
- После обычного Rewire.
- После ALSA XRUN.

При новой настоящей External DSP-сессии FIFO инициализируется заново.

## 7.5. Удаляемый код FrankenASP

Удалить:

- CFG2/TLV parser.
- Старые wire envelopes.
- Обработку `HELLO`.
- `ACK2`, `INP2`, `RSP2`.
- `NotReady` response.
- Wire event/status messages.
- H:/E: diagnostic codes собственной интеграции.
- Per-frame sequence и origin.
- Устаревшие startup confirmation branches.
- Обвязку, существующую исключительно ради старого протокола.

Сохранить vendor DSP error handling там, где он необходим внутренней реализации, но не превращать его в новый IPC status protocol.

---

# 8. SoftwareCut

## 8.1. Единая политика

TECHO5 остаётся единственным владельцем software mute policy.

Не добавлять privacy state machine в FrankenASP.

На каждый capture block значение `privacy.SoftwareCut()` получается один раз.

## 8.2. Раннее обнуление native PCM

При активном SoftwareCut:

- Native MIC samples должны быть обнулены до передачи helper.
- Native REF samples должны быть обнулены в соответствии с согласованной software mute semantics.
- Для interleaved buffer обнуляются нужные samples по channel layout.
- Обнуление выполняется с учётом PCM packing, не повреждая границы samples.

Если capture buffer можно безопасно изменить на месте, использовать in-place zeroing.

Не создавать отдельный copied native block без необходимости.

## 8.3. Stock path

То же единожды полученное значение SoftwareCut применяется к штатному пути.

Сохранить существующую гарантию заглушения stock MIC/REF.

Допускается, что после раннего обнуления:

- Stock resampler получает нули.
- Состояние его фильтров меняется.
- Уровнемеры внутри conversion показывают заглушённый вход.
- Переходные samples при mute/unmute отличаются от прежней версии.

Это не считается самостоятельной регрессией.

## 8.4. Аппаратный mute

Не менять существующее аппаратное mute behavior.

В частности, при hardware mute REF может оставаться ненулевым, если так работает устройство.

Не добавлять новые hardware controls.

## 8.5. Запрещённые компенсации

Не добавлять:

- Повторный SoftwareCut после External DSP.
- Output masking.
- Privacy epochs.
- Принудительную очистку DSP history.
- Reset resampler/FIR.
- Специальные mute/unmute transitions.
- Дополнительную проверку перед MWW/STT/history.

Допустимый кратковременный хвост stateful DSP не является основанием для новой системы приватности.

---

# 9. Владение выходным PCM

## 9.1. Разделение ownership

Native input и опубликованный mono output имеют разные требования.

**Native input:** borrowed, действителен до следующего capture read.

**Published output:** неизменен в течение всего времени использования подписчиками.

## 9.2. Обязательная гарантия

Ни последующий capture read, ни следующий IPC exchange не могут перезаписывать PCM frame, уже переданный подписчикам.

Не публиковать напрямую переиспользуемый helper scratch buffer.

Использовать существующий механизм безопасной передачи output либо минимальную необходимую копию.

## 9.3. Недопустимые оптимизации

Не вводить ради этого:

- Новый output buffer pool.
- Audio queue.
- Reference-counted buffer framework.
- Несколько промежуточных copies.
- Новый ownership manager.

Если существующая публикация уже гарантирует стабильность кадра, сохранить её.

## 9.4. Обязательная регрессия

Тест:

1. Подписчик получает PCM frame A.
2. Удерживает ссылку на его данные.
3. Host выполняет несколько новых capture reads.
4. Выполняет следующие External exchanges.
5. Публикует frames B, C и далее.
6. Сравнивает содержимое удерживаемого A с исходным snapshot.

Frame A обязан остаться неизменным.

Тот же принцип проверить при fallback и переключениях режимов.

---

# 10. Управление ASP: Selected и Active

## 10.1. Только два смысловых состояния

Оставить:

**Selected** — режим, выбранный пользователем.

**Active** — фактически используемый источник итогового mono PCM.

Не добавлять отдельные:

- Desired.
- Confirmed.
- Pending.
- Keep.
- Ready state.
- Recovery state.
- LastGood.
- Health status.

Операционные поля, необходимые для владения текущим процессом и синхронизации, не должны превращаться в новую публичную state machine.

## 10.2. Persistence

При выборе External или Built-in:

- Selected меняется немедленно.
- Persisted configuration обновляется немедленно.
- Сохранение не зависит от READY.
- Сохранение не зависит от успешной обработки PCM.

При runtime fallback Selected остаётся External, Active становится Built-in.

## 10.3. Activation

Первый полностью успешный PCM exchange может активировать External.

Трёхкадровый preflight запрещён.

Не требовать отдельных confirmations.

До первого успешного exchange Active остаётся Built-in.

## 10.4. Helper unavailable

Отсутствие helper при старте echod не скрывает опцию External в Home Assistant.

Пользователь может выбрать External.

Если executable отсутствует или запуск неудачен:

- Selected сохраняется External.
- Active остаётся Built-in.
- Регистрируется fallback с причиной.

После установки helper допускается ручное переключение Built-in → External без restart echod.

Автоматический поиск по таймеру или auto retry не требуется.

---

# 11. Lifecycle и завершение helper

## 11.1. Владелец

Helper принадлежит External ASP session, управляемой существующим аудиокомпонентом echod.

Он **не принадлежит отдельному `Source.Run()`**.

## 11.2. Удалить старые связи

Удалить:

- `defer s.closeASP()` из `Source.Run()`.
- `s.closeASP()` из `Source.Close()`.

`Source.Close()` продолжает выполнять своё штатное назначение — закрытие capture.

## 11.3. Rewire

При обычном `Rewire()`:

- Capture закрывается.
- Capture переоткрывается.
- Helper сохраняется.
- IPC session сохраняется.
- Новый descriptor не отправляется.
- READY не запрашивается повторно.
- FIFO padding не повторяется.
- DSP не переинициализируется.

## 11.4. XRUN

Обычный ALSA overrun не является отказом External ASP.

Сохранить stock handling overrun и продолжение capture.

Не завершать helper и не перезапускать DSP.

## 11.5. Настоящие причины завершения

Helper завершается при:

1. Явном переключении External → Built-in.
2. Реальном отказе launch или IPC.
3. Завершении echod.

При этом необходимо:

- Закрыть IPC resources.
- Завершить child process.
- Выполнить `Wait`.
- Исключить zombie.
- Исключить утечки file descriptors.

## 11.6. Окончательное место cleanup

Использовать существующий долгоживущий shutdown аудиокомпонента echod, владеющего External-сессией; **не** `Source.Close()`, вызываемый и при штатном `Rewire()`. Не создавать нового supervisor.

**Порядок обязателен:**
1. Запретить запуск новых External/capture операций.
2. **Немедленно отменить** текущие launch/PCM exchanges и закрыть соответствующие pipes, чтобы разблокировать ожидающие I/O.
3. Затем дождаться выхода audio processing/Run, не блокируя отмену его IPC.
4. Завершить helper, освободить FD и дождаться завершения его process-reaper.

Для `cmd.Wait()` должен существовать **ровно один владелец**: shutdown ожидает результат уже выполняемого Wait, не вызывает второй Wait. Если нет подходящего shutdown hook, добавить один прямой вызов в существующий верхнеуровневый teardown, не перенося cleanup обратно в capture-only Close.

# 12. Отмена устаревших операций

## 12.1. Обязательная конкурентная гарантия

При переключении External → Built-in во время запуска или PCM exchange прежняя операция не должна:

- Повторно активировать External.
- Публиковать устаревший PCM.
- Заменить новую External-сессию.
- Изменить Active после отмены.
- Закрыть более новую сессию при позднем cleanup.

## 12.2. Сохранить минимальную синхронизацию

Необходимо сохранить:

- Cancellation текущего запуска.
- Cancellation текущего IPC exchange.
- Проверку identity завершившегося запуска.
- Проверку identity PCM response.
- Проверку актуального выбранного режима.
- Минимальное существующее locking.

## 12.3. Правило публикации

Перед принятием External response проверить, что:

- Session object всё ещё является текущим.
- Выбранный режим всё ещё External.
- Сессия не отменена.
- Ответ принадлежит текущему helper.
- Полный PCM block получен успешно.

Проверка и окончательное решение о публикации должны быть синхронизированы с переключением режима так, чтобы отменённый результат не мог пройти между проверкой и публикацией.

Использовать существующий механизм синхронизации, расширяя его только в пределах этой гарантии.

Нельзя удерживать общий audio publication mutex во время blocking IPC.

## 12.4. Не добавлять

- Generation counters.
- Stale-result age tracking.
- Audio queues.
- Новый session lifecycle manager.
- Повторную обработку отменённых PCM blocks.
- Retry после cancellation.

Session identity и отмена достаточны.

---

# 13. Failure и fallback

## 13.1. Ошибки

Основания для fallback:

- Executable отсутствует.
- Secure launch validation failed.
- Child process не запустился.
- Helper завершился до READY.
- Initialization deadline.
- Broken pipe.
- EOF.
- Short PCM response.
- PCM exchange timeout.
- Process termination во время работы.
- Другая реальная ошибка IPC.

Не вводить отдельные категории ошибок для каждой стадии обработки.

## 13.2. Последствия

При отказе:

1. Проверить, что отказ относится к текущей сессии.
2. Active переводится на Built-in.
3. Selected не меняется.
4. Ошибочная сессия закрывается.
5. Helper освобождается.
6. В host log фиксируется fallback с понятной причиной.

Не запускать helper повторно автоматически.

## 13.3. Ручной retry

Пользователь выполняет Built-in → External.

Новая активация:

- Повторно обнаруживает helper.
- Повторно проверяет executable.
- Читает актуальную конфигурацию.
- Создаёт новую сессию.
- Передаёт новый descriptor.

Не сохранять старый helper identity между независимыми ручными активациями.

---

# 14. Secure launch

## 14.1. Сохранить

- Безопасное обнаружение executable в `/data`.
- Secure path traversal.
- Необходимые owner/mode/type checks.
- `O_NOFOLLOW`.
- FD-pinning.
- Защиту от небезопасной замены executable в процессе проверки и запуска.
- Реально необходимые file descriptor и process cleanup guarantees.

Не ослаблять границы безопасности ради сокращения LOC.

## 14.2. Удалить

- Долгоживущий inode identity cache.
- `discoveredASP = sync.OnceValue(...)`.
- Запрет на использование заменённой установки без restart echod.
- Дублирующую validation уже проверенных внутренних структур.
- Лишние launch wrappers.
- Test-only production launch entry points, не имеющие иной функции.

Проверки на реальных trust boundaries сохранить.

## 14.3. Обновление helper

После замены executable/assets новая ручная активация должна использовать новую установленную версию без restart echod.

Не добавлять watcher или автоматический hot reload.

---

# 15. Home Assistant

## 15.1. Сохранить

- Выбор Built-in / External.
- Индикацию фактически активного ASP.
- Immediate persistence Selected.
- Корректное отражение runtime fallback.

## 15.2. Полностью удалить ASP Debug

Удалить не только entity, но и всю обслуживающую цепочку:

- Entity declaration.
- Registration.
- Initialization.
- State storage.
- Value updates.
- Synchronization.
- Publication.
- Formatter functions.
- Event subscriptions.
- Tests, существующие ради этой entity.

Не оставлять неиспользуемые поля и callbacks после удаления entity.

## 15.3. Не создавать замену

Не вводить:

- Новый text sensor.
- DSP health sensor.
- Status attributes.
- Diagnostic dropdown.
- Error history.
- Отдельные события HA для ошибок DSP.

Остаются только Selected и Active.

---

# 16. Logging и удаление diagnostic infrastructure

## 16.1. Ровно три типа host log events

Разрешены:

**1. External ASP launch**

Одно событие попытки запуска с необходимым контекстом.

**2. Fallback**

Одно событие переключения на Built-in с понятной причиной.

**3. External session termination**

Одно событие окончания сессии/процесса.

Не требуются отдельные сообщения READY, подтверждения первых кадров, health или успешности каждого exchange.

## 16.2. Удаляемая инфраструктура

Полностью удалить собственные:

- Numeric H:/E: codes.
- Status codes.
- Status messages.
- Origin/status structures.
- `asp_feedback`.
- `asp_log`.
- Custom log queues.
- Message suppression.
- Repetition counters.
- Diagnostic event routers.
- Status formatters.
- Feedback delivery functions.
- Поля `aspControl`, оставшиеся без назначения.

Не оставлять прежние диагностические механизмы в виде dead code.

## 16.3. Helper stderr

Обычный stderr helper допускается.

Он не является частью PCM protocol и не требует отдельного status parser в TECHO5.

Не создавать custom helper log collection subsystem.

---

# 17. Production code cleanup

## 17.1. Главный принцип

Каждая удаляемая подсистема должна удаляться полностью: от объявления типа до последнего теста и integration callback.

Не допускается замена старого сложного механизма новым механизмом аналогичного объёма.

## 17.2. Обязательный перечень удаления

| Подсистема | Требуемое действие |
|---|---|
| CFG2/TLV | Полностью удалить |
| HELLO/ACK2/INP2/RSP2 | Полностью удалить |
| Per-frame framing | Полностью удалить |
| Status codes/messages | Полностью удалить |
| H:/E: codes | Полностью удалить |
| ASP Debug | Полностью удалить |
| Feedback/log subsystem | Полностью удалить |
| Custom event/log queues | Полностью удалить |
| Message suppression | Полностью удалить |
| Three-frame preflight | Полностью удалить |
| Delayed persistence | Полностью удалить |
| Startup confirmation transaction | Полностью удалить |
| NotReady response | Полностью удалить |
| Separate launch confirmation states | Полностью удалить |
| Platform whitelist | Полностью удалить |
| Inode identity cache | Полностью удалить |
| Production synthetic DSP adapter | Полностью удалить |
| Async PCM worker/queue, если остались | Полностью удалить |
| Ненужные production test hooks | Удалить |
| Дублирующую validation | Сократить |
| Избыточные launch wrappers | Сократить |
| Дублирующие config copies | Сократить |
| Ненужные control-plane fields | Удалить |

## 17.3. Обязательный перечень сохранения

- `mic.Source`.
- Stock capture.
- Stock DSP.
- Native PCM access.
- SoftwareCut.
- Существующая mono publication.
- Native input ownership.
- Published output stability.
- Session identity checks.
- Cancellation.
- 5-second exchange deadline.
- READY.
- Selected/Active.
- Runtime fallback.
- Rewire persistence.
- XRUN tolerance.
- Process cleanup.
- Secure launch.

## 17.4. Дополнительные ограничения

Не вводить:

- Новую config framework.
- Descriptor registry.
- Новый generic validation layer.
- Новый lifecycle framework.
- Synthetic adapter в production.
- Универсальные process wrappers ради единственного helper.
- Дополнительный platform dispatch.

Тестовые утилиты должны находиться в test code, а не расширять production API.

---

# 18. Конкретные области изменения TECHO5

Исполнителю необходимо проверить текущие symbol names и call sites экспериментального HEAD и выполнить изменения в следующих областях.

## 18.1. `mic.Source`

- Упростить External runtime integration.
- Сохранить общую точку mono publication.
- Сохранить короткую синхронизацию управления.
- Удалить ненужные control-plane fields.
- Удалить закрытие helper из capture Run/Close.
- Сохранить защиту от stale sessions.

## 18.2. Capture backends

- Минимально предоставить native PCM там, где выполняется ранняя conversion.
- Сохранить `Read()` API.
- Сохранить stock conversion.
- Не создавать новую иерархию adapters.

## 18.3. ASP process / IPC

- Новый descriptor writer.
- READY handshake.
- Fixed-size synchronous PCM exchange.
- 5-second deadline.
- Process cancellation.
- Session cleanup.
- Secure launch без inode cache.
- Удаление старого protocol/control code.

## 18.4. HA integration

- Immediate Selected persistence.
- Active update.
- Удаление ASP Debug и всей обвязки.
- Удаление feedback APIs.

## 18.5. echod lifecycle

- Подключить один final helper cleanup к существующему shutdown path.
- Не менять stock capture shutdown semantics.
- Не вводить новый supervisor.

## 18.6. FrankenASP

- Новый descriptor parser.
- Новый READY.
- Fixed-size output.
- Однократный FIFO padding.
- Сохранение DSP adaptation.
- Полное удаление старого wire/status protocol.

---

# 19. Порядок выполнения будущей реализации

Каждый этап ведётся в отдельной управляемой ревизии/ветке с немедленным обновлением затронутых тестов. Не создавать временный old/new wire compatibility layer только для удобства промежуточных коммитов. Промежуточные codec fixtures допустимы без включения в текущий runtime. Рабочий переход принимается **парой TECHO5 + FrankenASP commit IDs**, а не независимо.

## R1 — baseline
Сверить точные HEAD, зафиксировать diff против upstream, файлы, LOC и действующие тесты.

## R2 — подготовка общего codec, без переключения runtime
Разработать и независимо проверить Go writer / C reader descriptor v1, READY и тестовые fixtures. Существующий рабочий handshake до парной миграции не переключать; новые codec modules пока могут быть не подключены к runtime. Тесты codec — в этом этапе.

## R3 — единый интеграционный переход обоих репозиториев
**Атомарно по итоговой приёмке согласованной пары ревизий** перейти на descriptor/READY/fixed-size PCM, вывести `NotReady` через 64-sample output FIFO padding и удалить старые IPC envelopes и **непосредственно зависимые** поля/обработчики `ConfigStatus`, `Status`, `Start`, `NotReady`, three-frame preflight, first-PCM и конфликтующие startup timers. Host и helper не обязаны работать с прежними wire версиями. Немедленно обновить affected IPC/control tests и обеспечить компилируемый Go/C набор.

## R4 — native tap и SoftwareCut
Универсальный доступ к native PCM до stock conversion там, где это необходимо; early SoftwareCut, borrowed input и неизменность published mono output. Обновить relevant tests этого же этапа.

## R5 — lifecycle после удаления preflight timers
Отвязать helper от capture Run/Close, сохранить при Rewire/XRUN, реализовать безопасный final cleanup и cancellation/identity guarantees. Не переносить сюда старые конфликтующие таймеры; тесты lifecycle — сразу.

## R6–R7 — окончательное упрощение control plane и диагностики
Сохранить только Selected/Active, immediate persistence, убрать остаточную статусную/feedback/log queue инфраструктуру и ASP Debug со всей обвязкой. Зависимый от wire state уже удалён в R3. При необходимости объединить R6/R7 в один исполняемый этап — не допускать висячих компиляционных зависимостей. Тесты HA/логирования обновлять при удалении.

## R8 — финальные regression и cleanup
Удалить оставшиеся tests удалённых механизмов, собрать поддерживаемые targets, провести race/security/stock regressions, проверить итоговый diff, LOC и dead code. Это не замена локальным тестам R2–R7.

## R9 — программная интеграционная приёмка
Проверить согласованную Go/C пару, конформность descriptor, sync PCM, mute, cancellation, fallback, Rewire/XRUN, secure launch и stock. Подготовить отчёт и парные commit IDs. **Не запускать аппаратные тесты автоматически.** Hardware/vendor-runtime acceptance оформляется отдельно по решению владельца.

# 20. Обязательные тесты

## 20.1. Descriptor / Go↔C

Проверить:

- Независимое одинаковое декодирование Go/C валидного descriptor (не означает обязательную DSP-поддержку каждой описываемой конфигурации).
- S16LE.
- S24_3LE.
- S32LE.
- 24 valid bits inside S32.
- Left/right bit alignment.
- Interleaved MIC/REF.
- Separate MIC/REF.
- REF absent.
- Channel order.
- Buffer offsets и stride.
- Доступные metadata.
- Отсутствующие metadata.
- Raw IDME/calibration bytes: побайтовая идентичность исходному fixture; вся derived calibration только на C стороне.
- Invalid descriptor length.
- Invalid buffer ranges.
- Unsupported version.
- Отдельно: успешная обработка **поддерживаемой** DSP-конфигурации и корректный отказ **валидной, но неподдерживаемой** конфигурации до READY.
- Отдельно: отказ при несовместимой 20-ms temporal geometry, без разработки нового resampler.
- Последний sample последнего frame не выходит за Length; Offset каждого buffer соответствует записанным bytes.

Не создавать production synthetic geometry emulator.

## 20.2. PCM IPC

- Один write/read exchange.
- Ровно 640-byte output.
- Partial read/write.
- EOF.
- Short response.
- Broken pipe.
- 5-second deadline.
- Один outstanding block.
- Отсутствие frame headers.
- Корректный первый output.
- Session остаётся работоспособной после истечения 10 seconds с момента launch, если READY уже получен.

## 20.3. FrankenASP

- One-time 64-sample output padding.
- Правильная работа с 8-ms DSP quantum.
- Полный первый 320-sample output.
- Отсутствие NotReady.
- Отсутствие повторного padding после mute/unmute.
- Сохранение FIFO после Rewire.
- Корректная обработка реального PCM layout.

## 20.4. SoftwareCut

- Native MIC обнулён до helper.
- Native REF обнулён согласно software mute semantics.
- Stock получает соответствующий muted input.
- Hardware mute не изменяет REF behavior искусственно.
- Нет непрерывной передачи живого MIC при SoftwareCut.
- Нет дополнительных output privacy checks.
- Переходные samples не требуют bit-exact equivalence.

## 20.5. Output ownership

Обязательная регрессия удержания опубликованного frame во время последующих capture reads и IPC exchanges.

## 20.6. Cancellation

Отдельно проверить:

1. External → Built-in во время launch.
2. External → Built-in во время READY.
3. External → Built-in во время PCM write.
4. External → Built-in во время PCM read.
5. Старый launch завершается после переключения.
6. Старый PCM response приходит после переключения.
7. Создаётся новая External session до завершения старой.
8. Старый cleanup не затрагивает новую session.

Ни один stale result не должен активировать External или публиковаться.

## 20.7. Lifecycle

- `Source.Close()` не завершает helper.
- `Source.Run()` exit не завершает helper.
- Обычный Rewire сохраняет PID/session.
- Не происходит повторный READY.
- Обычный ALSA overrun не завершает helper.
- Реальный IPC failure завершает session.
- Echon shutdown освобождает child.
- Выполняется `Wait`.
- Не остаются zombies и открытые pipes.

## 20.8. HA

- Selected persist immediately.
- Active меняется после первого успешного exchange.
- Runtime fallback не перезаписывает Selected.
- Helper unavailable не скрывает опцию.
- Новая ручная активация перечитывает executable.
- ASP Debug отсутствует.
- Нет status/feedback публикаций.

## 20.9. Secure launch

- Valid installation запускается.
- Небезопасные symlinks отвергаются.
- Owner/mode/type checks сохранены.
- Executable pinning сохранён.
- Замена установленного helper между ручными активациями работает.
- Нет долгоживущего inode cache.

## 20.10. Stock regression

Проверить:

- Built-in без helper.
- Existing mono subscribers.
- `ListenRaw()`.
- Noise/facing/standout/levels.
- MWW.
- STT.
- History.
- SIP/intercom.
- WebRTC AEC switching.
- Existing build tags.
- Android Panel build без специальной External integration.

Для SoftwareCut переходные sample differences допускаются в пределах утверждённых исключений.

---

# 21. Требования к тестовой инфраструктуре

Удалить тесты, обслуживающие исключительно исключённые механизмы:

- Status codes.
- Feedback queues.
- ASP Debug.
- Старые protocol envelopes.
- Confirmation transactions.
- Three-frame preflight.
- Old deadline 25 ms.
- Избыточные AST/structural checks.
- Дублирующие failure matrices.

Тесты нового IPC должны опираться на наблюдаемое поведение, а не на названия внутренних методов.

Для conformance допустим минимальный test-only fake helper.

Не переносить его в production TECHO5.

---

# 22. Сборки и проверка совместимости

В будущей реализации необходимо:

1. Запустить релевантные Go unit tests.
2. Проверить Go/C protocol conformance.
3. Проверить race-sensitive tests.
4. Собрать экспериментальный TECHO5 для целевой Linux ARM32 платформы.
5. Собрать соответствующий FrankenASP.
6. Проверить существующие supported build configurations.
7. Проверить Android Panel compile path.
8. Подготовить компоненты к аппаратной проверке.

**Программная приёмка** на этом этапе: сборки, Go/C codec conformance, integration/race/security/stock regression и проверяемые Go/C pair commits. **Аппаратная приёмка** (vendor DSP runtime на Echo, акустика) выполняется отдельно по явному решению владельца проекта. До разрешения её статус — **«НЕ ВЫПОЛНЕНА»**, а не ложный FAIL программного этапа.

Не запускать новые бинарники на Echo автоматически только потому, что прошли unit tests.

---

# 23. Метрики результата

Исполнитель должен предоставить сравнение до/после.

## 23.1. Production code

- TECHO5 production LOC added.
- TECHO5 production LOC removed.
- Итоговый net diff относительно upstream.
- FrankenASP production LOC added/removed.
- Количество полностью удалённых subsystems.
- Перечень удалённых Go/C files.
- Перечень удалённых типов и control-plane fields.

Ориентир аудита по сокращению TECHO5 — порядка 1 100–1 700 production Go LOC.

Это ориентир, а не обязательный количественный лимит.

## 23.2. Architecture review

Проверить, что после удаления старых подсистем не появились новые:

- Managers.
- Registries.
- Queues.
- Retry frameworks.
- State machines.
- Diagnostic systems.
- Generic validators.

Допустимы только минимальные структуры, непосредственно необходимые для descriptor, session ownership и IPC.

---

# 24. Условия приёмки

Рефакторинг считается завершённым только при выполнении всех условий.

### Архитектура

- [ ] External остаётся необязательным расширением `mic.Source`.
- [ ] Stock DSP полностью сохранён.
- [ ] Единый mono publication path сохранён.
- [ ] Platform DSP restrictions удалены из generic host.
- [ ] Native PCM доступен без DSP-specific host conversion.

### Протокол

- [ ] Go/C используют один descriptor v1.
- [ ] READY заменяет старый handshake.
- [ ] Каждый exchange возвращает mono S16LE/16 kHz/320 samples.
- [ ] Нет NotReady.
- [ ] Нет старых wire envelopes.
- [ ] Нет per-frame headers/counters/checksums.
- [ ] Нет wire-level statuses.
- [ ] 5-second PCM deadline сохранён.

### FrankenASP

- [ ] Mapping/resampling/FIFO находятся в helper.
- [ ] Initial 64-sample padding выполняется один раз.
- [ ] Первый exchange возвращает полный frame.
- [ ] Подтверждена работоспособность **программного** Go/C integration contract на поддерживаемых fixtures (аппаратное выполнение vendor DSP — отдельная приёмка).

### Приватность и buffers

- [ ] SoftwareCut применяется к native input до helper.
- [ ] Stock mute semantics сохранены.
- [ ] Нет повторного output privacy check.
- [ ] Native borrowed buffer lifetime соблюдён.
- [ ] Published PCM остаётся неизменным для подписчиков.

### Lifecycle

- [ ] Rewire не завершает helper.
- [ ] XRUN не завершает helper.
- [ ] Отмена запуска работает.
- [ ] Отмена PCM exchange работает.
- [ ] Stale responses не публикуются.
- [ ] Final echod shutdown завершает helper и выполняет Wait.
- [ ] Auto retry отсутствует.

### Control plane

- [ ] Остались только Selected и Active.
- [ ] Selected сохраняется немедленно.
- [ ] Preflight удалён.
- [ ] Delayed persistence удалён.
- [ ] ASP Debug полностью удалён.
- [ ] Feedback/log subsystem удалён.
- [ ] Status codes/messages полностью удалены.
- [ ] Host logging ограничен тремя типами событий.

### Безопасность

- [ ] Secure launch из `/data` сохранён.
- [ ] FD-pinning сохранён.
- [ ] Inode cache удалён.
- [ ] Новая ручная активация видит обновлённый helper.

### Регрессия

- [ ] Go tests успешны.
- [ ] Go/C conformance успешен.
- [ ] Race-sensitive tests успешны.
- [ ] Built-in regression успешна.
- [ ] Supporting builds успешны.
- [ ] Нет неоправданных новых production abstractions.

### Аппаратная приёмка (отдельно, только по явному поручению)

- [ ] Проверен vendor DSP runtime на настоящем Echo — **пока НЕ ВЫПОЛНЕНА**, не является условием закрытия программного рефакторинга.
- [ ] Проведены акустические тесты — **пока НЕ ВЫПОЛНЕНЫ**, не запускать автоматически.

---

# 25. Требования к Git и результатам работы

Работы TECHO5 проводить исключительно в предназначенной для External ASP ветке или изолированном worktree, созданном от её актуального HEAD.

Не переносить экспериментальные изменения в `main` репозитория `dmit2k/techo5`.

Официальный `main` остаётся независимым от экспериментального External ASP до отдельного решения.

Сопряжённые изменения FrankenASP вести согласованно с изменениями TECHO5.

На этапе реализации исполнитель должен предоставить:

- Зафиксированные исходные commit IDs.
- Перечень изменённых файлов.
- Описание удалённых подсистем.
- Результаты тестов.
- Сравнение production LOC.
- Финальные commits соответствующих репозиториев.
- Отчёт о соблюдении 58 пунктов и пяти уточнений.
- Список реальных оставшихся ограничений, если они будут обнаружены.

Не выполнять merge в upstream TECHO5 `main` без отдельного указания.

---

# 26. Критерий окончательной оценки

Результат должен представлять собой **не упрощённую версию старой экспериментальной архитектуры, а минимальную реализацию принятого контракта**.

В TECHO5 должны остаться только:

- Получение native PCM.
- Передача фактического descriptor.
- Безопасный запуск helper.
- Синхронный PCM exchange.
- Выбор итогового mono PCM.
- Минимальная Selected/Active логика.
- Необходимые cancellation/session identity checks.
- Fallback.
- Cleanup.

Во FrankenASP должны остаться:

- Descriptor reader.
- Проверка DSP compatibility.
- Native format adaptation.
- Vendor DSP execution.
- DSP FIFO.
- Fixed-size mono PCM output.
- Минимальный process/error handling.

Всё остальное должно быть удалено, если оно не требуется существующему Built-in, другим функциям TECHO5 или конкретной сохраняемой гарантии.

**Главный критерий:** минимальный production diff, сохраняющий согласованную функциональность, а не максимальное количество реализованных защитных механизмов.

---

# 27. Принцип работы исполнителя

Архитектурные решения закрыты.

При реализации не возвращаться к обсуждению:

- Async worker вместо sync exchange.
- Status protocol.
- Diagnostic HA entities.
- Retry manager.
- Generation counters.
- Второго privacy check.
- Автоматической реинициализации после Rewire.
- Platform whitelist.
- Production synthetic adapter.
- Старой CFG2/TLV compatibility.

Если обнаружится конкретное противоречие с фактическим кодом, исполнитель должен показать:

1. Конкретный участок кода.
2. Нарушаемый контракт.
3. Воспроизводимый сценарий.
4. Минимальное исправление.
5. Почему нельзя сохранить согласованную простую архитектуру.

Теоретические failure scenarios без подтверждённого практического значения не являются основанием для расширения архитектуры.

**Конечный результат:** компактная, проверяемая и пригодная для upstream интеграция External ASP при полностью сохранённом штатном TECHO5 Built-in audio path.
