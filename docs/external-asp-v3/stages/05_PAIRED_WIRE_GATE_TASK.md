# 05 — Secure host session и первый paired Go/C wire gate

**Внимание:** `04_HANDOFF.md` должен быть создан по PCM-only архитектуре; Go/C codec header=16 bytes, никаких IDME/metadata sections. Проверить resource prep launcher отдельно от PCM pipes.

**Цель:** построить настоящий Go session/IPC runtime и принять его со скомпилированным новым C helper как точную пару commits. Production `mic.Source` пока не включает External; это делает 06 без временной state machine.

## Вход и чтение

Принятый `stages/04_HANDOFF.md`: T_in/C_in = output integration HEAD, acceptance ACCEPTED, integration COMPLETE. T_in содержит codec/native/privacy, C_in — новый runtime/helper/build. Проверить обоих remotes и actual commands/paths из handoff. Старый FrankenASP не является runtime-стороной.

V2 §§5–7, 11.5–11.6, 12–14, 16–17, 20.1–20.3, 20.6–20.9, 21–22; D10–D11, D25–D34, D40–D42, D50–D55/D57–D58; F2–F4; A1/A5–A8.

## Реализовать в Go

Новый компактный process/session component рядом с mic integration; codec 02 переиспользовать. Точные файлы выбрать в новом дереве и записать. Source/HA/config не подключать здесь, но production module уже готов для direct use из 06; test-only helper injection не должен стать публичным production API.

1. Secure launch: literal approved /data installation path, traversal через pinned directory FDs, O_NOFOLLOW, owner/mode/type checks, check/use protection. FD3 executable, FD4 assets; не re-resolve проверенный executable по pathname.
2. Reference `asp_secure_open_linux.go` и `asp_pinned_exec_linux.go` на 9ef0792a… содержат проверенные trust rules. Сохранить точные исключения для известных ancestors /data и /data/techo5-linux, если нужны этому install contract; не расширять writable-path trust. Удалять дубли можно только после доказательства границы. Не импортировать old wrappers/test-only production entrypoints/inode cache.
3. Каждая новая ручная session заново открывает установку. Нет sync.OnceValue/lifetime inode pin. Проверить обновление executable/assets между двумя сессиями без restart host.
4. Process stdin/stdout — PCM pipes; stderr обычный отдельный поток, без collector/parser. Согласовать C launch contract из 04, сохранить минимальное безопасное environment без loader overrides.
5. Startup отдельным конечным 10s budget: preparation→Start→descriptor→READY. После READY снять startup deadline, не отменять lifetime context child. Session-ready не означает Active=External; это решение Source 06 после первого полного PCM.
6. Exchange строго sync: последовательно writeFull всех native buffers, затем readFull 640 bytes, один общий deadline now+5s на все операции. Нет async PCM worker, headers, sequence/origin, idle timer или queue.
7. Borrowed input больше не используется после возврата exchange. Output ownership API явно описать: scratch нельзя отдавать подписчикам без stable ownership; 06 обеспечит одну необходимую копию/owned output. В 05 проверить, что documented lifetime соблюдён.
8. Любая ошибка/отмена закрывает реальные pipes и child; успешный Start имеет **ровно одного Wait owner/reaper**. Close idempotent, повторное ожидание reaper не вызывает второй Wait. Release launch FDs на всех fail branches. Устаревший session object может закрыть только собственные ресурсы.
9. Session API даёт 06 cancellation без ожидания блокирующего exchange. Финальный owner/identity/publication и Rewire binding остаются в 06; не создать временный global manager.
10. Если здесь создаются ASP logs, только launch/fallback/termination с понятной причиной; fallback Active/Selected делает 06. Не логировать READY/каждый frame и не создавать status codes.

## Согласованная C сторона

Использовать runtime из C_in; C не менять ради формального второго commit. Если совместный test обнаружит нужную correction codec/wire/FIFO, scope относится к этому paired gate: отдельный `stage05`, C commit, повторить C local tests и оба integration builds. Любая согласованная пара фиксирует реальный C SHA, даже если он равен C04.

## Tests этого gate

