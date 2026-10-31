# 06 — Source integration, Selected/Active, HA и lifecycle

**Внимание:** Source использует PCM-only native descriptor; не подключать никакие raw-IDME providers или дополнительные launch metadata. Vendor resources принадлежат FrankenASP launcher.

**Цель:** соединить принятые native/privacy и Go/C session components в минимальный production External path. Сразу завершить control и lifecycle, без временной preflight/confirmation/state architecture.

## Вход

T_in/C_in — accepted integrated pair из `stages/05_HANDOFF.md`; execution PASS, acceptance ACCEPTED, integration COMPLETE. Проверить exact remote refs и pair test receipt. Читать strategy, MASTER_CONTRACT, данный TASK, V2 §§3–4, 6.6, 8–16, 18.1/18.4/18.5, 20.4–20.10, 24; D01–D03, D10–D22, D25–D27, D35–D49, D51/D55–D58; F1/F3/F4/F5; A6/A7/A8.

Native API и metadata из 03, session/secure launch/Wait contract из 05 — входные артефакты, не повод перепроектировать codec.

## Реальные code seams

- `echod/internal/hardware/mic/mic.go`: Source, Run, broadcast, owned stock frame, remember/spectrum/listeners, Close.
- `mic/gain.go`: Rewire и stock supervisor restart.
- `echod/internal/boot/boot.go`: долгоживущий Run/run, inner context и update-triggered stop; `service/group.go`: wg.Wait прежде Close.
- `config/microphone.go` и `feature/microphone/microphone.go`: immediate settings persistence, HA build/Restore/Entities.
- `feature/phone/call.go`: stock callCanceller/CallCancelEngine; `mic/cancel.go`/webrtc.go и raw subscribers сохраняются.

Новые ASP файлы — только небольшие session/control binding рядом с этими seams. Прежних asp_control/feedback/log файлов в baseline нет; не переносить их для удаления.

## Обязательное поведение

1. Helper принадлежит External session в долгоживущем Source/echod lifetime, не capture Run. Startup вне capture loop; stock Built-in обрабатывает frames в ожидании READY.
2. Вызвать sync exchange на borrowed native input до следующего read и **вне publication mutex**. Сохранить stock conversion/DSP целиком и sensors на stock data.
3. Одна точка выбора mono после stock DSP, до remember/spectrum/listeners. External ответ допустим только полный и от текущего session/helper, при Selected External и неотменённой session. Иначе stock. Не менять ListenRaw/SIP/intercom.
4. Проверка identity и окончательная публикация синхронизированы с переключением режима; cancellation не может вклиниться между check и publish так, чтобы прошёл stale frame. Старый launch/cleanup не активирует и не закрывает новую session. Нет generation/age tracking.
5. Сохранить existing stock owned frame. Для External использовать owned PCM или одну необходимую copy перед общей публикацией; не публиковать reused scratch. Подписчик может удерживать A во время B/C/exchanges/fallback.
6. Единственные смысловые состояния Selected/Active. Selected и config persist немедленно на user choice независимо от launch/READY/PCM. До первого полного успешного exchange Active Built-in; после первого допустим External. Fallback меняет только Active, закрывает виновную session, пишет причину.
7. HA всегда предлагает Built-in/External независимо от executable при startup. Установка/замена helper замечается при ручном Built-in→External без echod restart. Нет auto retry/scanning/watchers. Restore сохранённого Selected запускает только обычную попытку при доступном capture/session setup, не retry на каждом Run.
8. HA создаёт только selection и фактический Active; нет ASP Debug/health/status attributes. Три типа новых ASP host events: launch, fallback с причиной, termination; stdout/stderr helper не парсить.
9. Source.Close и выход Run закрывают/останавливают **capture**, не helper. Rewire сохраняет PID/session/pipes/FIFO, не отправляет descriptor/READY заново. Штатный ALSA ErrOverrun делает stock continue; helper не restart. После reopen нет descriptor/geometry revalidation.
10. Не смешивать новый helper lifecycle со stock service.Restart. Capture restart допустим; auto retry отказавшего External запрещён. Пауза capture не timeout PCM exchange, если exchange не выполняется.
11. Final shutdown hook обязан закрыть допуск новых launch/capture operations и отменить launch/exchange/закрыть pipes **до** ожидания Run, затем завершить child и дождаться sole reaper. В current boot.run cancellation может прийти от signal или update; учесть также early service error. Не ставить единственный cleanup в defer после group.Run или Source.Close. Допустим один прямой hook/привязка к существующему долгоживущему контексту; generic lifecycle framework не нужен.
12. Никакого второго SoftwareCut после DSP, FIR/history resets, preflight, first-PCM deadline или delayed persistence. Не удалять unrelated stock logs/queues/WebRTC/platform logic.

