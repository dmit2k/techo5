# PCM-only boundary — архитектурная коррекция V3, 2026-10-31

**Решение владельца:** восстановить проверенную в предыдущем FrankenASP границу ответственности: **универсальный TECHO5 передаёт только native MIC/REF PCM и его геометрию; FrankenASP launcher самостоятельно подготавливает factory ASP resources, в том числе нужные vendor backend IDME aliases, до запуска C child, а vendor Calibrator применяет calibration**. Эта поправка является нормативной для дальнейших Stage02–07 и **заменяет** прежнее требование передавать сырые IDME/hardware metadata через External descriptor. Не добавлять два режима протокола.

**Статус исполнения:** архитектурное решение принято; **новая Go/C implementation ещё не принята**. Предыдущие Stage01/02/03 отчёты, commits и acceptance receipts исторически действительны только для предыдущего wire contract. Запрещается считать T03/C02 соответствующими PCM-only до корректирующих code gates.

## 1. Причина и подтверждённый reference

В предыдущем FrankenASP launcher был создан именно для безопасной подготовки vendor resources. Подготовка aliases `/proc/idme/miccal.0..3` на подлинные Device Tree values, isolated FS/mount/FD/exec выполнялась на стороне FrankenASP **без чтения самих calibration bytes**. C vendor ABI adapter вызывал `asp_init`, `asp_create_pipeline`, `asp_process` и не получал Q30 через generic PCM wire; vendor Calibrator был единственным authority для gains. Проверенный reference: `dmit2k/techo5` historical `tools/external-asp/launcher/launcher.c`, `factory_adapter.c` и security review/README. **Перенести ответственность, а не старый CFG2/TLV/installer/OTA/status framework или весь launcher целиком.** Из прошлого переносить только действительно нужные, отдельно проверенные trusted path, verified aliases, FD/mount/exec и Bionic substrate.

Чистый stock TECHO5 `layout.Idme()` и Athena eMMC fallback обслуживают существующие stock identity paths; stock `subband.Load()` уже самостоятельно выбирает vendor coefficients для собственного beamformer. Нет основания делать External generic host ответственным за чтение `miccal.*`. Наличие IDME/калибровки зависит от платформы и не должно быть prerequisite для native PCM.

## 2. Нормативная граница

**TECHO5:** штатная настройка capture и platform-specific drivers, фактические MIC/REF PCM bytes, buffer/channel geometry (rate, format, valid bits/shift, frames, stride, offsets, hardware roles), общий SoftwareCut до штатной conversion, stock Built-in DSP/sensors, secure host launch, PCM IPC, Selected/Active, fallback, mono publication/lifetime. **Никакого чтения IDME/Device Tree/eMMC для External и никакого IDME/platform/calibration в descriptor.** Прежние stock-only вызовы могут продолжать читать свои данные независимо от External.

**External descriptor:** только PCM geometry, без платформенных полей и произвольных секций. Fixed 16-byte header `ASP1`, v1, two uint8 counts, uint32 descriptor/input lengths; buffer records 28 bytes; channel records 8 bytes. Descriptor length = `16+28*buffers+8*channels` (предел 9196 bytes). Существующий one-byte readiness `0xA5`, fixed input exchange и 640-byte mono output сохраняются. Старый metadata-bearing 24-byte header **несовместим и удаляется**; никакой legacy decoder или migration negotiation. Это pre-deployment replacement; новый код принимается только согласованной парой Go/C.

**FrankenASP launcher:** в собственном, проверенном runtime namespace до запуска child безопасно открывает/mounts разрешённые vendor assets и read-only factory path aliases, которые нужны *этому* backend. Не читает/не парсит calibration, не умножает PCM на gains. Точные `/proc/idme/miccal.N` aliases — частная потребность factory backend, а не требование к другим helpers. Не требовать mount всех этих путей там, где их нет; отсутствие необходимого ресурса — ошибка запуска **конкретного helper**, а не универсального host/descriptor.