| ID | Обязательное доказательство |
|---|---|
| W01 | Native integration binary использует production Go session и compiled C codec/runtime/bridge с test-only DSP binding; реальные OS pipes, shared fixture manifest, никакого Python/fake wire вместо C |
| W02 | Descriptor→однократное READY-подтверждение строго по принятому Stage02 Go/C fixture (проверить точные bytes и длину, не предполагать 5 ASCII bytes); поддерживаемый fixture проходит; валидный unsupported descriptor отклонён до READY; каждый input даёт ровно 640 bytes, first output содержит ожидаемый one-time padding |
| W03 | Partial reads/writes и EINTR на C, EOF/short response/broken pipe/process exit; нулевой progress не вечный loop; один outstanding exchange без framing |
| W04 | Общий 5s budget не перезапускается на каждом buffer/read/write; blocked write и blocked read отменяются; пауза между обменами >5s не закрывает helper |
| W05 | Startup timeout до READY; после READY продолжить exchanges после >10s от старта — child жив; startup cancellation не убивает новую/готовую session |
| W06 | Single Wait и no zombies/FD leaks на success, Start failure, READY error, timeout/cancel/EOF; late old-session cleanup не закрывает новую session resources |
| W07 | Secure path tests: symlinks/components/rename race, wrong owner/mode/type, executable/assets FD pinning, fixed env; valid installation и replacement между двумя активациями |
| W08 | FD3/FD4 действительно executable/assets, PCM идёт через 0/1 и C wire FD, stdout chatter не портит protocol |
| W09 | Relevant Go race tests, codec/native tests и C tests; production TECHO5 Linux ARM32 и C production helper/launcher сборки из exact commits; hashes/toolchains |
| W10 | Проверка absence CFG2/old helpers/compatibility/status/preflight/first-PCM/inode cache/new manager/async worker; mocks только test build |

Tests используют isolated temporary trusted-root fixtures; не изменять настоящий /data и не запускать privileged C launcher на validation host/Echo. Это проверяет host secure exec + compiled C IPC, а production native launcher/vendor execution остаётся явно hardware-unverified. Не требовать аппаратный запуск для software wire PASS.

Timed tests должны доказывать 5s total и >10s survival; допустимы bounded test clock seams без production time configurator, но минимум один compiled process тест подтверждает survival после фактических исходных 10s.

## Worktrees, Git и атомарность приёмки

- Go: `${WORKSPACE}/techo5-v3-stage05`, `experimental/external-asp-v3-dev-stage05` от T_in.
- C: `${WORKSPACE}/frankenasp-v3-stage05` detached C_in для сборок; branch `experimental/v3-stage05` только если реально нужны C code/test changes.
- Сделать локальные Go commits (и C при изменениях); затем пересобрать/проверить exact candidate commits без незакоммиченного code diff. Stage push только в пределах отдельного поручения, никаких автоматических main/dev-main merge.

**Gate принимает пару T_candidate+C_candidate**, а не две независимые половины. Owner acceptance пары отдельно от execution PASS. Integration agent отдельным поручением продвигает changing refs, проверяет оба remote integration HEAD и заполняет T05/C05. Пока хотя бы один ref не проверен, integration PENDING и 06 запрещён. Документационная ветка в код не сливается.

## Handoff / PASS / STOP

По HANDOFF_TEMPLATE → docs coordinator → `stages/05_HANDOFF.md`: два входных SHA, два tested candidate SHA, exact commands/results W01–W10, обе production builds, native/mock vs production binary hashes, fixture digest, actual session API/lifetimes/close-wait contract, changed files, non-import inventory, owner decision и integration pair. Явно: Source/HA activation ещё не выполнена; hardware НЕ ВЫПОЛНЕНЫ. Следующий — 06.

**PASS:** вся W-матрица подтверждена на одной паре. **FAIL:** Go+C не запускались совместно, отсутствует production build одной стороны, старый C/compatibility, READY child погибает по startup timer, deadline на каждую syscall отдельно, double Wait или утечки. При противоречии STOP: file/SHA/пункт/scenario/minimal fix; не менять архитектуру.
