# 03 — Native capture, metadata и единый SoftwareCut

**Цель:** подготовить фактический input для helper и закрыть privacy boundary до первого включения External. Runtime ещё не запускается; production output остаётся Built-in.

## Вход и контекст

T_in/C_in — output integration HEAD принятого `stages/02_HANDOFF.md`; acceptance ACCEPTED, integration COMPLETE. Проверить оба remote integration refs и docs/V2 blob по MASTER_CONTRACT.

Читать V2 §§3–4, 5.3–5.3.6, 8–9, 18.1–18.2, 20.4–20.5, 20.10; D01–D24, D55–D56; F1/F2/F4/F5. 02 codec уже принят, новый descriptor не проектировать.

## Реальные точки изменения

- `echod/internal/hardware/mic/mic.go`: capture interface, Source.Run/broadcast, SoftwareCut, stock owned mono frame и publication.
- `echod/internal/hardware/mic/athena.go`: pairedCapture.Read, micBuf/refBuf, convert, softFill, meters.
- `capture_alsa.go`/captureDevice: plain interleaved bytes без перепаковки.
- `layout/layout.go`: Idme; `layout/device_cronos.go`: имеющиеся identity sources.
- `capture_panel.go` и device build tags — сохранение stock compatibility, без специальной новой External platform subsystem.

Проверить эти symbols на T_in. Старые retainNative/aspConfig/nativeFrame — **reference**, в upstream их нет. Не воспроизводить их Athena-only admission.

## Требуемая реализация

1. Добавить минимальный optional native access перед paired convert, сохранив обычный Read. Native MIC/REF остаются в исходных buffers/packing/order/rates; generic host не делает DSP mapping/resampling/FIFO. Для plain Read использовать имеющиеся данные напрямую.
2. Получать фактическую геометрию из capture settings, а не подгонять к MIC4/REF2/96 kHz. S32 signed24 low/high описать valid bits=24, shift=0/8. Last sample и bytes соответствуют codec 02.
3. При отсутствии hardware REF descriptor не обещает его наличие. Stock software reference, если он существовал, остаётся в Built-in. Не преобразовывать его специально в новый native hardware REF и не расширять DSP support. Panel сохраняет свой stock путь.
4. На block прочитать SoftwareCut **один раз** на общей границе до native conversion. По тому же значению обнулить native MIC+REF до helper/convert и сохранить stock clear(raw). Для packed/interleaved учитывать channel offsets/stride/sample width; in-place где безопасно.
5. Stock Read/convert/DSP/sensors продолжает работать. Ранние нули изменяют resampler state/внутренние meters и переходные samples — это разрешено; не компенсировать resets/filters.
6. Сформировать descriptor из фактического capture/native geometry и доступных неизменённых raw source bytes. Все hardware REF L/R должны сохраняться побайтово, включая независимые samples обоих каналов, порядок, частоту, формат и interleaving. Ни одного DSP-specific host преобразования.
7. Если factory IDME/metadata доступны в виде оригинальных bytes, предоставить их напрямую без layout.Idme()/parseIdme() normalised output. Никаких Go Q30, gain masks, availability/confidence, factory completeness, MIC-slot coefficient association или platform admission. Отсутствие raw source не заполнять derived values. Если stock API выдаёт только разобранные значения, локально добавить минимальное raw read без изменения existing Built-in behavior.

8. Native input borrowed до следующего read; будущий sync exchange 05/06 обязан завершить использование до его повторения. Не вводить input copies/pools/queues.
9. Сохранить stock mono ownership и raw copy. Проверить удержание stock frames сейчас; **External retention и окончательный выбор output реализуются в 06**, когда есть runtime. Не добавлять временный fake External selector в production.

## Собственные tests

| ID | Проверка / PASS |
|---|---|
| N01 | Paired native callback получает исходные MIC/REF bytes до convert; старый Read сохраняет stock geometry/output вне mute |
| N02 | Plain interleaved input и separate/no-REF без копии/перепаковки; аппаратный stereo REF L/R проверен независимыми ненулевыми markers, оба канала без host scaling/downmix/resampling/reorder |
| N03 | Одно принятое mute value используется native+stock: native MIC/REF и stock raw заглушены, в том числе при stock software reference; проверка packing/stride/canary bounds |
| N04 | Borrowed lifetime до следующего capture read; никаких retained input после read; instrumentation только test code |
| N05 | Stock frame retention через несколько reads, history/subscriptions/raw copy; stock sensors/DSP/cancel path не перенесены на External |
| N06 | Hardware mute REF semantics сохранены; mute/unmute не сбрасывает DSP/FIR/history, не содержит второго output check; transition samples не сравнивать побитово со старым mute |
| N07 | Byte-exact raw IDME source value fixtures, включая произвольные bytes/нули, неизвестные ключи и отсутствие значений; Go не вызывает normalised IDME, не вычисляет Q30/masks/confidence/availability/MIC slot mapping, helper interpretation остаётся вне Go |
| N08 | Relevant mic/privacy/layout/codec tests, `go test -race` для затронутых concurrent paths; tags default/dot/dot,donut/spot/panel compile и stock tests |

Для SoftwareCut tests наблюдать данные на boundary, не считать механический grep единственным доказательством single policy. Не вызывать real hardware для fixtures. Существующий XRUN reset **stock** decimator в pairedCapture.restart не удалять: запрет относится к новым DSP resets при mute и к helper lifecycle.

## Git / результат

Go worktree `/home/gptlab/dev/techo5-v3-stage03`, branch `experimental/external-asp-v3-dev-stage03` от T_in. C не изменяется, пустую stage03 не создавать, C_out=C_in. Коммиты только native/privacy/metadata/tests; no integration push/merge. Stage push — в пределах поручения.

Отчёт HANDOFF_TEMPLATE → docs coordinator → `stages/03_HANDOFF.md`: T03/C03, точные geometry/source/API contracts для 04/05/06, files, commands/results, input/output ownership доказательства, неперенесённые whitelist/CFG2/synthetic, upstream stock exceptions. Hardware НЕ ВЫПОЛНЕНЫ. Следующий этап 04.

**PASS:** N01–N08 успешны, stock сохранён, native privacy готова без runtime. **FAIL:** host DSP adaptation/admission, ложный REF/metadata, второй mute check, публикация borrowed output, нарушение Panel/stock. STOP по MASTER_CONTRACT при противоречии с конкретным code/SHA/scenario; не менять ТЗ и не изобретать adapter hierarchy.
