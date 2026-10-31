# 02 — Новый descriptor v1 и независимые Go/C fixtures

**Цель:** создать и проверить новый wire codec до runtime. Это component/conformance PASS, ещё не paired runtime gate.

## Вход и нормативные требования

Прочитать strategy, MASTER_CONTRACT и принятый `stages/01_HANDOFF.md`. T_in/C_in = его output integration HEAD с acceptance ACCEPTED и integration COMPLETE. Сверить точные remote refs; предыдущий snapshot из плана не заменяет handoff.

V2 §§4.4–4.5, 5, 6.1–6.3, 7.1, 19/R2, 20.1, 21; D04–D07, D28–D33, D55/D57; F1/F2; A1–A5/A8.

Обязательная wire форма, реализуемая **сразу без CFG2/TLV**:

| Header offset | Поле |
|---|---|
| 0 | 4 bytes ASP1 |
| 4 | uint16 LE version=1 |
| 6 / 7 | uint8 buffer count / channel count |
| 8 / 12 | uint32 descriptor size / input block size |
| 16 / 18 / 20 | Length fields for raw identity/hardware/factory metadata sections; final byte encoding and bounds agreed by independent Stage02 Go/C codec |

Header 24 bytes. Buffer record 28: offset/length/rate/frame count/frame stride — uint32 по offsets 0/4/8/12/16; format/valid bits/LSB shift/channels — uint8 по 20/21/22/23; first channel/reserved — uint16 по 24/26. Channel record 8: buffer index uint8, role uint8 (1 MIC, 2 REF), role index uint16, frame byte offset uint32.

Formats: 1 S16LE, 2 packed S24_3LE, 3 S32LE. Signed two's complement; valid bits 2…storage bits, включая знак; shift 0…storage-valid. S32 signed24 low/high: shifts 0/8. Никаких memory-copy Go/C structs.

Offsets buffers последовательны; total input = сумма реальных Length. Проверять overflow и последний sample: `(frames-1)*stride + channelOffset + sampleWidth <= Length`, а также sample внутри stride. Channel ranges непрерывны и покрыты ровно один раз; buffer indices согласованы; role indices непрерывны от 0 независимо для MIC/REF; ≥1 MIC. REF absent допустим.

Buffer count не ограничивать двумя. Выбрать конечный обоснованный предел descriptor с учётом фактического размера raw sources и ширины полей; 8192 из прошлого normalised gain contract не считать доказанно достаточным. Input size независим, весь предел при каждом кадре не выделять.

Metadata are raw source bytes, not Go-generated PlatformInfo / MicFactoryGain values. Передавать original factory IDME/metadata bytes вместе с минимальной однозначной source identity, если она действительно необходима; при отсутствии raw source section отсутствует. Нельзя вызывать layout.Idme()/parseIdme() как raw value, интерпретировать значения, создавать Q30/masks/availability/confidence или сопоставлять coefficients с MIC slots в TECHO5.

Исполнитель Stage02 самостоятельно проектирует **минимальный одинаково реализованный Go/C способ передачи raw values**, их source association, lengths и bounded validation. Не предписывать новую handshake sequence, дополнительную message систему или отдельные абстракции поверх согласованного минимального startup/exchange. Для существующих fixed descriptor fields, зависящих от normalized layout, разрешена минимальная корректировка с независимыми fixtures; зафиксировать wire contract в принятом этапе, не строить framework.

Аппаратный REF — byte-exact source stereo L/R, оба канала и native interleaving/order; никаких host-side scaling, downmix, synthetic REF. Содержимое raw metadata и преобразования vendor DSP — только в C helper.

Смысл startup readiness — однократное подтверждение успешной инициализации перед первым PCM. Конкретный минимальный Go/C формат readiness вместе с raw metadata encoding выбирает Stage02, без избыточного handshake или нового status framework. Далее fixed PCM без framing, 640-byte mono S16LE output. Runtime/timers/процесс здесь не подключаются.

## Scope и размещение

TECHO5: новый компактный codec (допустим новый `echod/internal/asp/`, в upstream такого пакета нет), value types только для необходимых descriptor данных, tests и testdata. C: новый reader/decoder, header и test-only conformance runner. Реальные имена C modules выбрать минимально и записать в handoff.

Reference Go `cfg2_metadata.go` и C `cfg2.c` использовать лишь как исторический пример ошибочного normalised contract, **не** переносить их value semantics, parser/TLV/status/error taxonomy. Для raw sources читать действительный hardware/source backend; interpret only in C. Decode validity отделить от будущего DSP acceptance.

Не менять mic Source, capture, config/HA, secure launch, C DSP/runtime. Не создавать production synthetic adapter, negotiation, registry, compatibility readers, per-frame protocol. Реальные удаления legacy по умолчанию отсутствуют.

## Тесты и приёмка

| ID | Обязательная проверка |
|---|---|
| C01 | Независимые exact-byte fixtures с ручными/spec-derived expected offsets; Go writer и C decoder дают одинаковые значения, expected не вычисляется тем же production encoder |
| C02 | S16/S24_3LE/S32, signed24 low/high, channel order, stride, >2 buffers, interleaved/separate, REF absent |
| C03 | Raw IDME/hardware bytes byte-exact across Go writer / C reader (including embedded zero, high-bit bytes, absent sources and explicit source tags); никаких derived Q30/masks/availability/confidence |
| C04 | Invalid magic/version/size/count/channel coverage/role indices; buffer gaps/overlap, integer overflow, последний sample вне Length; truncated reads; превышение принятого проверенного descriptor bound |
| C05 | Реальная геометрия MIC и stereo REF L/R: два REF channels с независимыми ненулевыми marker samples проходят byte-exact; descriptor bound и размер input обоснованы настоящими raw metadata fixtures, а не старым 186-byte value layout |
| C06 | Native C compiler warnings/errors и ASan/UBSan для decoder; Go codec tests; old CFG2/envelopes не распознаются; unaligned bytes безопасны |
| C07 | Сборка обоих code components из exact commits; fixtures manifest SHA256 одинаковый/проверяемый между репозиториями |

Codec поддерживает **описание**, не DSP работу всех layouts. Нельзя объявить no-REF/S24 DSP support по одному decode test. Граница output 640 фиксируется в constants/fixtures, не реализуется dummy processing.

## Worktrees, commits и handoff

- Go: `/home/gptlab/dev/techo5-v3-stage02`, `experimental/external-asp-v3-dev-stage02` от T_in.
- C: `/home/gptlab/dev/frankenasp-v3-stage02`, `experimental/v3-stage02` от C_in.
- Отдельные локальные commits codec+tests в каждом repo. Stage push только по поручению; integration refs не менять, docs branch не merge.
- Приёмка conformance согласованной парой T_candidate/C_candidate; она не называется runtime gate. После отдельной приёмки/integration записать T02/C02.

Отчёт по HANDOFF_TEMPLATE: оба input/output integration SHA, stage commits, реальные новые files, fixture digest и exact test/build commands, rejection cases, non-import inventory, limits. Docs coordinator сохраняет `stages/02_HANDOFF.md`. Следующий этап 03.

**PASS:** C01–C07 выполнены; чистая согласованная codec pair; runtime явно ещё не подключён. **FAIL:** fixtures проверяют сами себя, нарушена metadata семантика/границы, одна сторона отсутствует, создана old/new compatibility. STOP при конкретном противоречии с file/SHA/scenario и минимальным исправлением, V2 не менять.
