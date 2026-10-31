# 07 — Итоговая программная приёмка пары

**Внимание:** финальная приёмка проверяет новую accepted correction lineage Stage02-R/Stage03-R/Stage04–06; историческая T03/C02 пара не является PCM-only final.

**Режим:** software verification/review, без новых production features. Никакого Echo, target runtime/device launch или акустики.

## Вход и scope

Принятый `stages/06_HANDOFF.md`: T_final/C_final = output integration HEAD; acceptance ACCEPTED, integration COMPLETE. Сверить remote refs, docs SHA/V2 blob и tested candidate ancestry. Читать MASTER_CONTRACT, V2 §§20–27, REQUIREMENTS_TRACEABILITY и exact test/build commands из handoffs 02–06. Нормативные блоки читаются дополнительно по обнаруженному вопросу, не требуются все исторические аудиты.

Ответственность: финальный verification D01–D58/F1–F5/V01–V27/A1–A9; конкретные implementations принадлежат 02–06. R8/R9 объединены в один воспроизводимый software acceptance receipt.

## Worktrees и Git

- `/home/gptlab/dev/techo5-v3-stage07` detached T_final.
- `/home/gptlab/dev/frankenasp-v3-stage07` detached C_final.
- Отдельные build output directories вне tracked source; clean перед/после.
- Не создавать пустые stage07 refs/commits. Code pair не меняется.
- Найденный defect возвращается ответственному этапу с exact repro. Исправление требует принятой replacement pair в отдельных stage worktrees; не исправлять production скрыто в этой проверке. Затем повторить затронутые tests и финальный joint gate.
- Отчёт и evidence index сохраняет docs coordinator в `stages/07_HANDOFF.md` на design-staging отдельным docs commit/push. Code integration refs на 07 не продвигать.

## Программная матрица

| ID | Обязательный acceptance |
|---|---|
| S01 | Повторить codec conformance C01–C07 на final pair, включая **отсутствие IDME/metadata в descriptor**, REF L/R независимо, >2 buffers, no-REF representation, last sample/overflow и independent fixtures |
| S02 | Повторить compiled joint W01–W10: production Go session + C runtime/test-only ABI, startup >10s, shared 5s deadline, first/every output, fault/FD/stdout isolation |
| S03 | Native/privacy N01–N08 и External ownership L07/L08: muted bytes, stock behavior, retained frames через reads/IPC/fallback/toggles |
| S04 | Lifecycle/cancellation L03–L06: все 8 cancellation cases, synchronized publication, same PID/Rewire/XRUN, shutdown ordering, sole Wait/no leaks |
| S05 | Selected/Active/persistence/HA/logging L01/L02/L09, updated executable manual activation; option без helper и отсутствие retry |
| S06 | H01–H08: supported DSP fixtures vs parse-only unsupported; numeric mapping, 64 output padding, 8ms FIFO, production/mock boundary, provenance |
| S07 | `go test ./...` для default,dot,dot,donut,spot,panel в echod; targeted `go test -race` всех затронутых concurrent packages, existing stock tests включая phone/detect/wake/voice/history/raw/sensors |
| S08 | Сборка/vet по actual current upstream workflow: default,dot,dot,donut,spot,panel; Linux ARM32 GOOS=linux GOARCH=arm GOARM=7 и production C helper/launcher. Android Panel **реальные** GOOS=android NDK arm64-v8a и armeabi-v7a builds, не только linux -tags panel |
| S09 | Focused final diff/dead-code review: нет запрещённых новых ASP subsystems/compatibility/test hooks; unrelated stock mechanisms не удалены; минимальная architecture, no hidden dependencies |
| S10 | Exact pair receipt, artifact hashes, commands/exit codes/toolchain identities, LOC/file inventory, traceability results, clean worktrees, remote refs неизменны |

Перечень tags/Android commands перепроверить по T_final workflows и baseline 01. Сохраняемые supported configurations сверх таблицы также проверить, если они затронуты. Если toolchain или dependency отсутствует, зафиксировать BLOCKED конкретного software check; не выдавать Linux panel test за Android compile.

Нужны реальные tests существующей реализации, не новые AST matrices, совпадающие с названиями полей. Evidence предыдущего этапа разрешено переиспользовать только при доказанной неизменности соответствующих source/dependencies на final pair, но final compiled joint gate и build matrix выполняются на exact final commits. Описать любые skipped tests с причиной; обязательный пропуск не PASS.

## Metrics / review

Сравнить TECHO5 T_final с T01 (frozen TECHO5 main 62603a78), C_final с C01. Показать added/removed/net production LOC отдельно от tests/docs/generated artifacts, actual files/types/control fields. Legacy reference LOC можно дать отдельным контекстом, не считать отсутствующие old files «удалёнными V3».

Проверить:
- чистая родословная двух integration lines; docs branch и old implementation не merged;
- единственный Source/mono path, stock DSP/calls/wake consumers сохранены;
- codecs/secure session/native/Selected/Active только необходимого объёма;
- нет CFG2/status/preflight/Debug/feedback/whitelist/cache/synthetic/async/retry/new managers;
- C не получил installer/OTA/release tooling/proprietary assets; reuse обоснован license/provenance;
- 1100–1700 Go LOC — исторический ориентир, не квота; не урезать безопасность ради числа.

## Программная и аппаратная приёмка

**Software PASS** возможен только при успешных S01–S10 и выполненных обязательных §§20–24. **Software FAIL/BLOCKED** при обязательном failed/not-run software check; отделить baseline environmental failure от новой регрессии.

Hardware/vendor runtime на Echo, реальная IDME/clock calibration qualification, privileged device launcher behavior и акустика: **НЕ ВЫПОЛНЕНЫ**. Это отдельное будущее поручение владельца. Не требовать их для software PASS и не считать mock/crossbuild доказательством hardware PASS.

## Итоговый handoff

HANDOFF_TEMPLATE → `stages/07_HANDOFF.md`: точные T_final/C_final, неизменённые input/output refs, полный software verdict, S01–S10 и связи с D/F/V/A matrix, binary/fixture digests, build/race/security/stock evidence, реальные ограничения, hardware status, owner acceptance отдельно. Следующее действие — решение владельца о дальнейшем software review/PR либо отдельном hardware assignment; автоматического stage08/deploy нет.

STOP при нарушении контракта: exact repository/file/symbol/SHA, сценарий/команда, failed criterion, минимальное исправление и ответственный stage. Не редактировать V2, не вводить новую архитектуру, не расширять scope.
