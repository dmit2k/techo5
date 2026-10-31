# Исполнительский prompt — Stage04: C helper with restored factory launcher resource preparation

**ЗАПРЕЩЕНО запускать до ACCEPTED + INTEGRATION COMPLETE по обоим PCM-only correction handoffs Stage02-R и Stage03-R.**

Используй Desktop Commander на `anfisa-trainer` и GitHub. Прочитай `PCM_ONLY_ARCHITECTURE_DECISION.md`, `MASTER_CONTRACT.md`, `EXECUTION_BRANCH_STRATEGY.md`, `SPEC_V2_APPROVED.md` §§4–7,14,20.2–20.3, `stages/04_HELPER_RUNTIME_TASK.md`, `stages/02_PCM_ONLY_CORRECTION_HANDOFF.md`, `stages/03_PCM_ONLY_CORRECTION_HANDOFF.md`, `HANDOFF_TEMPLATE.md`. Проверить **фактические** remote Go/C dev-main SHA из последнего correction handoff; оба main frozen и worktrees clean. Не использовать старые T03/C02, если correction завершён.

Создать `experimental/v3-stage04` от нового C integration HEAD в `/home/gptlab/dev/frankenasp-v3-stage04` (STOP если branch/worktree уже существуют/заняты; не force). TECHO5 Go не менять.

**Восстановить сохранённый рабочий factory-launcher contract:** FrankenASP launcher до child подготавливает verified vendor assets и разрешённые read-only IDME aliases `/proc/idme/miccal.N` на подлинные источники в контролируемом namespace; сохраняет trusted FD3 executable/FD4 assets, verified paths/ownership, mount isolation, Bionic child/runtime environment и безопасный exec. Применение/чтение calibration находится **только у vendor Calibrator**; launcher и TECHO5 не сериализуют factory values, не рассчитывают Q30/gain masks, не масштабируют PCM. Это частная реализация FrankenASP: другие helpers и host не обязаны иметь IDME. Отсутствие необходимых vendor ресурсов → *только helper* stderr/exit до READY, сохранение host Built-in.

**Reference only**: исторические `tools/external-asp/launcher/launcher.c`, `factory_adapter.c`, `factory.c`, `factory_child.c` из предыдущего подтверждённого FrankenASP, с exact provenance. Перенести только доказанные resource mapping/FD/mount/exec/ABI fragments и licenses; **никаких CFG2/TLV, statuses, general installers/OTA, platform whitelist в TECHO5, old bootstrap/control**. Сохранить PCM-only 16/28/8 codec, readiness `0xA5`, DSP 768→128 FIFO +64 initial once, exact 640-byte output, REF stereo L/R, compiled wire/stdout isolation.

Проведи H01–H08, в особенности H02/H08: tests of launcher aliases/resources and trust boundary, failure when necessary vendor file missing, mock success without host metadata, no Q30 in wire, H07 actual ARM32/Bionic production crossbuild (no mock substituted), independent C sanitizers. Не запускать Echo/mount/chroot на устройстве. Документируй vendor-runtime limitations отдельно от codec/pass. Если H07 невозможно — BLOCKED с точной причиной, не называй PASS.

Commit и direct SSH push только C Stage04 branch, не изменять main/dev-main/docs. Verify SHA, clean. Подготовить `04_HANDOFF.md` для coordinator по template, H01–H08 и exact launch provenance; acceptance/integration PENDING. Не запускать Stage05.
