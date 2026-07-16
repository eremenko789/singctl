# Feature Specification: Task Checklist

**Feature Branch**: `009-task-checklist`

**Created**: 2026-07-16

**Status**: Draft

**Input**: User description: "F09 task checklist — CRUD /v2/checklist-item (parent = task). Depends F08. Out of scope: TUI checklist UI. Acceptance: 5 CLI-команд checklist; operations ChecklistItemController_* закрыты. Inputs: ТЗ §6.1 (чек-листы), coverage ChecklistItemController_*, OpenAPI, constitution."

## Clarifications

### Session 2026-07-16

- Q: Фильтры `task checklist list` сверх parent? → A: Только `<TASK_ID>` (parent); без `--limit` / `--offset` / `--removed`.
- Q: Флаг `--order` (`parentOrder`) на add/update? → A: Без `--order`; `parentOrder` с CLI не задаём.
- Q: Несуществующий parent (TASK_ID) на list/add? → A: Перед list/add проверка задачи (как `task get`); not found → exit `3` без вызова checklist API.
- Q: Пустой / whitespace-only `--title`? → A: Exit `1` до сети на add и update.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List and inspect checklist items (Priority: P1)

Пользователь или скрипт получает список пунктов чек-листа задачи и может посмотреть один пункт по идентификатору. Вывод идёт в выбранном формате (table/json/yaml/csv) с разделением stdout/stderr и кодами выхода из F07.

**Why this priority**: Без list/get нельзя автоматизировать работу с подзадачами чек-листа и закрыть read-путь `ChecklistItemController_list` / `ChecklistItemController_getById`.

**Independent Test**: С мок-API: `task checklist list <TASK_ID>` сначала проверяет существование задачи, затем передаёт parent = ID задачи и рендерит результат; `task checklist get <ID>` возвращает один пункт; при «не найдено» задачи или пункта — exit `3` и сообщение в stderr.

**Acceptance Scenarios**:

1. **Given** существующая задача и API возвращает набор пунктов, **When** пользователь выполняет `singctl task checklist list <TASK_ID>`, **Then** в stdout — список в текущем формате вывода (F06), exit `0`, stderr на обычном пути пуст; запрос к checklist API фильтрует по parent = TASK_ID; ответ проверки задачи в stdout не попадает.
2. **Given** известный ID существующего пункта, **When** выполняется `singctl task checklist get <CHECKLIST_ITEM_ID>` с `-o json` (или `yaml`), **Then** корень stdout — один объект пункта (не массив); при `-o table`/`csv` — одна запись.
3. **Given** ID несуществующего пункта (API «не найдено»), **When** выполняется `checklist get` / `update` / `delete`, **Then** exit `3`, сообщение в stderr, stdout пуст.
4. **Given** существующая задача и API возвращает пустой список пунктов, **When** выполняется `task checklist list <TASK_ID> -o json`, **Then** stdout — пустой массив `[]` (и валидные пустые представления для других форматов), exit `0`.
5. **Given** существующая задача и API возвращает набор пунктов, **When** выполняется `task checklist list <TASK_ID> -o json`, **Then** корень — массив объектов пунктов.
6. **Given** TASK_ID несуществующей задачи, **When** выполняется `task checklist list <TASK_ID>`, **Then** exit `3`, сообщение в stderr, stdout пуст; checklist list API не вызывается.

---

### User Story 2 - Add and update checklist items (Priority: P1)

Пользователь добавляет пункт чек-листа к задаче с обязательным заголовком и обновляет существующий пункт (заголовок и/или статус выполнения).

**Why this priority**: Вместе с list/get закрывает write-путь (`ChecklistItemController_create` / `ChecklistItemController_update`) и acceptance «все operations ChecklistItemController_*».

**Independent Test**: Мок: `task checklist add <TASK_ID> --title "…"` при существующей задаче создаёт пункт и показывает результат; несуществующая задача — exit `3` без create; `task checklist update <ID> --done` патчит статус; без `--title` на add — ошибка использования (exit `1`) без сети; update без изменяющих флагов — exit `1` без сети.

**Acceptance Scenarios**:

