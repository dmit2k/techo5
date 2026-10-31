# Межэтапный контракт чистой External ASP V3

Статус пакета: **OWNER ACCEPTED 2026-10-31**. Stage01 принят, Stage02–07 выполняются отдельными поручениями. Контракт не расширяет согласованную архитектуру.

## 1. Авторитет и вход

Читать [EXECUTION_BRANCH_STRATEGY.md](EXECUTION_BRANCH_STRATEGY.md), этот файл, конкретный TASK, указанные в TASK разделы [SPEC_V2_APPROVED.md](SPEC_V2_APPROVED.md) и принятый handoff. Полные исторические аудиты повторно читать не требуется.

Приоритет: 58 решений с пятью финальными уточнениями → исправленная SPEC V2 для V3 restart (принята владельцем 2026-10-31); Git и clean implementation — по EXECUTION_BRANCH_STRATEGY. Исторический AUDIT_V2_FINAL относится к прежней normalized спецификации и не утверждает raw-IDME поправку. Фактический Git blob текущей SPEC_V2_APPROVED.md: `b324d2161e5ea316618f23bbe6d6aedf6c2df5c8`.

Снимок исходных ревизий и проверенных точек интеграции находится в [DECOMPOSITION_PLAN.md](DECOMPOSITION_PLAN.md). Нельзя использовать документационное дерево или старый эксперимент как production baseline.

## 2. Минимальная система

TECHO5 владеет hardware capture, чтением фактических metadata/доступной заводской калибровки, stock conversion/DSP/sensors, выбором режима, безопасным запуском, синхронным IPC, fallback и общей mono publication внутри `mic.Source`.

Helper владеет qualification своего DSP, MIC/REF mapping, DSP-specific format/rate/quantum adaptation, calibration interpretation, DSP и FIFO. Сохраняются подтверждённые конфигурации; новых resamplers, metadata providers и расширения vendor compatibility нет. Представимость descriptor не означает DSP support.

Built-in остаётся полным и работает без helper. Mixer/AEC/denoise/leveling не отключать ради CPU. Noise/facing/standout/levels получают штатные данные. History/MWW/STT/spectrum/mono subscribers получают один выбранный поток. `ListenRaw()`, SIP/intercom и stock WebRTC AEC switching сохраняются.

## 3. Native PCM и descriptor

Host передаёт фактические buffers до скрытой stock conversion. Там, где Read уже возвращает native bytes, использовать их напрямую. Для paired capture — минимальный необязательный native tap до convert, с сохранением обычного Read. Имена новых API определяет исполнитель по принятому коду, новая иерархия adapters не нужна.

Не разделять interleaved MIC/REF, не менять channel order/format/rate и не собирать объединённый input ради DSP. Borrowed buffers используются синхронно до следующего read.

Descriptor v1 один на External-сессию:
- little-endian, magic ASP1 только в descriptor; header/buffer/channel records = 24/28/8 bytes;
- список физических buffers в порядке записи, непрерывные offsets, сумма Length = total input bytes;
- S16LE, packed S24_3LE, S32LE; signed valid bits включают знак, LSB shift описывает выравнивание;
- captured MIC/REF role indices; отсутствие REF — отсутствие REF records;
- platform/factory IDME доступны только как raw byte-preserving source values §5.3.4; Go не создаёт normalized PlatformInfo/MicFactoryGain и не вычисляет Q30, masks, confidence, availability или captured MIC gain mapping;
- bounded descriptor size обосновывается реальными raw metadata sources в Stage02, не предел PCM. Нет generic host лимита двумя buffers либо MIC4/REF2/96 kHz;
- проверяются sizes/overflow/last sample/channel coverage/reserved bytes; DSP admission только у helper.

Не выдумывать metadata, unity gain или физическую геометрию. Factory IDME gain semantics и MIC-slot association находятся исключительно в helper. Host не вычисляет applied_to_pcm, confidence, masks или platform admission всего host. Vendor Calibrator остаётся авторитетом применения заводской калибровки в сохраняемом factory backend: host не масштабирует PCM коэффициентами.

## 4. IPC и время

FD3 закрепляет executable, FD4 — assets directory. Они **не PCM pipes**. PCM: stdin/stdout дочернего процесса; C сохраняет выделенный wire FD и направляет обычный runtime stdout в stderr до DSP initialization.

