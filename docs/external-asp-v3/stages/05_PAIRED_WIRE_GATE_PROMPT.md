# Стартовый промпт 05 — secure session и первый Go/C wire gate

> Текст ниже готов к отправке целиком в новый Work-чат. Он разрешает этап только когда владелец отдельно отправит его как поручение; публикация пакета сейчас не является поручением.

Выполни **только этап 05 TECHO5 / FrankenASP V3 — secure session и первый Go/C wire gate**. Используй Work и authorized terminal tools на validation host.

Сначала полностью прочитай в ${WORKSPACE}/techo5-v3-design-staging/docs/external-asp-v3/:
1. EXECUTION_BRANCH_STRATEGY.md.
2. MASTER_CONTRACT.md и PCM_ONLY_ARCHITECTURE_DECISION.md.
3. stages/05_PAIRED_WIRE_GATE_TASK.md.
4. SPEC_V2_APPROVED.md, §§5–7, 11.5–11.6, 12–14, 16–17, 20.1–20.3, 20.6–20.9, 21–22; нужные уточнения/решения перечислены в TASK.
5. HANDOFF_TEMPLATE.md и **принятый stages/04_HANDOFF.md**.

Проверь опубликованный docs HEAD, V2 blob actual SPEC blob from new clean-base docs commit, verified and pinned before Stage, remotes, clean worktrees и applicable AGENTS.md. Входы — exact output integration SHA из принятого handoff 04: execution PASS, **independent delegated technical acceptance ACCEPTED** и integration COMPLETE; это координаторская техническая приёмка, а не вымышленная дословная декларация владельца. Актуальный принятый pair: Go `b0c0d5fa96576eff8f7497d48e39fa30a5348e76`, C `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`; всё равно независимо проверить фактические GitHub refs до работы. Сверь их с remote TECHO5 experimental/external-asp-v3-dev-main и dmit2k/frankenasp-v3:experimental/v3-dev-main. Отсутствие принятого handoff или несовпадение refs означает STOP; не используй latest/незавершённую stage-ветку.

Создай Go stage05/worktree; C проверь на accepted SHA в isolated worktree, stage05 только при реальном scoped C change. Выполни compiled Go/C integration и обе production builds. Фиксируй tested pair T_SHA+C_SHA; production Source activation остаётся этапу06.

Разрешены изменения только scope TASK, локальные commits и push только собственных stage refs. Не merge/commit/push в integration refs, TECHO5 main или документационную ветку; приёмка и интеграция выполняются отдельно. Не начинай другие этапы.

Выполни **все tests/builds своего TASK** и запиши точные команды, результаты, pair SHA и actual files. Сохрани stock и минимальный контракт без CFG2/TLV, statuses/feedback, ASP Debug, preflight, async queue/retry, platform whitelist, inode cache или compatibility layers. Не переносить автоматически legacy code/installer/OTA/vendor assets.

Echo, аппаратные/акустические тесты и target runtime launch запрещены. При конкретном противоречии остановись с file/SHA/контрактом/repro/minimal fix, не меняя утверждённую архитектуру.

Итог верни по HANDOFF_TEMPLATE, готовый для публикации координатором в stages/05_HANDOFF.md: execution PASS/FAIL/BLOCKED, acceptance PENDING до решения владельца, inputs/tested candidates, output integration status, tests, risks/non-imports, следующий этап. Не выдумывай будущие SHA или выполненную приёмку.
