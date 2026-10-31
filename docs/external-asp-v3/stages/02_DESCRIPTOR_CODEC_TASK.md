# 02 — Новый descriptor v1 и независимые Go/C fixtures

**Цель:** создать и проверить PCM-only wire codec до runtime. При correction применять этот TASK вместе с `02_PCM_ONLY_CORRECTION_PROMPT.md`: старый codec с raw metadata признан историческим и должен быть заменён, а не расширен dual reader.

## Вход и нормативные требования

При PCM-only correction прочитать strategy, MASTER_CONTRACT, `PCM_ONLY_ARCHITECTURE_DECISION.md`, исторические `stages/02_HANDOFF.md` и `stages/03_HANDOFF.md` только как evidence. **T_in/C_in = текущая интегрированная пара T03/C02**, проверенная по remote refs, а не старая Stage01 pair. Новый correction Stage02-R handoff создаётся после test/acceptance/integration; не начинать следующий этап до COMPLETE.

V2 §§4.4–4.5, 5, 6.1–6.3, 7.1, 19/R2, 20.1, 21; D04–D07, D28–D33, D55/D57; F1/F2; A1–A5/A8.

Обязательная wire форма, реализуемая **сразу без CFG2/TLV**:

| Header offset | Поле |
|---|---|
| 0 | 4 bytes ASP1 |
| 4 | uint16 LE version=1 |
| 6 / 7 | uint8 buffer count / channel count |
| 8 / 12 | uint32 descriptor size / input block size |
| 16+ | **Никаких полей:** header заканчивается на offset 16, его длина **16 bytes** |

Header **16 bytes**. Buffer record 28: offset/length/rate/frame count/frame stride — uint32 по offsets 0/4/8/12/16; format/valid bits/LSB shift/channels — uint8 по 20/21/22/23; first channel/reserved — uint16 по 24/26. Channel record 8: buffer index uint8, role uint8 (1 MIC, 2 REF), role index uint16, frame byte offset uint32.

Formats: 1 S16LE, 2 packed S24_3LE, 3 S32LE. Signed two's complement; valid bits 2…storage bits, включая знак; shift 0…storage-valid. S32 signed24 low/high: shifts 0/8. Никаких memory-copy Go/C structs.

Offsets buffers последовательны; total input = сумма реальных Length. Проверять overflow и последний sample: `(frames-1)*stride + channelOffset + sampleWidth <= Length`, а также sample внутри stride. Channel ranges непрерывны и покрыты ровно один раз; buffer indices согласованы; role indices непрерывны от 0 независимо для MIC/REF; ≥1 MIC. REF absent допустим.

Buffer count не ограничивать двумя. Точный descriptor size = **16 + 28×buffer count + 8×channel count**, максимум **9196 bytes** при uint8 counts. Input PCM size независим; никакой raw metadata section.

Никаких metadata fields или platform identity в External wire: ни IDME, ни factory gain, ни Q30/masks, ни source names. Существующий stock `layout.Idme()` и другие stock readers не изменять. Вне Stage02 C vendor launcher автономно предоставит DSP необходимые ресурсы.

Исполнитель Stage02 удаляет из Go/C codec **все прежние raw fields/sections и связанную validation**, обновляет independent byte fixtures; сохраняет existing one-byte readiness `0xA5`. Не добавлять startup handshake, дополнительные message types или dual-format compatibility. Для существующих fixed descriptor fields, зависящих от normalized layout, разрешена минимальная корректировка с независимыми fixtures; зафиксировать wire contract в принятом этапе, не строить framework.

Аппаратный REF — byte-exact source stereo L/R, оба канала и native interleaving/order; никаких host-side scaling, downmix, synthetic REF. Vendor resources и calibration находятся **вне descriptor**, во владении FrankenASP launcher/child.

