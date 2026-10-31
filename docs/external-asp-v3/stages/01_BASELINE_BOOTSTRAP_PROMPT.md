# Стартовый промпт 01 — baseline и Git bootstrap

> Текст ниже готов к отправке целиком в новый Work-чат. Он разрешает этап только когда владелец отдельно отправит его как поручение; публикация пакета сейчас не является поручением.

Выполни **только этап 01 TECHO5 / FrankenASP V3 — baseline и Git bootstrap**. Используй Work и Desktop Commander на anfisa-trainer.

Сначала полностью прочитай в /home/gptlab/dev/techo5-v3-design-staging/docs/external-asp-v3/:
1. EXECUTION_BRANCH_STRATEGY.md.
2. MASTER_CONTRACT.md.
3. stages/01_BASELINE_BOOTSTRAP_TASK.md.
4. SPEC_V2_APPROVED.md, §§1–4, 18–19/R1, 22–27; нужные уточнения/решения перечислены в TASK.
5. HANDOFF_TEMPLATE.md, DECOMPOSITION_PLAN.md §1–2 для исходного снимка.

Проверь опубликованный docs HEAD, V2 blob actual SPEC blob from new clean-base docs commit, verified and pinned before Stage, remotes, clean worktrees и applicable AGENTS.md. Повтори remote/ancestry checks; исторические SHA из плана не подставляй как неизменные будущие bases.

Только после отдельного разрешения cleanup создать TECHO5 experimental/external-asp-v3-dev-main от фиксированного T01=62603a7854283a3a07473a17c831c6ef0fd6f8d4 и FrankenASP V3 experimental/v3-dev-main от C01=e986ae18643e08e95f29a38aae1dd74762f6362e. Оба main не менять, FF запрещён. Реализация V3 запрещена. При divergence/существующей неизвестной dev-ветке остановись без reset/merge/rebase/force push.

Разрешены только описанные bootstrap refs и baseline проверки. Сохрани их точные SHA, отдельные verification worktrees и отчёт; V3 code commits не создавай.

Выполни **все tests/builds своего TASK** и запиши точные команды, результаты, pair SHA и actual files. Сохрани stock и минимальный контракт без CFG2/TLV, statuses/feedback, ASP Debug, preflight, async queue/retry, platform whitelist, inode cache или compatibility layers. Не переносить автоматически legacy code/installer/OTA/vendor assets.

Echo, аппаратные/акустические тесты и target runtime launch запрещены. При конкретном противоречии остановись с file/SHA/контрактом/repro/minimal fix, не меняя утверждённую архитектуру.

Итог верни по HANDOFF_TEMPLATE, готовый для публикации координатором в stages/01_HANDOFF.md: execution PASS/FAIL/BLOCKED, acceptance PENDING до решения владельца, inputs/tested candidates, output integration status, tests, risks/non-imports, следующий этап. Не выдумывай будущие SHA или выполненную приёмку.