**Vendor DSP / C child:** проверяет PCM geometry и фактические доступные ему factory resources; сам vendor Calibrator читает и применяет coefficients. Никаких Q30/masks/confidence/slot-gain fields в IPC. При невозможности init нет READY; процесс завершается с понятным stderr, TECHO5 сохраняет Built-in. Не изображать успех mock как фактический запуск vendor library.

**Инварианты:** REF L/R byte-exact, отсутствие REF допустимо в generic codec, native before conversion, один SoftwareCut boundary, без копий input ради DSP, no fake hardware, без per-frame framing/status, сохранение secure launch и lifecycle. Реальная поддержка конфигураций остаётся вопросом конкретного DSP.

## 3. Исправление по этапам и acceptance gates

1. **Stage02-R — codec correction (Go+C)**: вход — текущая **интегрированная** пара Go T03 `4fb329cf789b156bbeb8d13da20b8c4bda70f108`, C C02 `0f39216a5240306344af7c2457716f4eaa466f40`. В отдельных *новых* correction branches заменить Go/C codec, independent fixtures, тесты. Удалить минимально зависящий от RawField bridge Stage03, если нужен для компиляции; полноценное удаление raw readers — Stage03-R. Без изменения MIC capture/SoftwareCut, без main/old stage branches. Кодовый исполнитель публикует свои branches; независимый coordinator проверяет C01–C07, принимает точную пару, FF обе dev-main, публикует `stages/02_PCM_ONLY_CORRECTION_HANDOFF.md`. **STOP: Stage03-R без обоих ACCEPTED+COMPLETE.**

2. **Stage03-R — native/privacy correction (Go only)**: вход — новая интегрированная пара Stage02-R. Удалить весь V3-only raw-IDME/identity reader/bridge/test code, не трогать stock `layout.Idme/readIDME/idmeFallback`. Проверить отсутствие изобретённого PCM format и REF role; при необходимости минимально исправить capture geometry, preserving existing stock. Перезапустить N01–N08, race, build tags и no-IDME invariants. Новая Go branch; coordinator review, FF только Go dev-main, публикация `stages/03_PCM_ONLY_CORRECTION_HANDOFF.md`. **STOP: Stage04 без ACCEPTED+COMPLETE.**

3. **Stage04 — C helper / launcher**: только от Stage03-R принятой пары; сохранять audited, minimal resource preparation из прежнего factory launcher, без переноса CFG2/status/installer. H01–H08, production ARM32/Bionic build, source/security proof. При необходимости Toolchain/H07 BLOCKED честно. Не запускать Echo. После независимой приёмки/integration — Stage05–07 по скорректированному контракту.

4. **Stage05–07**: no raw/IDME in IPC; 05 paired compiled Go/C wire проверяет только PCM descriptor, standalone factory resource preparation/sandbox/permissions — H02/H08/W08, 06 source/control/lifecycle, 07 финальная pair verification. Прежние historical handoffs 01–03 **не редактировать**; ссылаться на correction receipts.

## 4. Git и разрешения

Документы: `dmit2k/techo5:experimental/v3-design-staging`. Стартовые обе dev-main: T03/C02 выше; protected main — Go `62603a7854283a3a07473a17c831c6ef0fd6f8d4`, C `e986ae18643e08e95f29a38aae1dd74762f6362e`. Перед каждым этапом повторно проверять refs/clean worktrees. Никаких reset/force/cherry-pick/squash, automatic merges, изменения main или Echo. В каждом correction отдельно фиксировать tested SHA, accepted pair и remote integration HEAD. Прямая авторизованная публикация stage refs — с `validation host`.

**Задания для отдельных рабочих чатов**: `stages/02_PCM_ONLY_CORRECTION_PROMPT.md`, `stages/03_PCM_ONLY_CORRECTION_PROMPT.md`, `stages/04_PCM_ONLY_RESTART_PROMPT.md`. Исполнять **строго последовательно**, и каждый следующий prompt только после записи предыдущего accepted/integrated handoff. Публикация этих prompts не запускает worker сама по себе.