## Проверки своего этапа

| ID | Обязательные scenarios / PASS |
|---|---|
| L01 | Built-in продолжает frames во время slow startup/READY; Active Built-in до первого successful PCM, становится External с первого, Selected persisted немедленно |
| L02 | Faults launch/missing executable/validation/READY/EOF/timeout/process exit: Active Built-in, Selected External; ручной retry работает, auto retry отсутствует |
| L03 | Cancellation во время launch, ожидания READY, PCM write, PCM read; старый launch/response приходит позже; новая session создана до old completion; old cleanup не затрагивает её — все 8 случаев §20.6 |
| L04 | Publication race: mode switch конкурирует с final response check/publish; stale response не попадает ни в history, ни MWW/STT/других mono listeners; subscription/control не заблокированы IPC mutex |
| L05 | Rewire/Source.Close/Run exit сохраняют PID/session; reopen возобновляет обмен; count descriptor/READY/vendor init/padding не растёт. XRUN сохраняет helper. Пауза >5s и >10s вне exchange не закрывает session |
| L06 | Signal/update/early-error final shutdown: cancel/pipe close раньше ожидания Run, один Wait, закрыты FDs, child reaped; блокированный exchange не удерживает teardown |
| L07 | Subscriber удерживает frame A через следующие reads/exchanges B/C, fallback и toggles; A неизменен. Один output во всех mono consumers, ListenRaw сохранён |
| L08 | Native SoftwareCut/MIC+REF перед helper, stock mute, hardware REF semantics и допустимые переходы; нет второго policy check/сброса DSP/FIFO |
| L09 | HA option без helper, immediate persistence/restore/runtime fallback, поздняя установка/replacement, только Selected/Active и три ASP event types; нет Debug/status/feedback |
| L10 | Stock mixer/AEC/denoise/leveling/sensors, MWW/STT/history/raw, calls/intercom и callCanceller restore/overlapping calls; новые wake front-end APIs не откатились |
| L11 | Relevant Go tests + race; compiled paired Go/C regression 05 и C FIFO preservation; оба production builds, supported tag checks для затронутых paths |

Для L05 можно instrument test C binding и fake capture только в tests; assertions наблюдают PID/handshake/padding behavior, не требуют аппаратуры. L07 обязательно проверяет **External** output, stock-only test 03 недостаточен.

## Git / handoff

Go worktree `/home/gptlab/dev/techo5-v3-stage06`, branch `experimental/external-asp-v3-dev-stage06` от T_in. C worktree `/home/gptlab/dev/frankenasp-v3-stage06` detached C_in для joint tests. Если доказанно требуется C change, создать `stage06` от C_in и согласовать scope с этим TASK; вне scope — STOP, не расширять DSP.

Локальные commits Source/control/lifecycle/tests, push stage refs только по поручению. Integration только отдельной owner acceptance **пары** T_candidate/C_candidate и поручением coordinator, даже когда C unchanged. До проверки обоих remote refs handoff integration PENDING. Docs branch не merge в code.

Handoff `stages/06_HANDOFF.md` по HANDOFF_TEMPLATE через docs coordinator: inputs, tested pair, actual Source/API/hook paths, lock/ownership rationale, L01–L11 commands/results, regression/build evidence, LOC/non-import, acceptance/integration SHA. Hardware НЕ ВЫПОЛНЕНЫ. Следующий — 07.

**PASS:** полный External path и все L checks; нет отложенных privacy/concurrency fixes. **FAIL:** один лишь wire PASS без Source/HA, stale publication, close helper на Rewire, second Wait, отложенная persistence, нарушение stock. STOP по exact file/SHA/contract/scenario/minimal fix; approved architecture не открывать заново.
