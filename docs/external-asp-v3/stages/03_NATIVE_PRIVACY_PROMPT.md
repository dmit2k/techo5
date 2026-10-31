# Стартовый промпт 03 — native PCM, metadata и SoftwareCut

> **SUPERSEDED for all future execution as of 2026-10-31.** Этот исторический стартовый prompt относится к прежнему raw-metadata контракту и НЕ является разрешением на исполнение. Используй [новый correction prompt](03_PCM_ONLY_CORRECTION_PROMPT.md) и `../PCM_ONLY_ARCHITECTURE_DECISION.md`; для Stage04 дождись двух correction handoffs.


> Текст ниже готов к отправке целиком в новый Work-чат. Он разрешает этап только когда владелец отдельно отправит его как поручение; публикация пакета сейчас не является поручением.

Выполни **только этап 03 TECHO5 / FrankenASP V3 — native PCM, metadata и SoftwareCut**. Используй Work и Desktop Commander на anfisa-trainer.

Сначала полностью прочитай в /home/gptlab/dev/techo5-v3-design-staging/docs/external-asp-v3/:
1. EXECUTION_BRANCH_STRATEGY.md.
2. MASTER_CONTRACT.md.
3. stages/03_NATIVE_PRIVACY_TASK.md.
4. SPEC_V2_APPROVED.md, §§3–4, 5.3–5.3.6, 8–9, 18.1–18.2, 20.4–20.5, 20.10; нужные уточнения/решения перечислены в TASK.
5. HANDOFF_TEMPLATE.md и **принятый stages/02_HANDOFF.md**.

Проверь опубликованный docs HEAD, V2 blob actual SPEC blob from new clean-base docs commit, verified and pinned before Stage, remotes, clean worktrees и applicable AGENTS.md. Входы — exact output integration SHA из handoff 02, только если execution PASS, owner acceptance ACCEPTED и integration COMPLETE. Сверь их с remote TECHO5 experimental/external-asp-v3-dev-main и dmit2k/frankenasp-v3:experimental/v3-dev-main. Отсутствие принятого handoff или несовпадение refs означает STOP; не используй latest/незавершённую stage-ветку.

Работай только в Go stage03/worktree. C зафиксируй unchanged, пустую ветку не создавай. Построй native boundary и единый mute до conversion, сохрани stock; External runtime пока не подключай.

Разрешены изменения только scope TASK, локальные commits и push только собственных stage refs. Не merge/commit/push в integration refs, TECHO5 main или документационную ветку; приёмка и интеграция выполняются отдельно. Не начинай другие этапы.

Выполни **все tests/builds своего TASK** и запиши точные команды, результаты, pair SHA и actual files. Сохрани stock и минимальный контракт без CFG2/TLV, statuses/feedback, ASP Debug, preflight, async queue/retry, platform whitelist, inode cache или compatibility layers. Не переносить автоматически legacy code/installer/OTA/vendor assets.

Echo, аппаратные/акустические тесты и target runtime launch запрещены. При конкретном противоречии остановись с file/SHA/контрактом/repro/minimal fix, не меняя утверждённую архитектуру.

Итог верни по HANDOFF_TEMPLATE, готовый для публикации координатором в stages/03_HANDOFF.md: execution PASS/FAIL/BLOCKED, acceptance PENDING до решения владельца, inputs/tested candidates, output integration status, tests, risks/non-imports, следующий этап. Не выдумывай будущие SHA или выполненную приёмку.
