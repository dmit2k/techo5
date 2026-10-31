# Стартовый промпт 06 — production Source, lifecycle и Selected/Active

> Текст ниже готов к отправке целиком в новый Work-чат. Он разрешает этап только когда владелец отдельно отправит его как поручение; публикация пакета сейчас не является поручением.

Выполни **только этап 06 TECHO5 / FrankenASP V3 — production Source, lifecycle и Selected/Active**. Используй Work и authorized terminal tools на validation host.

Сначала полностью прочитай в ${WORKSPACE}/techo5-v3-design-staging/docs/external-asp-v3/:
1. EXECUTION_BRANCH_STRATEGY.md.
2. MASTER_CONTRACT.md и PCM_ONLY_ARCHITECTURE_DECISION.md.
3. stages/06_SOURCE_CONTROL_LIFECYCLE_TASK.md.
4. SPEC_V2_APPROVED.md, §§3–4, 6.6, 8–16, 18.1/18.4/18.5, 20.4–20.10, 24; нужные уточнения/решения перечислены в TASK.
5. HANDOFF_TEMPLATE.md и **принятый stages/05_HANDOFF.md**.

Проверь опубликованный docs HEAD, V2 blob actual SPEC blob from new clean-base docs commit, verified and pinned before Stage, remotes, clean worktrees и applicable AGENTS.md. Входы — exact output integration SHA из handoff 05: execution PASS, **independent delegated technical acceptance ACCEPTED** и integration COMPLETE (это техническое решение координатора, а не вымышленная дословная декларация владельца). Принятая Stage05 pair: Go `da8d1b0cf5fa768dd44f2c3f9755d4876ac5dab9`, C `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`; непосредственно перед Stage06 всё равно заново проверять remote SHA. Сверь их с remote TECHO5 experimental/external-asp-v3-dev-main и dmit2k/frankenasp-v3:experimental/v3-dev-main. Отсутствие принятого handoff или несовпадение refs означает STOP; не используй latest/незавершённую stage-ветку.

Создай Go stage06/worktree. C pin из accepted pair05, без пустой ветки; scoped C corrections только с повторной парной проверкой. Подключи Source/HA/persistence и сразу весь lifecycle/cancellation/output ownership. Принимай совместную пару, даже если C unchanged.

Разрешены изменения только scope TASK, локальные commits и push только собственных stage refs. Не merge/commit/push в integration refs, TECHO5 main или документационную ветку; приёмка и интеграция выполняются отдельно. Не начинай другие этапы.

Выполни **все tests/builds своего TASK** и запиши точные команды, результаты, pair SHA и actual files. Сохрани stock и минимальный контракт без CFG2/TLV, statuses/feedback, ASP Debug, preflight, async queue/retry, platform whitelist, inode cache или compatibility layers. Не переносить автоматически legacy code/installer/OTA/vendor assets.

Echo, аппаратные/акустические тесты и target runtime launch запрещены. При конкретном противоречии остановись с file/SHA/контрактом/repro/minimal fix, не меняя утверждённую архитектуру.

Итог верни по HANDOFF_TEMPLATE, готовый для публикации координатором в stages/06_HANDOFF.md: execution PASS/FAIL/BLOCKED, acceptance PENDING до решения владельца, inputs/tested candidates, output integration status, tests, risks/non-imports, следующий этап. Не выдумывай будущие SHA или выполненную приёмку.
