# План реализации расширяемого ASP2 в TECHO5 (OFFLINE FIRST)

**Статус:** утверждённое архитектурное направление; нормативный протокол описан, исходники НЕ менялись. Дата 2026-10-27.
**Исходная ревизия нашего TECHO5:** `c2cf4c1ed888206245f42bbc1ab50b83ced2b03d`.
**Заводская основа:** официальный TECHO5 v1.6.9 `c0665429eed5cb563589ddbcaeeba3a474841f81`.
**Норма:** [ASP2_PROTOCOL.md](ASP2_PROTOCOL.md). Парный документ находится в отдельном FrankenASP.

## Что относится именно к TECHO5

| Исходный файл/модуль | Обязательное изменение |
|---|---|
| `echod/internal/asp/protocol.go` | Заменить фиксированный `ConfigSize=40` расширяемым CFG2 v2: 16-байтный заголовок, TLV registry, bounded encode/decode, strict mandatory/optional, immutable Config, canonical serialization. Не менять INP2/RSP2 без доказанной необходимости. |
| `echod/internal/asp/protocol_test.go` | Golden bytes, negative/fuzz, unknown optional/required, legacy CFG2 reject, malformed length, duplicate/order, max-size tests. |
| `echod/internal/asp/synthetic/adapter.go` | Синтетический generic helper принимает/игнорирует корректные optional extensions; сохраняет raw MIC; тестирует новые ACK/REJECT. |
| `echod/internal/hardware/mic/asp.go` | `nativeFrame` и `Input` несут неизменяемый session descriptor, создаваемый один раз до начала обменов. Не читать IDME в 20-ms capture loop. Конфигурация отправляется один раз, проверяется на следующем exchange без блокировок сверх 25 ms. |
| Новый `echod/internal/hardware/mic/asp_platform.go` (пример имени) | Формировать необязательный `PLATFORM_INFO` TLV `0x0003` из `layout.Board`, фиксировать CONFIRMED/INFERRED, учитывать `layout.BoardDoubt` и надёжность идентификации конкретной платформы. Не отдавать raw BoardDoubt, MAC/serial; helper не должен воспринимать codename как доверенный источник PCM shape. |
| Новый `echod/internal/hardware/mic/asp_calibration_athena.go` (пример имени) | Заимствовать только штатный `layout.Idme`, прочитать четыре реальных `miccal.N` на старте, провалидировать и нормализовать factory-legacy Q30, привязать к captured slots. Никаких коэффициентов в PCM. При отсутствии/ошибках выдавать NONE/PARTIAL, не выдавать фальшивую unity. |
| `echod/internal/hardware/mic/asp_path.go` | Путь `/data/techo5-linux/external-asp/bin/external-asp`, новая безопасная политика обнаружения и запуска за закреплённый файловый дескриптор, детально ниже. |
| `echod/internal/hardware/mic/asp_control_test.go`, `asp_test.go`, `asp2_test.go`, `asp_handoff_test.go`, `asp2_timing_test.go` | Обновить контракт и regression fixtures, включая полный stock fallback. |
| Настройки/телеметрия `asp_control.go`, `asp_feedback.go`, `asp_log.go` | Не передавать сами коэффициенты в HA, логи или diagnostics; только статус factory gains available/partial/none, backend/authority. Не менять текущее устройство логирования и ASP timeout semantics. |

### Порядок работ внутри TECHO5

1. Перед работой создать новую ветку/worktree от pinned HEAD; никакого правления canonical upstream, релизного stock, постоянной установленной ARM32 копии.
2. Реализовать wire codec и независимые golden test vectors, включая `PLATFORM_INFO`: подтверждённая `athena`, guessed/fallback, неизвестный, отсутствие TLV, некорректная кодировка. Первые 40 байт не сохранять ради видимой «совместимости»: протокол формально новый pre-release CFG2 layout, версия остаётся ASP2=2.
3. Вынести сбор метаданных в **холодный путь сессии** с immutable snapshot: MIC/REF format и flags из источника, заводские gains из `layout.Idme`. Убедиться, что штатный `layout.Idme` уже сначала читает device tree, а boot1 — fallback; не клонировать парсер IDME.
4. Разделить Parse/Normalize/Availability. При четырёх известных каналах пробрасывать четыре signed-Q30 значения в исходном capture-slot порядке; при неверном/отсутствующем канале корректно формировать PARTIAL/NONE; не повышать общий gain.
5. Обновить synthetic DSP и совместимость с helper, сохраняющим абсолютно raw MIC байты. Алгоритм ASP2 handoff, FFT/history/listener и встроенная fallback-обработка остаются прежними.
6. Изменить путь и политику безопасного запуска (см. раздел ниже) и изолированно прогнать race/security tests.
7. Провести `go test` для затронутых пакетов и полный `go test ./...` в совместимом окружении, затем cross-build ARM32 с теми же параметрами, что и проверенный `c2cf4c1`. Go SDK может отсутствовать в обычном PATH рабочего окружения — при реализации явно указать toolchain, не считать тесты пройденными до запуска.

## Конкретный путь и политика доверия

