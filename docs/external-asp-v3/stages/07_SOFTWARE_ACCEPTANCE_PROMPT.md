# Стартовый промпт 07 — финальная программная приёмка

> Текст ниже готов к отправке целиком в новый Work-чат. Он разрешает этап только когда владелец отдельно отправит его как поручение; публикация пакета сейчас не является поручением.

Выполни **только этап 07 TECHO5 / FrankenASP V3 — финальная программная приёмка**. Используй Work и Desktop Commander на anfisa-trainer.

Сначала полностью прочитай в /home/gptlab/dev/techo5-v3-design-staging/docs/external-asp-v3/:
1. EXECUTION_BRANCH_STRATEGY.md.
2. MASTER_CONTRACT.md и PCM_ONLY_ARCHITECTURE_DECISION.md.
3. stages/07_SOFTWARE_ACCEPTANCE_TASK.md.
4. SPEC_V2_APPROVED.md, §§20–27; нужные уточнения/решения перечислены в TASK.
5. HANDOFF_TEMPLATE.md и **принятый stages/06_HANDOFF.md**.

Проверь опубликованный docs HEAD, V2 blob actual SPEC blob from new clean-base docs commit, verified and pinned before Stage, remotes, clean worktrees и applicable AGENTS.md. Входы — exact output integration SHA из handoff 06: execution PASS, independent delegated technical acceptance ACCEPTED и integration COMPLETE (решение координатора, не вымышленная дословная декларация владельца). Принятая Stage06 пара: TECHO5 Go cce2f220aeab9835e10a7f7ab659f5ca390011e8, FrankenASP V3 C f6266907edb90e7c9c2f414e8e46e6f215a21b7b. Локальный tested Go commit cc7254208991dbbba002394f9c2c18f6dda35b28 отличается от опубликованного, но дерево SHA 09f798b6d9515f5de7dc1e7e4d2fd235f268e995 совпало; использовать только remote integration SHA и перед работой заново проверить refs. Сверь их с remote TECHO5 experimental/external-asp-v3-dev-main и dmit2k/frankenasp-v3:experimental/v3-dev-main. Отсутствие принятого handoff или несовпадение refs означает STOP; не используй latest/незавершённую stage-ветку.

Используй отдельные detached stage07 worktrees на принятой паре06. Production не менять, пустых stage refs/commits не создавать. Выполни final software matrix и верни дефекты ответственному этапу. Hardware не запускать.

Code refs не изменяй. Подготовь отчёт для документационного координатора; программный PASS и аппаратное НЕ ВЫПОЛНЕНЫ раздели.

Выполни **все tests/builds своего TASK** и запиши точные команды, результаты, pair SHA и actual files. Сохрани stock и минимальный контракт без CFG2/TLV, statuses/feedback, ASP Debug, preflight, async queue/retry, platform whitelist, inode cache или compatibility layers. Не переносить автоматически legacy code/installer/OTA/vendor assets.

Echo, аппаратные/акустические тесты и target runtime launch запрещены. При конкретном противоречии остановись с file/SHA/контрактом/repro/minimal fix, не меняя утверждённую архитектуру.

Итог верни по HANDOFF_TEMPLATE, готовый для публикации координатором в stages/07_HANDOFF.md: execution PASS/FAIL/BLOCKED, acceptance PENDING до решения владельца, inputs/tested candidates, output integration status, tests, risks/non-imports, следующий этап. Не выдумывай будущие SHA или выполненную приёмку.