Одна последовательность по смыслу: descriptor → однократное подтверждение успешной инициализации (READY) → полный input → ровно 640 output bytes. Конкретные bytes, длину и минимальное wire-представление READY совместно определяет Go/C исполнитель Stage02 и закрепляет независимыми fixtures; пять ASCII bytes не являются заранее установленным требованием. Output любого helper: mono S16LE, 16 kHz, 320 samples, 20 ms. Один outstanding exchange. Никаких дополнительных handshakes, per-frame headers, counters, origin, checksum, status или NotReady.

Один общий **5 s** deadline на полный write/read exchange, включая все input buffers и partial syscalls. Это не idle timeout между блоками и не целевая latency.

Один **10 s startup deadline** покрывает подготовку, запуск, descriptor и READY. Startup вне capture loop; Built-in продолжает работать. После READY этот deadline снимается, session cancellation живёт независимо. Нельзя привязать успешно запущенный child к истекающему startup CommandContext. Нет HELLO 500 ms, first-PCM 1 s, preflight или retry.

C output FIFO получает **64 нулевых samples один раз при новой DSP-сессии**, затем каждый 20-ms exchange возвращает 320 samples. Padding не попадает во вход и не повторяется после mute, Rewire или XRUN. Сохраняется 8-ms vendor quantum.

## 5. SoftwareCut и ownership

На capture block получить SoftwareCut один раз в общей границе до native conversion; по тому же значению:
1. заглушить native MIC и REF до helper/conversion, с правильными sample widths/strides;
2. сохранить stock clear(raw), включая stock software REF, который мог быть сформирован conversion.

Hardware mute и аппаратные controls не менять. Допустимы нулевые данные в stock resampler/level meters и отличия переходных кадров. Запрещены второй output privacy check, epochs, masking после DSP, reset FIR/DSP/history и компенсация переходов.

Published mono frame принадлежит подписчикам и неизменен при следующих capture/IPC. Сохранить существующий owned stock frame; External scratch не публиковать повторно используемым. Только необходимая копия, без новых pools/queues/ownership framework. Обязателен retention-test при External, fallback и переключениях.

## 6. Session, lifecycle и concurrency

Session принадлежит долгоживущему аудиокомпоненту echod, а не отдельному Run/Close. Capture Close, Rewire и штатный ALSA overrun сохраняют helper PID/session/pipes/FIFO. Нет повторных descriptor/READY, DSP init, padding или geometry revalidation после reopen.

Завершение: явный Built-in, реальный launch/IPC/process failure, окончательный echod shutdown. Порядок shutdown:
1. запрет новых launch/capture операций;
2. немедленная cancellation launch/exchange и закрытие pipes;
3. ожидание завершения audio processing/Run;
4. освобождение child/FD и ожидание единственного reaper.

У cmd.Wait ровно один владелец; shutdown ожидает его результат, не вызывает Wait повторно. Существующий service.Group ждёт Run до Close: final hook должен срабатывать до этого ожидания, а не в capture Close.

На завершении launch и на публикации PCM проверить identity текущего session object, helper, Selected=External и отсутствие cancellation. Решение принять ответ и фактическая публикация синхронизированы с выбором режима. Старый cleanup не закрывает новую session. Publication mutex не держать на blocking IPC. Generation counters, stale age tracking и новые managers запрещены.

## 7. Selected/Active, fallback и logs

Только Selected и Active. Selected и persisted config обновляются немедленно при выборе, независимо от READY/PCM. До первого полного успешного exchange Active=Built-in; первый успешный ответ может включить External.

Fallback меняет Active на Built-in, сохраняет Selected=External, закрывает только виновную session, пишет причину. Нет auto retry. Ручной Built-in → External заново открывает и проверяет актуальную установку. External option доступна в HA даже без executable на startup.

Ровно три типа **новых ASP host log events**: launch, fallback с причиной, termination. Это не запрет существующих stock logs. Нет H/E, ASP Debug, health/status attributes, feedback/custom log queues/suppression или parser helper stderr.

## 8. Secure launch и C scope

Сохранить secure traversal из /data, owner/mode/type checks, O_NOFOLLOW, FD pinning, защиту check/use и cleanup всех ветвей. Отдельная ручная активация не использует lifetime inode cache. Уменьшать дублирующую проверку можно только при сохранении реальных trust boundaries.