1. **Given** существующая задача, валидный токен и API принимает создание, **When** пользователь выполняет `singctl task checklist add <TASK_ID> --title "Пункт"` (опционально с `--done`), **Then** пункт создаётся с parent = TASK_ID, в stdout — полный пункт (тот же логический набор полей, что у `checklist get`), exit `0`; ответ проверки задачи в stdout не попадает.
2. **Given** команда `task checklist add` без `--title` или без TASK_ID, **When** она запускается в неинтерактивном режиме, **Then** ошибка использования (exit `1`) без вызова API; интерактивный prompt — вне scope (F26).
3. **Given** существующий пункт, **When** выполняется `singctl task checklist update <CHECKLIST_ITEM_ID>` с одним или несколькими флагами `--title`, `--done`, `--undone`, **Then** на API уходит частичное обновление только указанных полей, в stdout — полный обновлённый пункт (как `get`), exit `0`.
4. **Given** `task checklist update <ID>` без каких-либо изменяющих флагов, **When** команда выполняется, **Then** ошибка использования (exit `1`) без сетевого вызова.
5. **Given** одновременно `--done` и `--undone`, **When** update, **Then** ошибка использования (exit `1`) до вызова API.
6. **Given** TASK_ID несуществующей задачи, **When** выполняется `task checklist add <TASK_ID> --title "Пункт"`, **Then** exit `3`, сообщение в stderr, stdout пуст; checklist create API не вызывается.
7. **Given** `--title` пустой или только пробелы, **When** `add` или `update`, **Then** exit `1` до сети (и до pre-check задачи на add).

---

### User Story 3 - Delete checklist items (Priority: P1)

Пользователь безвозвратно удаляет пункт чек-листа по идентификатору.

**Why this priority**: Закрывает `ChecklistItemController_delete` из coverage и полный CRUD scope F09.

**Independent Test**: Мок: `task checklist delete <ID>` вызывает удаление; успех — exit `0`, stdout пуст; not found — exit `3`.

**Acceptance Scenarios**:

1. **Given** существующий пункт, **When** выполняется `singctl task checklist delete <CHECKLIST_ITEM_ID>`, **Then** пункт удаляется, stdout пуст, exit `0`; без интерактивного подтверждения (scriptability; интерактивные формы — F26).
2. **Given** пустой или whitespace-only ID, **When** delete/get/update, **Then** ошибка использования (exit `1`) до сети.

---

### User Story 4 - Discoverable CLI help for checklist commands (Priority: P2)

Пользователь через `--help` узнаёт подгруппу `task checklist` и пять команд list/get/add/update/delete, их аргументы и флаги, согласованные с ТЗ §6.1.

**Why this priority**: Acceptance F09 («5 CLI-команд checklist»); discoverability и единообразие с `task` из F08.

**Independent Test**: Вызвать `singctl task checklist --help` и `--help` каждой подкоманды; убедиться, что все пять команд и ключевые флаги описаны.

**Acceptance Scenarios**:

1. **Given** установленный CLI, **When** пользователь выполняет `singctl task --help`, **Then** видна подгруппа/команда `checklist` (или эквивалентная навигация к ней).
2. **Given** пользователь выполняет `singctl task checklist --help`, **Then** видны подкоманды `list`, `get`, `add`, `update`, `delete` с кратким назначением каждой.
3. **Given** пользователь смотрит `--help` каждой подкоманды, **When** читает флаги, **Then** описаны аргументы/поля scope F09 (ТЗ §6.1); нет обещаний TUI чек-листа или kanban/`move`.
4. **Given** неизвестная подкоманда или флаг, **When** пользователь ошибается в вызове, **Then** exit `1`, сообщение в stderr (контракт F07).

---

### User Story 5 - Adapter coverage with unit tests (Priority: P1)

Слой поверх API обеспечивает вызовы всех операций чек-листа (`list`, `create`, `getById`, `update`, `delete`) и покрывается unit-тестами с мок-HTTP (happy path и ключевые ошибки), без живого API в обязательном DoD.

**Why this priority**: Acceptance F09 («operations ChecklistItemController_* закрыты») и constitution III/IV/IX: CLI не должен содержать ручной HTTP; адаптер тестируем независимо от cobra-рендера.

**Independent Test**: Для каждой из пяти operations ChecklistItemController_* — unit-тест адаптера с мок-HTTP (успех); плюс хотя бы not-found или не-2xx на get/update/delete.

**Acceptance Scenarios**:

