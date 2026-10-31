# Стартовый промпт 07 — финальная программная приёмка

> Текст ниже готов к отправке целиком в новый Work-чат. Он разрешает этап только когда владелец отдельно отправит его как поручение; публикация пакета сейчас не является поручением.

Выполни **только этап 07 TECHO5 / FrankenASP V3 — финальная программная приёмка**. Используй Work и Desktop Commander на anfisa-trainer.

Сначала полностью прочитай в /home/gptlab/dev/techo5-v3-design-staging/docs/external-asp-v3/:
1. EXECUTION_BRANCH_STRATEGY.md.
2. MASTER_CONTRACT.md.
3. stages/07_SOFTWARE_ACCEPTANCE_TASK.md.
4. SPEC_V2_APPROVED.md, §§20–27; нужные уточнения/решения перечислены в TASK.
5. HANDOFF_TEMPLATE.md и **принятый stages/06_HANDOFF.md**.

Проверь опубликованный docs HEAD, V2 blob actual SPEC blob from new clean-base docs commit, verified and pinned before Stage, remotes, clean worktrees и applicable AGENTS.md. Входы — exact output integration SHA из handoff 06, только если execution PASS, owner acceptance ACCEPTED и integration COMPLETE. Сверь их с remote TECHO5 experimental/external-asp-v3-dev-main и dmit2k/frankenasp-v3:experimental/v3-dev-main. Отсутствие принятого handoff или несовпадение refs означает STOP; не используй latest/незавершённую stage-ветку.

Используй отдельные detached stage07 worktrees на принятой паре06. Production не менять, пустых stage refs/commits не создавать. Выполни final software matrix и верни дефекты ответственному этапу. Hardware не запускать.

Code refs не изменяй. Подготовь отчёт для документационного координатора; программный PASS и аппаратное НЕ ВЫПОЛНЕНЫ раздели.

Выполни **все tests/builds своего TASK** и запиши точные команды, результаты, pair SHA и actual files. Сохрани stock и минимальный контракт без CFG2/TLV, statuses/feedback, ASP Debug, preflight, async queue/retry, platform whitelist, inode cache или compatibility layers. Не переносить автоматически legacy code/installer/OTA/vendor assets.

Echo, аппаратные/акустические тесты и target runtime launch запрещены. При конкретном противоречии остановись с file/SHA/контрактом/repro/minimal fix, не меняя утверждённую архитектуру.

Итог верни по HANDOFF_TEMPLATE, готовый для публикации координатором в stages/07_HANDOFF.md: execution PASS/FAIL/BLOCKED, acceptance PENDING до решения владельца, inputs/tested candidates, output integration status, tests, risks/non-imports, следующий этап. Не выдумывай будущие SHA или выполненную приёмку.
