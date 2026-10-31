# 01 — Baseline и отдельный Git bootstrap

**Режим будущего этапа:** baseline/Git administration, без реализации V3. Исполнять только после отдельного поручения владельца на 01, явно разрешающего bootstrap только после очистки. Подготовка этого TASK таким поручением не является.

## Цель и границы

Получить проверенные чистые bases двух проектов и карту актуальных stock seams до первого code stage. Проверить доступность software build/test environment. Не переносить старые ASP-файлы.

Основание: EXECUTION_BRANCH_STRATEGY §§1–5; V2 §§1–4, 18–19/R1, 22–27; D01–D09, D23–D24, D54–D58; F1/F5. Новые исходники должны происходить из upstream и независимого frankenasp-v3, tests выполняются offline, hardware отдельно.

## Входные ревизии

Docs checkout: `/home/gptlab/dev/techo5-v3-design-staging`, branch `experimental/v3-design-staging`. Определить опубликованный docs HEAD, прочитать strategy, MASTER_CONTRACT и этот TASK; проверить V2 blob `new SPEC blob to be pinned from the single clean-base documentation commit`.

Фиксированные чистые исходные точки:
- fork main: `62603a7854283a3a07473a17c831c6ef0fd6f8d4`;
- upstream: только read-only reference, не источник FF;
- new C main: `e986ae18643e08e95f29a38aae1dd74762f6362e`.

Повторно проверить remote refs, tags, worktrees, clean status и ancestry. TECHO5 main должен строго равняться T01=62603a7854283a3a07473a17c831c6ef0fd6f8d4, FrankenASP V3 main — C01=e986ae18643e08e95f29a38aae1dd74762f6362e. Оба main остаются неизменными Stage01–07. До Stage01 нужен отдельно утверждённый cleanup прежних refs/worktrees.

## Работа

1. Проверить hostname, remotes, branch/worktree inventory, clean status, applicable AGENTS.md. Не переключать существующий чужой checkout. Подтвердить, что C origin — именно `dmit2k/frankenasp-v3`, не `dmit2k/FrankenASP`.
2. Снять before refs и чистоту. Сравнить current upstream с предыдущим снимком и old reference только для анализа. Основные проверенные пути — таблица §2 DECOMPOSITION_PLAN: Source.Run/broadcast/Close, pairedCapture.Read/convert, Rewire, boot.run, Group.stop, config/HA, consumers.
3. Проверить отсутствие экспериментальной ASP-системы в upstream. Stock callCanceller/CallCancelEngine, wake-model front end, service restart, WebRTC, Panel не считать legacy.
4. Проверить и зафиксировать, что TECHO5 main == T01 и FrankenASP V3 main == C01. Никаких FF, reset или продвижения main в Stage01–07.
5. Создать TECHO5 integration experimental/external-asp-v3-dev-main строго от T01 после согласованного cleanup, worktree отдельно. Если ветка уже занята — STOP без reset.
6. Создать FrankenASP V3 integration experimental/v3-dev-main строго от C01 после согласованного cleanup. Main не назначать integration ref, не использовать старый stage SHA.
7. Для isolated baseline проверки создать **detached** worktrees `/home/gptlab/dev/techo5-v3-stage01` на T01 и `/home/gptlab/dev/frankenasp-v3-stage01` на C01. Не создавать пустые code stage01 refs.
8. Зафиксировать stock geometry: MIC/REF buffers, rates/formats/valid bits/shift, conversion boundary, REF absent/stock soft reference, metadata source availability. Не запускать hardware readers на Echo.
9. Определить доступные toolchains и точные offline команды по текущим go.mod/workflows и C reference build requirements. Для C здесь только readiness inventory, C-кода ещё нет. Private runtime/assets не скачивать, не копировать в Git, не запускать.
10. Выполнить baseline Go tests/builds в проверяемом объёме; записать existing failures отдельно. Снять production LOC/files по T01 и C01 и ориентир старого reference отдельно, без искусственных «удалений».

## Проверки этого этапа

| ID | Проверка / критерий |
|---|---|
| B01 | Remote refs, ancestry и fixed main SHA документированы; dev-main == frozen clean main; V3 code patches отсутствуют |
| B02 | Карта актуальных symbols/stock dependencies и native/metadata seams подтверждена git show/diff, без assumptions из old reference |
| B03 | Stock baseline: из echod выполнить `go test ./internal/hardware/mic ./internal/hardware/privacy ./internal/config ./internal/feature/microphone ./internal/service ./internal/boot`; дополнить actual package set из current tree |
| B04 | Сверить tags/default,dot,dot,donut,spot,panel и ARM32/Android build commands с workflows; выполнить доступные baseline tests/builds, записать compiler versions; отсутствие обязательной зависимости явно BLOCKED |
| B05 | C main frozen initial; C dev-main from C01; inventory необходимого ARM32 Linux/Bionic toolchain без переноса installer/OTA/assets; old C остаётся reference |

Baseline test не может обращаться к реальному capture, HA или Echo: отделить device-tagged tests от offline. Не запускать target binaries. Если штатный test требует hardware, отметить и использовать штатную offline конфигурацию, не переписывать тест ради baseline PASS.

## Git, результат и передача

01 — административное исключение: не требуется пустой code commit. Реальные output integration IDs T01/C01 фиксируются после публикации разрешённых bootstrap refs; docs coordinator сохраняет `stages/01_HANDOFF.md` отдельным docs commit.

Отчёт строго по HANDOFF_TEMPLATE: входные remote main, upstream, fixed T01, C01, docs SHA/blob, команды ancestry, созданные refs/worktrees, stock paths/test inventory, LOC baseline, toolchain prerequisites, execution PASS/FAIL, owner acceptance и фактические remote refs. Hardware = **НЕ ВЫПОЛНЕНЫ**. Следующий этап — 02, только после принятия 01.

**PASS:** bases чистые, clean bases подтверждены, две integration lines проверены, seams и воспроизводимые baseline checks записаны. **FAIL/STOP:** divergence, занятый ref неизвестного происхождения, dirty чужое дерево, неподтверждённый C delta, отсутствующие обязательные prerequisites без плана устранения. Не подменять этот результат кодовыми исправлениями.

**Не делать:** V3 implementation, dev-stage branches для будущих этапов, merge docs/old reference, vendor runtime/device tests, installer/OTA или новые platform APIs. При конфликте с контрактом — file/SHA/scenario/minimal fix и STOP по MASTER_CONTRACT §11.