**НЕ изменять права/владельцев:** `/data` (устройство mainstream; 0777) и `/data/techo5-linux` (официальный TECHO5; 0777).
Сейчас `/data/techo5-linux/external-asp` не существует, это НАШ каталог.

Плановая структура с явным владельцем root:

~~~text
/data/techo5-linux/external-asp/    root:root 0700
├── bin/                            root:root 0700
│   └── external-asp                 root:root 0755  native Linux ELF launcher
├── etc/                            root:root 0700
│   ├── profile.conf                root:root 0600
│   └── assets.sha256               root:root 0600
└── lib/                            root:root 0700
    └── rootfs/                     root:root 0700
        ├── system/                 original owner-local Bionic runtime
        ├── vendor/                 RO bind real /vendor (in PRIVATE mount ns)
        └── opt/external-asp/       original C/Bionic factory helper
~~~

Не распространять эту Bionic-структуру как требование для generic External ASP: он может быть обычным Linux ELF и вообще не использовать chroot.

### Критическая проверка безопасности

Сейчас `findASP()` выполняет `EvalSymlinks`, `Stat` и затем `exec.Command(path)`, т.е. проверяет путь, а запускает **путь заново**. Простое исключение двух родительских `0777` из ACL оставит TOCTOU между проверкой и exec; так делать нельзя.

Для нашего точного preferred path реализовать на Linux:

1. От доверенного `/` последовательно открывать каждый компонент через `openat(O_DIRECTORY|O_NOFOLLOW|O_CLOEXEC)`. Для уже существующих `/data` и `/data/techo5-linux` проверить root ownership, но разрешить их **фактическую штатную** group/world-write только по этим двум точным уровням, не обобщать exception на произвольные пути.
2. Наш `external-asp/`, `bin/` и terminal file проверять через `fstat` уже открытых FD: root-owned, без group/other write, корректные типы/исполняемость, без symlink и специальных устройств.
3. Сохранить FD открытого исполняемого файла **до фактического exec**, не перерезолвить pathname после `fstat`. На Linux провести прототип `exec.Cmd.ExtraFiles` + `/proc/self/fd/3` (либо проверенный `execveat(AT_EMPTY_PATH)`). Передать launcher также pinned FD каталога `external-asp` (например fd 4) для `openat` его профиля/rootfs: подмена имени предков не должна менять запускаемые assets.
4. Проверить поведение для статического ELF launcher, интерпретатора, отсутствующего `/proc/self/fd`, неисправных дескрипторов и архитектуры ARM32. Сохранять штатный fallback при любом отказе. Не утверждать успешность механизма до реальных race/security тестов на Linux.
5. Для PATH/fallback кандидатов сохранить строгую старую политику или убрать сомнительные fallback совсем после оценки совместимости. Исключения для `0777` НЕ должны расширяться на произвольные места.
6. Новый рабочий helper обнаруживается один раз на lifetime daemon; установка после запуска требует явного restart через штатную процедуру, не скрытой подмены executable.
7. Installer создаёт только `external-asp/*`, запрещает symlink-подмены и проверяет атомарный secure staging/activation. Он не трогает parent chmod/chown и не монтирует ничего глобально.

Без FDs и проверенного exec в этом окружении — **NO GO** для production activation. Запасной вариант — держать trusted launcher в действительно защищённой mainstream-controlled директории, но пользователь установил требование единого каталога; предпочтительна FD-based схема.

## Нагрузочные и отказоустойчивые тесты

- Byte-exact raw MIC и REF сквозь CFG+INP, никакой host-Q30;
- Factory IDME complete/partial/none; documented legacy scaling, valid-mask, ownership of channel maps; `PLATFORM_INFO` confidence, board-independent stream validation, unknown/absent codename;
- Mandatory vs optional TLV, duplicates, invalid/replayed config, unknown encoding, max 8192;
- 500 ms HELLO, 25 ms per-exchange, existing 3 DATA preflight, stock fallback, closed children/FD, restart with new metadata snapshot;
- File substitution race: rename `external-asp` under 0777 parent, symlink attacks at every component, non-root owner, writable launcher, script/ELF interpreter failure, atomic installer replacement;
- Native launcher mount namespace does not contaminate mainline mount namespace, factory Bionic failure still falls back, no change to existing /data mount or A/B slots.

**Acceptance:** host codec + helpers golden conformance PASS, full Go tests PASS, ARM32 cross-build PASS, secure FD-exec proof PASS, no regressions in current stock / MWW / playback fallback. Device test only under fresh DEVICE GO, not in the documentation task.

## Delivery checkpoints

- T1: normative wire spec + Go codec/golden vectors.
- T2: source-owned factory calibration descriptor + generic synthetic helper.
- T3: secure path + pinned-FD launch under `external-asp/`, Linux security/race tests.
- T4: paired FrankenASP implementation/ARM32 Bionic build and offline real DSP replay.
- T5: offline integration sign-off: signed artifacts/hashes, install/rollback runbook, DEVICE GO only after separate review.

**Publication:** GitHub push forbidden until explicit approval. Canonical A/B stock is left untouched.
