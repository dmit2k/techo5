# Шаблон передачи этапа External ASP V3

Копировать при исполнении в отчёт; docs coordinator публикует как `stages/NN_HANDOFF.md` в `dmit2k/techo5:experimental/v3-design-staging`. Кодовый исполнитель не коммитит в docs checkout. Этот шаблон не является выполненным handoff и не разрешает старт этапа.

## 1. Идентичность и статусы

- Stage ID / TASK / дата (Europe/Riga) / исполнитель:
- Docs commit, использованный для TASK/MASTER_CONTRACT/strategy:
- V2 blob: `SPEC_BLOB_FROM_NEW_APPROVED_DOCS_COMMIT` — фактическая проверка:
- Execution result: **PASS / FAIL / BLOCKED**.
- Owner acceptance: **PENDING / ACCEPTED / REJECTED**, точная ссылка/дата/текст явного решения:
- Integration: **PENDING / COMPLETE**, исполнитель отдельного integration поручения:
- Software gate level: baseline / codec component / native component / C component / paired wire / full Source / final software.
- Hardware tests: **НЕ ВЫПОЛНЕНЫ**.
- Acoustic tests: **НЕ ВЫПОЛНЕНЫ**.
- Vendor runtime on Echo: **НЕ ВЫПОЛНЕН**.

PASS тестов не равен owner acceptance. Следующий stage требует execution PASS + owner ACCEPTED + integration COMPLETE; для 01 также явно разрешённый и проверенный bootstrap.

## 2. Ревизии

Все SHA полные 40 символов. На неизменённой стороне повторить реальный SHA и написать unchanged; не писать «latest», «main» вместо SHA и не создавать пустой commit.

| Поле | TECHO5 | frankenasp-v3 |
|---|---|---|
| Origin URL | dmit2k/techo5 | dmit2k/frankenasp-v3 |
| Integration ref | experimental/external-asp-v3-dev-main | experimental/v3-dev-main |
| Input accepted integration SHA | Заполнить | Заполнить |
| Previous accepted handoff docs SHA/path | Заполнить | Заполнить |
| Stage branch (или detached/не требовалась) | Заполнить | Заполнить |
| Worktree absolute path | Заполнить | Заполнить |
| Stage commits в порядке | Заполнить | Заполнить |
| Tested candidate SHA | Заполнить | Заполнить |
| Owner-accepted candidate SHA | Заполнить после решения | Заполнить после решения |
| Output integration SHA | Заполнить после integration | Заполнить после integration |
| Remote verified SHA / время проверки | Заполнить | Заполнить |
| Clean worktree evidence | Команда/результат | Команда/результат |

Для 01 дополнительно: fork main before, upstream remote before, ancestry command/exit, frozen TECHO5 main 62603a78, dev-main start SHA, C initial/current delta, remote before/after, отсутствие V3 code patches.

Для 07 output pair = input pair, code refs неизменны; документальная приёмка получает собственный docs commit.

## 3. Согласованная Go/C пара

Обязательно для 05, 06, 07 и любого изменения IPC; 02 тоже фиксирует codec pair, но не называет её runtime integration.

- Pair: **T_SHA = … ; C_SHA = …** (заполнить только фактическими IDs).
- Команда получения source/fixture manifests, SHA256:
- Native integration Go test binary / compiled C runtime identity:
- Production TECHO5 ARM32 build hash / C helper и launcher hashes:
- Toolchain versions / build tags / target ABI:
- Совместный запуск, команды, cwd, exit codes, artifacts:
- Какая DSP binding использовалась в offline test; что реально собрано production:
- Совместимые supported fixtures / unsupported rejection evidence:
- Предел доказательства: codec vs wire vs Source vs hardware:
- Подтверждение: old FrankenASP не использован как новая runtime сторона.
- После приёмки: какие changing refs продвинуты, какие unchanged, обе remote проверки:
- При частичной публикации: integration PENDING, следующая стадия заблокирована; **без автоматического rollback/force push**.

Receipt действителен только для указанной пары. Если SHA изменился, связать новый test evidence с новой парой.

## 4. Изменения и сохранность

| Repo / actual file | Added/changed/deleted | Назначение | Requirement IDs |
|---|---|---|---|
| Заполнить | Заполнить | Заполнить | D/F/V/A |

- Public interfaces/lifetimes, которые использует следующий stage:
- Stock functionality, проверенная как сохранённая:
- Реально удалённые files/types (или «не было; удалений нет»):
- Неперенесённые legacy subsystems, с evidence:
- Reference reuse: repo/SHA/path/фрагмент, применимая license/provenance:
- Proprietary assets/installer/OTA/release automation: подтверждение отсутствия:
- Production LOC added/removed/net; baseline T01/C01, метод подсчёта; tests/docs отдельно:

## 5. Tests и builds

| Test ID | Command + cwd + environment (без secrets) | Tested SHA pair | Exit / PASS/FAIL/BLOCKED | Evidence path/hash |
|---|---|---|---|---|
| B/C/N/H/W/L/S | Заполнить точную команду | Заполнить | Заполнить | Заполнить |

- Relevant unit tests:
- Independent Go/C codec conformance:
- Compiled Go/C integration:
- Native/FIFO/SoftwareCut/output-retention:
- Cancellation/race/Rewire/XRUN/sole Wait:
- Secure launch и executable replacement:
- HA/persistence/stock consumers/calls:
- Supported tags/Linux ARM32/Android NDK builds:
- Not-run software checks и причины (обязательный пропуск не PASS):
- Existing baseline failures отдельно от новых:
- Hardware/acoustic/vendor-device checks: **НЕ ВЫПОЛНЕНЫ**, если нет отдельного явного поручения.

## 6. Отклонения / STOP

Для каждого blocker: repo + SHA + file/symbol, нарушенный пункт, точная команда/сценарий, observed/expected, минимальное исправление, ответственный stage. Не вносить архитектурные решения в handoff как будто они утверждены. Если отклонений нет — явно «нет».

Реальные ограничения, оставшиеся вне software evidence (mock vs vendor, privileged device launch, акустика): перечислить без ложного аппаратного PASS.

## 7. Передача дальше

- Следующий разрешённый по плану stage:
- Какие готовые paths/APIs/build commands он получает:
- Какие exact output integration HEAD он должен проверить:
- Required owner acceptance/integration decision завершены или ожидаются:
- Hardware stage/deploy автоматически **не разрешены**.

## 8. Публикация координатором

Docs coordinator сверяет stage report, decision и remote refs; заполняет output integration SHA и acceptance поля **после** фактических действий. Commit/push только docs branch; проверяет опубликованный файл и clean worktree. Не merge документацию в code. Итоговый docs publication SHA сообщается отдельно, без самоссылочного будущего SHA внутри собственного commit.