1. **Given** мок отвечает успехом на list/create/get/update/delete, **When** адаптерный фасад чек-листа выполняет соответствующий вызов, **Then** результат смапплен в доменное/типизированное представление, пригодное для CLI.
2. **Given** мок возвращает «не найдено» на get/update/delete, **When** вызывается адаптер, **Then** ошибка классифицируется так, что CLI может завершиться кодом `3` (F05/F07).
3. **Given** набор тестов адаптера чек-листа, **When** они запускаются без сети к production API, **Then** все обязательные сценарии US5 проходят на моках; фикстуры токенов — только фиктивные (constitution VII).

---

### Edge Cases

- Нет токена / ошибка конфигурации: exit `2`, без вызова API (F02/F07).
- Ошибка API/транспорт (не not-found): exit `1`, сообщение в stderr, stdout пуст (F05/F07).
- `list` / `add` без TASK_ID: ошибка использования (exit `1`) до сети.
- Перед `list` / `add`: проверка существования задачи тем же механизмом, что `task get` (F08); not found → exit `3`, stdout пуст, checklist API не вызывается; ответ проверки не пишется в stdout. Прочие ошибки проверки (транспорт и т.п.) — по F05/F07 (обычно exit `1`).
- `get` / `update` / `delete` пункта: pre-check родительской задачи не требуется (достаточно ID пункта).
- Родитель чек-листа в F09 — всегда задача (TASK_ID); смена parent через update — вне F09.
- `delete` не требует `--force` и не спрашивает подтверждение в F09; при успехе stdout пуст.
- Успешные add/update пишут в stdout полный пункт (тот же логический набор полей, что `get`).
- Пустой или whitespace-only ID/TASK_ID: ошибка использования (exit `1`) до сети.
- Пустой или whitespace-only `--title` на `add` / `update`: exit `1` до сети (на `add` — до pre-check задачи).
- TUI-экран чек-листа и хоткей `c` — вне F09 (отдельные TUI-фичи); help CLI не обещает TUI.
- Поля OpenAPI сверх ТЗ (`crypted`, `parentOrder`) — вне CLI-surface F09; `--order` нет.
- `task checklist list` MUST NOT принимать `--limit` / `--offset` / `--removed` (и аналоги); в API уходит только фильтр parent = TASK_ID (пагинация/`includeRemoved` OpenAPI — вне F09).
- Смена parent пункта через `update` — вне F09.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Система MUST предоставлять CLI-подгруппу `task checklist` с подкомандами `list`, `get`, `add`, `update`, `delete`.
- **FR-002**: `task checklist list <TASK_ID>` MUST сначала проверить существование задачи (как `task get`); при not found MUST завершаться exit `3` без вызова checklist list; при успехе MUST запрашивать список пунктов с фильтром parent = TASK_ID и выводить результат в формате F06; MUST NOT принимать флаги пагинации или `--removed` / `includeRemoved`; stdout MUST содержать только результат списка пунктов (не тело проверки задачи).
- **FR-003**: `task checklist get <CHECKLIST_ITEM_ID>` MUST возвращать один пункт по идентификатору.
- **FR-004**: `task checklist add <TASK_ID>` MUST требовать `--title` (непустой после trim пробелов; иначе exit `1` до сети и до pre-check); MUST сначала проверить существование задачи (как `task get`); при not found MUST завершаться exit `3` без create; при успехе MUST создавать пункт с parent = TASK_ID; MAY принимать `--done` (по умолчанию не выполнен); MUST NOT принимать `--order` / `parentOrder`; stdout при успехе MUST содержать только созданный пункт (не тело проверки задачи).
- **FR-005**: `task checklist update <CHECKLIST_ITEM_ID>` MUST принимать опциональные `--title`, `--done`, `--undone` (взаимоисключающие `--done`/`--undone`); если `--title` передан, он MUST быть непустым после trim (иначе exit `1` до сети); MUST выполнять частичное обновление только переданных полей; при отсутствии изменяющих флагов MUST завершаться ошибкой использования (exit `1`) без вызова API; MUST NOT принимать `--order` / `parentOrder` / смену parent; MUST NOT требовать pre-check родительской задачи.
- **FR-006**: `task checklist delete <CHECKLIST_ITEM_ID>` MUST безвозвратно удалять пункт; при успехе stdout MUST быть пустым.
- **FR-007**: Успешные `add` / `update` MUST писать в stdout полный пункт с тем же логическим набором полей, что успешный `get` (формат — F06).
- **FR-008**: В `json`/`yaml`: `list` MUST давать корень-массив объектов; успешный `get` / `add` / `update` MUST давать один объект пункта (не массив из одного элемента). Для `table`/`csv`: list — много строк; один пункт — одна запись/строка данных.
- **FR-009**: Все пять operations ChecklistItemController_* (`list`, `create`, `getById`, `update`, `delete`) MUST быть достижимы через адаптерный слой и CLI F09.
- **FR-010**: Команды checklist MUST соблюдать контракты F06/F07: форматы вывода, отсутствие ANSI в pipe, stdout vs stderr, exit codes `0`/`1`/`2`/`3`.
- **FR-011**: `--help` для `task checklist` и каждой подкоманды MUST документировать назначение, аргументы и флаги в scope F09.
- **FR-012**: Адаптер чек-листа MUST быть покрыт unit-тестами с мок-HTTP для каждой ChecklistItemController-операции (минимум happy path) без обязательного live API в DoD.
- **FR-013**: Система MUST NOT реализовывать в F09 TUI checklist UI, kanban-link, `task move` или интерактивные prompt’ы (TUI / F10 / F26).
- **FR-014**: Вызовы API чек-листа MUST идти через существующий адаптер/codegen-клиент, без ручных HTTP CRUD/DTO (constitution III).
- **FR-015**: При отсутствии токена или иной ошибке конфигурации команды checklist MUST завершаться кодом `2` до сетевого вызова (где применимо).

