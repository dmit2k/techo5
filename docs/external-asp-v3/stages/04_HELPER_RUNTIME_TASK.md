# 04 — Минимальный C helper, DSP bridge и build

**Внимание:** использовать входные SHA **только из принятого `03_PCM_ONLY_CORRECTION_HANDOFF.md`**, не из исторического `03_HANDOFF.md`. Не запускать Stage04 до обоих correction gates.

**Цель:** создать новую C runtime-сторону в frankenasp-v3, готовую к paired wire gate 05. Это C component PASS, не совместная интеграция и не vendor hardware qualification.

## Вход и нормативный контекст

Принятый **`stages/03_PCM_ONLY_CORRECTION_HANDOFF.md`**: T_in/C_in = output integration HEAD, acceptance ACCEPTED, integration COMPLETE. До этого Stage04 — STOP. C_in содержит принятый PCM-only correction codec; T_in содержит исправленный native/privacy Stage03 без External metadata readers. Проверить remote refs, independent C origin и clean worktree.

Читать V2 §§4–7, 14, 17.4, 18.6, 20.1–20.3, 21–24; D05–D07, D25–D34, D50/D52–D55/D57–D58; F1/F2. §7 требует сохранения **подтверждённой** DSP geometry, а не реализации всех представимых formats.

## Scope нового repository

Новый C repository имеет собственную минимальную структуру: codec 02, native input reader, fixed output writer, DSP compatibility/adaptation, bridge/FIFO, vendor ABI binding, необходимые launch/build components и tests. Реальные filenames определить здесь и записать в handoff; старые `tools/external-asp/` paths являются только reference.

Читать exact C reference `52d51de33e3105501b298c1c214db55f05a2eb82`:
- core.c/.h: captured-slot mapping, low/high signed24 MIC, REF promotion, 768-input-frame/128-output quantum;
- factory_adapter.c/.h, factory.c: existing ABI/lifecycle/error handling;
- cfg2_runtime.c: лишь исторический пример недопустимого admission по host-derived metadata, **не parser**;
- factory_child.c: wire FD / stdout→stderr;
- launcher/launcher.c и build_bionic.sh/BIONIC_FACTORY.md: необходимые FD-pinned launch и ARM32/Bionic requirements.

Минимально переиспользованные фрагменты получают origin SHA/path и applicable MIT notices. Не делать copy каталога или cherry-pick старой истории.

## Реализовать

1. Прочитать согласованный **PCM-only descriptor** без IDME/hardware sections. Подтверждённая vendor factory geometry: MIC4 S32 signed24 low/high, REF2 stereo L/R S16, 96 kHz, 1920 frames/20 ms. Квалифицировать PCM geometry и реальные vendor DSP prerequisites, а не отсутствующие metadata в wire. Ни host admission, ни derived Go gain.
2. Валидный wire descriptor без REF, с unsupported rate/layout/format/20-ms temporal geometry отклонять до READY: stderr и exit/EOF. Новые resamplers/универсальный DSP не нужны.
3. Numeric adaptation — исключительно C: MIC signed24 alignment, независимые REF L/R, S16 promotion, DSP-specific mapping/conversions. **Factory calibration не передаётся и не парсится из descriptor**: vendor Calibrator сам использует factory resources, подготовленные launcher. Не добавлять внешнего Q30/gain multiplier, host-derived masks/slot association или имитации vendor calibration.
4. Существующий DSP input quantum 768 frames при 96 kHz → 128 mono samples при 16 kHz. Инициализировать output FIFO 64 нулями **один раз**. Первый exchange: 2 quantum дают 256 samples +64 padding =320; следующий получает достаточно output благодаря сохранённому input remainder. Каждый successful exchange выдаёт 320 samples. Тестировать длительную последовательность, не только первый блок.
5. Ни NotReady, ни sequence/origin/status не передаются. Внутренние необходимые FIFO occupancy/quantum bookkeeping допустимы; не переносить старые wire timeline/diagnostic counters.
6. После успешных descriptor/qualification/vendor init выдать только согласованный Stage02 readiness marker, затем читать полные buffers и писать ровно 640 bytes little-endian. Partial I/O/EINTR обрабатывать до completion; реальные errors завершают session, без retry протокола.
7. До dlopen/vendor init сохранить отдельный wire FD, обычный stdout направить в stderr. Vendor chatter не попадает в PCM.
8. **Обязательно восстановить назначение проверенного FrankenASP launcher**: до запуска C child подготовить его разрешённые factory resources, read-only vendor assets и, где они действительно требуются конкретному backend, `/proc/idme/miccal.N` aliases на подлинные Device Tree/sysfs значения. Проверить provenance, trusted paths, owner/mode, FD-pinning, mount/exec isolation и фиксированное environment. Launcher не читает calibration bytes, Calibrator остаётся единственным authority. Это частная platform-specific функция FrankenASP, не generic host. Если требуется native launcher + ARM32/Bionic child, перенести только необходимые механизмы без installer/OTA/supervisor. Не импортировать старый installer/OTA acquisition/source discovery. Запуск должен потреблять уже подготовленный owner-provided trusted runtime; documented prerequisites не означают, что он установлен.
9. Сохранить реальные launch trust boundaries и vendor lifecycle cleanup, без runtime-wide hash/manifest framework. Закрепить конкретный минимальный launch/build contract для Go 05: expected executable/assets paths, inherited FDs, отсутствие CLI bypass, child replacement/process ownership.
10. Production build исключает test DSP/mock modes. Native test build связывает **тот же codec/wire/bridge** с test-only independent ABI mock. Никаких synthetic production fallback и fake READY, если настоящий backend init не успешен.