Startup readiness уже frozen как **один byte `0xA5`**; correction его не меняет. Далее fixed PCM без framing, 640-byte mono S16LE output. Runtime/timers/процесс здесь не подключаются.

## Scope и размещение

TECHO5: изменить имеющийся Stage02 codec в `echod/internal/asp/` строго до PCM-only; value types только для PCM geometry, обновить tests/testdata. C: новый reader/decoder, header и test-only conformance runner. Реальные имена C modules выбрать минимально и записать в handoff.

Reference Go `cfg2_metadata.go` и C `cfg2.c` использовать лишь как исторический пример ошибочного normalised contract, **не** переносить их value semantics, parser/TLV/status/error taxonomy. Не читать hardware или IDME; fixture input — только синтетическая PCM geometry. Это codec-only stage. Decode validity отделить от будущего DSP acceptance.

Не менять mic Source, capture, config/HA, secure launch, C DSP/runtime. Не создавать production synthetic adapter, negotiation, registry, compatibility readers, per-frame protocol. Реальные удаления legacy по умолчанию отсутствуют.

## Тесты и приёмка

| ID | Обязательная проверка |
|---|---|
| C01 | Независимые exact-byte fixtures с ручными/spec-derived expected offsets; Go writer и C decoder дают одинаковые значения, expected не вычисляется тем же production encoder; отдельно закрепить и проверить точные bytes/длину минимального readiness contract, принятого совместно Go/C |
| C02 | S16/S24_3LE/S32, signed24 low/high, channel order, stride, >2 buffers, interleaved/separate, REF absent |
| C03 | Отсутствие любых metadata fields/sections: exact 16-byte header, strict total length, reject старые 24-byte descriptors и trailing bytes; PCM кодируется независимо от существования IDME |
| C04 | Invalid magic/version/size/count/channel coverage/role indices; buffer gaps/overlap, integer overflow, последний sample вне Length; truncated reads; превышение принятого проверенного descriptor bound |
| C05 | Native MIC/stereo REF L/R: два REF channels с независимыми marker samples byte-exact; descriptor bounded только количеством buffer/channel records, PCM input независим |
| C06 | Native C compiler warnings/errors и ASan/UBSan для decoder; Go codec tests; old CFG2/envelopes не распознаются; unaligned bytes безопасны |
| C07 | Сборка обоих code components из exact commits; fixtures manifest SHA256 одинаковый/проверяемый между репозиториями |

Codec поддерживает **описание**, не DSP работу всех layouts. Нельзя объявить no-REF/S24 DSP support по одному decode test. Граница output 640 фиксируется в constants/fixtures, не реализуется dummy processing.

## Worktrees, commits и handoff

- Go: `${WORKSPACE}/techo5-v3-stage02-pcmonly`, `experimental/external-asp-v3-dev-stage02-pcmonly` от T03.
- C: `${WORKSPACE}/frankenasp-v3-stage02-pcmonly`, `experimental/v3-stage02-pcmonly` от C02.
- Отдельные локальные commits codec+tests в каждом repo. Stage push только по поручению; integration refs не менять, docs branch не merge.
- Приёмка conformance согласованной парой T_candidate/C_candidate; она не называется runtime gate. После отдельной приёмки/integration записать T02/C02.

Отчёт по HANDOFF_TEMPLATE: оба input/output integration SHA, stage commits, реальные новые files, fixture digest и exact test/build commands, rejection cases, non-import inventory, limits. Docs coordinator сохраняет `stages/02_PCM_ONLY_CORRECTION_HANDOFF.md`. Следующий этап 03-R только после ACCEPTED + COMPLETE.

**PASS:** C01–C07 выполнены; чистая согласованная codec pair; runtime явно ещё не подключён. **FAIL:** fixtures проверяют сами себя, в descriptor остались metadata/IDME/platform поля или old/new compatibility, одна сторона codec отсутствует. STOP при конкретном противоречии с file/SHA/scenario и минимальным исправлением, V2 не менять.