### Key Entities

- **Checklist item**: Пункт чек-листа задачи — идентификатор, заголовок, статус выполнения (`done`), родитель (ID задачи), порядок в списке (если присутствует в ответе API).
- **Checklist list query**: Выборка пунктов по parent = ID задачи.
- **Checklist write input**: Поля создания/обновления из CLI (title, done); parent задаётся позиционным TASK_ID на add/list.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Пользователь может выполнить полный цикл «добавить → получить → изменить (в т.ч. отметить выполненным) → удалить» по документации/`--help` с первой попытки без обращения к исходному коду.
- **SC-002**: 100% operations ChecklistItemController_* из матрицы покрытия достижимы через поставку F09 (CLI + адаптер); TUI checklist при этом отсутствует.
- **SC-003**: Для каждой ChecklistItemController-операции есть автоматический unit-тест адаптера с мок-HTTP (happy path), проходящий без доступа к production API.
- **SC-004**: `singctl task checklist --help` и help каждой из пяти подкоманд перечисляют команды/флаги scope F09; ревьюер за ≤5 минут подтверждает соответствие ТЗ §6.1 (чек-листы).
- **SC-005**: Типичные ошибки (нет токена, not found задачи/пункта, misuse флага, сбой API) дают коды `2` / `3` / `1` / `1` соответственно и не смешивают текст ошибки с data-stdout (контракт F07); для `list`/`add` not found задачи — exit `3` без вызова checklist API.
- **SC-006**: Скрипт может выполнить `task checklist list <TASK_ID> --output json` и получить массив записей без ANSI при redirect; `task checklist get -o json` — один объект пункта (не массив), exit `0` при успехе.

## Assumptions

- Зависимости F08 (и транзитивно F01–F07) доступны: группа `task` (включая `task get` для pre-check), конфиг/токен, codegen, адаптер, ошибки/retry, рендер, scriptability.
- CLI-surface совпадает с ТЗ §6.1: пять команд `list` / `get` / `add` / `update` / `delete`; parent всегда задача.
- Pre-check parent на `list`/`add` использует тот же путь, что получение задачи по ID (F08); лишний round-trip осознан ради явного exit `3`.
- `--done` / `--undone` — явное включение/снятие статуса; на `add` достаточно опционального `--done` (без `--undone`).
- Пустой / whitespace-only `--title` (после trim) → exit `1` до сети на add и update; непустое значение передаётся в API как указано пользователем.
- Поля OpenAPI `crypted`, `parentOrder`, пагинация/`includeRemoved` list — вне F09 (clarified: list только parent; без `--order`); порядок пунктов при создании оставляет API.
- Интерактивный ввод и confirm на delete отложены на F26; в F09 — только флаги/аргументы.
- TUI checklist (правая панель / хоткей `c`) — вне F09; alias `singctl t` — F25.
- Интеграционные/live API-тесты — F33; в DoD F09 достаточно мок-unit-тестов адаптера + CLI-проверка help/поведения на моках или эквивалентном harness.
- Формат представления пункта и колонок списка опирается на общий рендерер F06; разумный минимум колонок table (id, title, done, parent) может уточняться в plan без расширения scope API.