Новый C repository содержит минимальный helper runtime, DSP adapter/bridge, необходимый launch substrate, build и tests. Если vendor backend требует Bionic launcher, перенести только доказанно необходимый FD/mount/exec путь с license/provenance; это не разрешение копировать старый проект целиком. Installer/OTA/download/release automation и proprietary runtime assets не импортировать. Программная проверка mock ABI не доказывает работу vendor DSP на устройстве.

## 9. Контроль отсутствия legacy

Не создавать/не переносить CFG2/TLV, старые envelopes, status/feedback/log pipeline, ASP Debug, preflight/confirmation/delayed persistence, NotReady, platform whitelist, inode cache, production synthetic adapter, async PCM worker/queue, auto retry, новые managers/registries/generic validators, compatibility layers.

Физическое удаление допустимо только если механизм действительно существует в принятой новой базе и доказано, что он относится к ASP. Stock queues, service restart, WebRTC helper, diagnostics и converters не удалять по совпадению имени. Количество удалений может быть нулём; LOC — метрика diff, не квота.

## 10. Git, handoff и принятие

Интеграция Go: `dmit2k/techo5:experimental/external-asp-v3-dev-main` от frozen T01 62603a7854283a3a07473a17c831c6ef0fd6f8d4. C: `dmit2k/frankenasp-v3:experimental/v3-dev-main`. Оба репозитория main никогда не получает V3-код.

Кодовые этапы NN работают в отдельных worktrees:
- TECHO5: `/home/gptlab/dev/techo5-v3-stageNN`, branch `experimental/external-asp-v3-dev-stageNN`;
- C: `/home/gptlab/dev/frankenasp-v3-stageNN`, branch `experimental/v3-stageNN`.

Начало только от последних принятых интеграционных SHA, записанных в `stages/NN_HANDOFF.md`. Не считать PASS тестов или наличие remote branch приёмкой владельца. Handoff содержит отдельно execution PASS/FAIL, acceptance PENDING/ACCEPTED и integration PENDING/COMPLETE. Следующий этап требует ACCEPTED + COMPLETE.

Реальные новые commits создавать только на меняющейся стороне; вторую фиксировать неизменённым SHA. Нет пустых stage-веток/коммитов. Bootstrap и чистая приёмка используют отдельные detached verification worktrees; не имитируют изменение кода.

Stage executor делает локальные commits, возвращает отчёт по [HANDOFF_TEMPLATE.md](HANDOFF_TEMPLATE.md); push только своих stage refs в пределах отдельного поручения. Без прямого commit/merge в integration refs. Владелец принимает результат; отдельное поручение integration agent продвигает refs, по возможности FF без squash/cherry-pick. На обеих сторонах IPC принимается точная **пара T_SHA + C_SHA**, включая неизменённую сторону. Cross-repository Git не атомарен: до проверки обоих remote HEAD и записи пары integration=COMPLETE следующий этап запрещён.

Отчёты/решения о приёмке публикует документационный координатор в `experimental/v3-design-staging`; кодовый исполнитель туда не коммитит. Документальную ветку не merge/rebase в code. При изменении принятой пары повторить затронутые совместные тесты. Несовпадение HEAD, чужой dirty worktree, divergence, отсутствие accepted handoff → STOP без reset/force-push.

## 11. Проверки и STOP

Каждый этап выполняет свои tests; 07 объединяет доказательства и проверяет финальную пару, не заменяет локальные проверки. Software PASS требует builds/conformance/integration/race/security/stock; hardware/vendor runtime на Echo и акустика отдельно: **НЕ ВЫПОЛНЕНЫ**, пока владелец явно не поручит их.

Никаких Echo operations, установки или запуска target binaries автоматически. Offline compiled tests выполняются на anfisa-trainer в тестовых каталогах без воздействия на устройства.

При фактическом противоречии: STOP, точный file/symbol/SHA, нарушенный пункт, воспроизводимый сценарий, минимальное исправление и объяснение ограничения. Не менять APPROVED V2 и не вводить новую архитектуру. Отсутствие toolchain — BLOCKED/FAIL соответствующей проверки, а не фиктивный PASS.
