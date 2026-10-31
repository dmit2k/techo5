# Исполнительский prompt — Stage02-R: PCM-only descriptor correction (Go + C)

**ВЫПОЛНИ код, tests, commits и публикацию двух новых stage branches. Это коррекция уже интегрированного Stage02 после PCM-only архитектурного решения, а не повтор исходного Stage02.**

На `validation host` с authorized terminal tools и GitHub. Сначала прочитай полностью `PCM_ONLY_ARCHITECTURE_DECISION.md`, `EXECUTION_BRANCH_STRATEGY.md`, `MASTER_CONTRACT.md`, `stages/02_DESCRIPTOR_CODEC_TASK.md`, `SPEC_V2_APPROVED.md` §§4–7, 20.1, `stages/03_HANDOFF.md` и `HANDOFF_TEMPLATE.md`.

**Проверенный исходный pair (после исходного Stage03):**
- TECHO5 `experimental/external-asp-v3-dev-main`: `4fb329cf789b156bbeb8d13da20b8c4bda70f108` (T03).
- FrankenASP V3 `experimental/v3-dev-main`: `0f39216a5240306344af7c2457716f4eaa466f40` (C02).
- Docs: `experimental/v3-design-staging` — прочитай **текущий remote HEAD и фактический SPEC blob**; никаких старых handoff SHA вместо этого контракта.
- Go/C `main` должны оставаться `62603a7854283a3a07473a17c831c6ef0fd6f8d4` / `e986ae18643e08e95f29a38aae1dd74762f6362e`.

Новые branches (если заняты — STOP без force/reset):
- Go `experimental/external-asp-v3-dev-stage02-pcmonly`, worktree `${WORKSPACE}/techo5-v3-stage02-pcmonly`, base T03.
- C `experimental/v3-stage02-pcmonly`, worktree `${WORKSPACE}/frankenasp-v3-stage02-pcmonly`, base C02.

**Обязательная коррекция:** заменить действующий metadata-bearing descriptor 24/28/8 на **PCM-only 16/28/8**, endian LE; header ASP1/version1/counts/descriptor size/PCM bytes; strict length = `16+28*buffers+8*channels`, максимум 9196 bytes при uint8 counts. **Удалить** Platform/Hardware/Factory RawField, source name/value encoding, raw metadata length fields, старые metadata fixtures и пределы 64KiB/64 fields/32KiB. Сохранить действующие PCM physical buffers, MIC/REF roles/offsets/valid bits, signed24 low/high, stereo REF L/R и absent REF, readiness **0xA5**, output 640 bytes. Не создавать dual decoder, wire v1/v2 registry, compatibility or fake calibration. Старые 24-byte descriptors обязательно отвергать.

С учётом входа T03 минимально удалить или изменить `NativeBlock.DescriptorWithRaw` (он зависит от удаляемых fields), чтобы Go code компилировался. **Не удалять** пока `layout.RawFactorySources` или не рефакторить capture и SoftwareCut: полное удаление External raw readers делает Stage03-R. Не вводить новых обращений к hardware.

**Тесты:** повтор C01–C07 на новой точной pair, независимые exact-byte fixtures Go/C, malformed/truncated/trailing старый формат, >2 buffers, S16/S24/S32, valid bits/stride/overflow, REF L/R distinct sample markers, no REF, byte-exact input; Go race/vet и compiled C sanitizer. Проверить отсутствие metadata sections в реальных *encoded bytes*, а не только grep. Stock regression не ухудшить.

Commit/push только двух новых correction branches напрямую с validation host, перепроверь remote SHA и оба main/dev-main (dev-main **не продвигать**). Подготовь `02_PCM_ONLY_CORRECTION_HANDOFF.md` в отдельном evidence каталоге для координатора: inputs, output candidates, diff, new wire bytes, fixtures SHA256, C01–C07, build/tests, risks, owner acceptance PENDING, integration PENDING. Не писать в docs branch. На выявленном противоречии остановиться с точными кодом/SHA/воспроизведением; не менять норматив самостоятельно. Stage03-R и Stage04 не запускать.