## Проверки этапа

| ID | Обязательные tests / artifacts |
|---|---|
| H01 | C decoder + supported configuration positive path; valid unsupported/no-REF/rate/layout/format/temporal cases fail до READY, malformed bytes fail безопасно |
| H02 | REF L/R markers byte-exact; C descriptor без factory metadata. Проверить launcher resource views/IDME aliases для child, отсутствие передачи Q30/calibration через wire, no-duplicate-gain, корректный vendor-init отказ при отсутствии необходимого ресурса (только helper), MIC alignment/permutation в C |
| H03 | FIFO 64 zeros только output, первый и каждый последующий 320-sample frame, 8-ms quantum/remainder, длинный nonzero stream без потерь/повторов, repeated mute input не дополняет padding |
| H04 | Process wire тест compiled C runtime: descriptor→READY→fixed PCM; partial I/O/EINTR/EOF, invalid input truncation, чистый PCM при runtime stdout chatter |
| H05 | Vendor ABI mock success/failure: init/create/ref/process counts, buffer bounds/canaries, terminal errors, exactly-once destroy/deinit; mock не доказывает hardware behavior |
| H06 | Native C tests с warnings-as-errors и ASan/UBSan; production/test source/link inventory, никаких mock paths в production binary |
| H07 | ARM32 production helper и необходимый launcher crossbuild, readelf ABI/interpreter/dependencies и SHA256; toolchain/private runtime provenance, vendor bytes не в Git |
| H08 | File/provenance/security review: minimal preserved factory launcher resource preparation/aliases, FD/mount/exec, no CFG2/TLV/old framing/status/NotReady/installer/OTA; vendor bytes отсутствуют в Git |

Не запускать privileged mounts/chroot, target binaries или Echo. C launch substrate здесь принимается по source/security review и crossbuild; настоящее vendor runtime/device launch отдельно. Отсутствие нужного toolchain/owner-provided link dependencies — BLOCKED, не заменять H07 native mock build.

## Git и передача

C worktree `/home/gptlab/dev/frankenasp-v3-stage04`, branch `experimental/v3-stage04` от C_in. TECHO5 не меняется: T_out=T_in, пустой Go stage04 не создавать. Локальные C commits для runtime/bridge/build/tests; stage push только по поручению; C experimental/v3-dev-main не обновлять автоматически.

Handoff по HANDOFF_TEMPLATE в `stages/04_HANDOFF.md` через docs coordinator: T04/C04, реальные modules/entrypoints, reproducible native и ARM32 команды, launch prerequisites, production vs mock binary identities, fixture digest, H01–H08, exact reuse/license map, неперенесённые legacy, ограничение «ещё не Go/C integration». Hardware НЕ ВЫПОЛНЕНЫ. Следующий — 05.

**PASS:** H01–H08 выполнены и C component совместим с новым codec. **FAIL:** только parser/mock без production build, первый NotReady, input padding, ложная поддержка форматов, импорт старого проекта целиком. При code/contract conflict STOP с exact file/SHA/scenario/minimal fix, без изменения V2.
