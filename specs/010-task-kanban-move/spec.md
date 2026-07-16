# Feature Specification: Task Kanban Link & Move

**Feature Branch**: `010-task-kanban-move`

**Created**: 2026-07-16

**Status**: Draft

**Input**: User description: "F10 task kanban link & move — CRUD /v2/kanban-task-status + UX task move. Depends F08 (рекомендуется F13). Out of scope: управление колонками (F13). Acceptance: полный CRUD + task move как create/update. Inputs: ТЗ §6.1 (канбан-связь), coverage KanbanTaskStatusController_*, OpenAPI, constitution."

## Clarifications

### Session 2026-07-17

- Q: `kanban create`, если у задачи уже есть активная связь? → A: Всегда create в API; дубликаты на клиенте не блокируем.
- Q: Флаг `--order` на `task move`? → A: Без `--order` на `move`; порядок при create через move — default API, при update `kanbanOrder` не меняем.
- Q: `task move`, если задача уже в этой же колонке? → A: При 1 связи всегда update `statusId` (даже если колонка уже та же).
- Q: Pre-check задачи на `task kanban list --task`? → A: Без pre-check на list; всегда list API (с фильтрами или без).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List and inspect kanban links (Priority: P1)

Пользователь или скрипт получает список связей задача↔канбан-колонка (опционально фильтруя по задаче и/или колонке) и может посмотреть одну связь по идентификатору. Вывод идёт в выбранном формате (table/json/yaml/csv) с разделением stdout/stderr и кодами выхода из F07.

**Why this priority**: Без list/get нельзя автоматизировать обзор размещения задач на доске и закрыть read-путь `KanbanTaskStatusController_list` / `KanbanTaskStatusController_getById`.

**Independent Test**: С мок-API: `task kanban list` рендерит связи; с `--task` / `--status` фильтры уходят в запрос; `task kanban get <LINK_ID>` возвращает одну связь; при «не найдено» — exit `3` и сообщение в stderr.

**Acceptance Scenarios**:

1. **Given** API возвращает набор связей, **When** пользователь выполняет `singctl task kanban list`, **Then** в stdout — список в текущем формате вывода (F06), exit `0`, stderr на обычном пути пуст.
2. **Given** известные фильтры, **When** выполняется `task kanban list --task <TASK_ID>` и/или `--status <COLUMN_ID>`, **Then** запрос к API отражает соответствующие фильтры `taskId` / `statusId`; stdout содержит только результат этого списка.
3. **Given** известный ID существующей связи, **When** выполняется `singctl task kanban get <LINK_ID>` с `-o json` (или `yaml`), **Then** корень stdout — один объект связи (не массив); при `-o table`/`csv` — одна запись.
4. **Given** ID несуществующей связи (API «не найдено»), **When** выполняется `kanban get` / `update` / `delete`, **Then** exit `3`, сообщение в stderr, stdout пуст.
5. **Given** API возвращает пустой список, **When** выполняется `task kanban list -o json`, **Then** stdout — пустой массив `[]` (и валидные пустые представления для других форматов), exit `0`.
6. **Given** API возвращает набор связей, **When** выполняется `task kanban list -o json`, **Then** корень — массив объектов связей.
7. **Given** несуществующий `--task` (или `--status`), **When** выполняется `task kanban list` с этим фильтром, **Then** CLI не делает pre-check `task get` / column get; уходит list API; типичный пустой результат → stdout пустой список, exit `0` (если API не вернул ошибку).

---

### User Story 2 - Create and update kanban links (Priority: P1)

Пользователь явно создаёт связь задачи с колонкой и обновляет существующую связь (колонка, задача и/или порядок в колонке).

**Why this priority**: Вместе с list/get закрывает write-путь (`KanbanTaskStatusController_create` / `KanbanTaskStatusController_update`) и acceptance «полный CRUD».

**Independent Test**: Мок: `task kanban create --task … --column …` создаёт связь и показывает результат; без обязательных флагов — exit `1` без сети; `task kanban update <LINK_ID> --column …` патчит поля; update без изменяющих флагов — exit `1` без сети.

**Acceptance Scenarios**:

1. **Given** валидный токен и API принимает создание, **When** пользователь выполняет `singctl task kanban create --task <TASK_ID> --column <COLUMN_ID>` (опционально с `--order N`), **Then** связь создаётся, в stdout — полная связь (тот же логический набор полей, что у `kanban get`), exit `0`.
2. **Given** `task kanban create` без `--task` или без `--column`, **When** команда запускается в неинтерактивном режиме, **Then** ошибка использования (exit `1`) без вызова API; интерактивный prompt — вне scope (F26).
3. **Given** существующая связь, **When** выполняется `singctl task kanban update <LINK_ID>` с одним или несколькими флагами `--task`, `--column`, `--order`, **Then** на API уходит частичное обновление только указанных полей, в stdout — полная обновлённая связь (как `get`), exit `0`.
4. **Given** `task kanban update <ID>` без каких-либо изменяющих флагов, **When** команда выполняется, **Then** ошибка использования (exit `1`) без сетевого вызова.
5. **Given** перед `create` проверка существования задачи (как `task get`) даёт not found, **When** выполняется `kanban create --task <TASK_ID> --column …`, **Then** exit `3`, stdout пуст; create API не вызывается; ответ проверки задачи в stdout не попадает.
6. **Given** у задачи уже есть одна или несколько активных связей, **When** выполняется `task kanban create --task … --column …`, **Then** CLI всё равно вызывает create (без клиентской проверки уникальности); успех/ошибка — по ответу API; при появлении >1 связи последующий `move` завершится exit `1` (US3).

---

### User Story 3 - Move task between columns (Priority: P1)

Пользователь перемещает задачу в канбан-колонку одной командой `task move`: если связи ещё нет — создаётся; если есть ровно одна активная — обновляется колонка; явный CRUD связи при этом остаётся доступен через `task kanban *`.

**Why this priority**: Acceptance F10 («task move как create/update») и основной UX из ТЗ §6.1 / wiki («перемещение = create/update связи»).

**Independent Test**: Мок: задача без связи → `task move <ID> --column <COLUMN_ID>` вызывает create; задача с одной связью → update той же связи; без `--column` — exit `1`; несколько активных связей → ошибка использования/конфликта без «тихого» выбора.

**Acceptance Scenarios**:

1. **Given** существующая задача без активной связи task↔column, **When** выполняется `singctl task move <TASK_ID> --column <COLUMN_ID>`, **Then** создаётся новая связь (create), в stdout — полная связь (как `kanban get`/`create`), exit `0`.
2. **Given** существующая задача с ровно одной активной связью, **When** выполняется `task move <TASK_ID> --column <COLUMN_ID>`, **Then** обновляется `statusId` этой связи (update), в stdout — полная обновлённая связь, exit `0`; в том числе если текущий `statusId` уже равен `--column` — всё равно выполняется update (не short-circuit).
3. **Given** `task move` без `--column` или без TASK_ID, **When** команда запускается, **Then** ошибка использования (exit `1`) без сети.
4. **Given** TASK_ID несуществующей задачи, **When** выполняется `task move`, **Then** exit `3` (после pre-check как `task get`), stdout пуст; list/create/update связи не выполняются (или не доходят до create/update после not found задачи).
5. **Given** у задачи больше одной активной связи, **When** выполняется `task move`, **Then** команда завершается с ошибкой (exit `1`), сообщение в stderr объясняет неоднозначность и указывает на `task kanban list` / `update`; create/update через move не выполняются.
6. **Given** успешный `move`, **When** пользователь смотрит stdout, **Then** ответ промежуточного list (поиск связи) в stdout не попадает — только итоговая связь после create или update.

---

### User Story 4 - Delete kanban links (Priority: P1)

Пользователь безвозвратно удаляет связь задачи с колонкой по идентификатору связи.

**Why this priority**: Закрывает `KanbanTaskStatusController_delete` из coverage и полный CRUD scope F10.

**Independent Test**: Мок: `task kanban delete <LINK_ID>` вызывает удаление; успех — exit `0`, stdout пуст; not found — exit `3`.

**Acceptance Scenarios**:

1. **Given** существующая связь, **When** выполняется `singctl task kanban delete <LINK_ID>`, **Then** связь удаляется, stdout пуст, exit `0`; без интерактивного подтверждения (scriptability; интерактивные формы — F26).
2. **Given** пустой или whitespace-only LINK_ID, **When** delete/get/update, **Then** ошибка использования (exit `1`) до сети.

---

### User Story 5 - Discoverable CLI help for kanban and move (Priority: P2)

Пользователь через `--help` узнаёт подгруппу `task kanban` (list/get/create/update/delete), команду `task move` и их аргументы/флаги, согласованные с ТЗ §6.1.

**Why this priority**: Acceptance F10 (полный CRUD + move); discoverability и единообразие с `task` из F08.

**Independent Test**: Вызвать `singctl task --help`, `singctl task kanban --help`, `--help` каждой подкоманды и `singctl task move --help`; убедиться, что команды и ключевые флаги описаны.

**Acceptance Scenarios**:

1. **Given** установленный CLI, **When** пользователь выполняет `singctl task --help`, **Then** видны `kanban` и `move` (или эквивалентная навигация к ним).
2. **Given** пользователь выполняет `singctl task kanban --help`, **Then** видны подкоманды `list`, `get`, `create`, `update`, `delete` с кратким назначением каждой.
3. **Given** пользователь смотрит `--help` каждой подкоманды и `task move --help`, **When** читает флаги, **Then** описаны аргументы/поля scope F10 (ТЗ §6.1); у `move` — `--column`, без `--order`; нет обещаний управления колонками проекта (F13) или TUI-диалога перемещения.
4. **Given** неизвестная подкоманда или флаг, **When** пользователь ошибается в вызове, **Then** exit `1`, сообщение в stderr (контракт F07).

---

### User Story 6 - Adapter coverage with unit tests (Priority: P1)

Слой поверх API обеспечивает вызовы всех операций канбан-связи (`list`, `create`, `getById`, `update`, `delete`) и покрывается unit-тестами с мок-HTTP (happy path и ключевые ошибки), без живого API в обязательном DoD. Логика `move` (выбор create vs update) тестируется на уровне адаптера или CLI с моками.

**Why this priority**: Acceptance F10 («operations KanbanTaskStatusController_*» через CRUD + move) и constitution III/IV/IX: CLI не должен содержать ручной HTTP; адаптер тестируем независимо от cobra-рендера.

**Independent Test**: Для каждой из пяти operations KanbanTaskStatusController_* — unit-тест адаптера с мок-HTTP (успех); плюс хотя бы not-found или не-2xx на get/update/delete; плюс сценарии move: 0 связей → create, 1 связь → update.

**Acceptance Scenarios**:

1. **Given** мок отвечает успехом на list/create/get/update/delete, **When** адаптерный фасад kanban-task-status выполняет соответствующий вызов, **Then** результат смапплен в доменное/типизированное представление, пригодное для CLI.
2. **Given** мок возвращает «не найдено» на get/update/delete, **When** вызывается адаптер, **Then** ошибка классифицируется так, что CLI может завершиться кодом `3` (F05/F07).
3. **Given** набор тестов адаптера / move-фасада, **When** они запускаются без сети к production API, **Then** все обязательные сценарии US6 проходят на моках; фикстуры токенов — только фиктивные (constitution VII).

---

### Edge Cases

- Нет токена / ошибка конфигурации: exit `2`, без вызова API (F02/F07).
- Ошибка API/транспорт (не not-found): exit `1`, сообщение в stderr, stdout пуст (F05/F07).
- `kanban create` / `move`: pre-check существования задачи тем же механизмом, что `task get` (F08); not found → exit `3`, stdout пуст; ответ проверки не пишется в stdout. Прочие ошибки проверки — по F05/F07 (обычно exit `1`).
- `task kanban list`: без pre-check задачи/колонки (clarified); несуществующий `--task` может дать пустой список и exit `0`, если API так ответил.
- Проверка существования колонки (kanban-status) в F10 не выполняется: управление колонками — F13; невалидная COLUMN_ID обрабатывается ответом API (типично not found / ошибка валидации → exit `3` или `1` по классификации F05).
- `kanban list` MAY вызываться без фильтров; `--task` и `--status` опциональны и комбинируемы.
- `task kanban list` MUST NOT принимать `--limit` / `--offset` / `--removed` (и аналоги); пагинация/`includeRemoved` OpenAPI — вне F10.
- `get` / `update` / `delete` связи: pre-check родительской задачи не требуется (достаточно LINK_ID).
- `delete` не требует `--force` и не спрашивает подтверждение в F10; при успехе stdout пуст.
- Успешные create/update/move пишут в stdout полную связь (тот же логический набор полей, что `get`).
- Пустой или whitespace-only ID / TASK_ID / COLUMN_ID / LINK_ID: ошибка использования (exit `1`) до сети.
- `--order` на create/update: если передан, MUST быть числом, допустимым для API-поля порядка (неотрицательное); иначе exit `1` до сети. На `move` флаг `--order` MUST NOT приниматься (ошибка использования / неизвестный флаг → exit `1`); порядок при create через move — default API; при update через move `kanbanOrder` не трогаем.
- Несколько активных связей на одну задачу: `move` отказывается (exit `1`); явный `kanban update`/`delete` по LINK_ID остаётся способом разрешения.
- `kanban create` MUST NOT отказывать только из‑за уже существующих активных связей той же задачи: уникальность не enforcing на клиенте; отказ возможен только от API.
- TUI-диалог перемещения (хоткей `m`) — вне F10; help CLI не обещает TUI.
- CRUD колонок (`project column *`, `/v2/kanban-status`) — вне F10 (F13).
- Поля OpenAPI сверх CLI-surface (`externalId`, `crypted`/`modificated` и т.п.) — вне обязательных флагов F10.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Система MUST предоставлять CLI-подгруппу `task kanban` с подкомандами `list`, `get`, `create`, `update`, `delete` и команду `task move`.
- **FR-002**: `task kanban list` MUST запрашивать список связей и выводить результат в формате F06; MUST принимать опциональные `--task` (→ taskId) и `--status` (→ statusId / COLUMN_ID); MUST NOT выполнять pre-check существования задачи или колонки перед list (в т.ч. при `--task` / `--status`); MUST NOT принимать флаги пагинации или `--removed` / `includeRemoved`.
- **FR-003**: `task kanban get <LINK_ID>` MUST возвращать одну связь по идентификатору.
- **FR-004**: `task kanban create` MUST требовать `--task` и `--column` (непустые после trim; иначе exit `1` до сети); MUST сначала проверить существование задачи (как `task get`); при not found MUST завершаться exit `3` без create; при успехе MUST создавать связь taskId + statusId; MAY принимать `--order` (kanbanOrder); MUST NOT блокировать create на клиенте из‑за уже существующих активных связей той же задачи (уникальность — ответственность API/пользователя; `move` при >1 связи — отдельно, FR-007); stdout при успехе MUST содержать только созданную связь (не тело проверки задачи).
- **FR-005**: `task kanban update <LINK_ID>` MUST принимать опциональные `--task`, `--column`, `--order`; MUST выполнять частичное обновление только переданных полей; при отсутствии изменяющих флагов MUST завершаться ошибкой использования (exit `1`) без вызова API; MUST NOT требовать pre-check задачи.
- **FR-006**: `task kanban delete <LINK_ID>` MUST безвозвратно удалять связь; при успехе stdout MUST быть пустым.
- **FR-007**: `task move <TASK_ID> --column <COLUMN_ID>` MUST: (a) проверить существование задачи (как `task get`; not found → exit `3`); (b) найти активные связи по taskId; (c) при 0 связей — create; при 1 — update statusId этой связи (MUST выполнять update даже если текущий statusId уже равен `--column`); при >1 — exit `1` без create/update; MUST NOT принимать `--order` (и аналоги для kanbanOrder); при create через move порядок — default API; при update через move MUST NOT изменять `kanbanOrder`; stdout при успехе MUST содержать только итоговую связь; промежуточный list MUST NOT попадать в stdout.
- **FR-008**: Успешные `create` / `update` / `move` MUST писать в stdout полную связь с тем же логическим набором полей, что успешный `get` (формат — F06).
- **FR-009**: В `json`/`yaml`: `list` MUST давать корень-массив объектов; успешный `get` / `create` / `update` / `move` MUST давать один объект связи (не массив из одного элемента). Для `table`/`csv`: list — много строк; одна связь — одна запись/строка данных.
- **FR-010**: Все пять operations KanbanTaskStatusController_* (`list`, `create`, `getById`, `update`, `delete`) MUST быть достижимы через адаптерный слой и CLI F10 (`kanban *` и/или `move`).
- **FR-011**: Команды kanban/move MUST соблюдать контракты F06/F07: форматы вывода, отсутствие ANSI в pipe, stdout vs stderr, exit codes `0`/`1`/`2`/`3`.
- **FR-012**: `--help` для `task kanban`, каждой подкоманды и `task move` MUST документировать назначение, аргументы и флаги в scope F10.
- **FR-013**: Адаптер kanban-task-status MUST быть покрыт unit-тестами с мок-HTTP для каждой KanbanTaskStatusController-операции (минимум happy path) без обязательного live API в DoD; логика ветвления `move` (0/1/>1 связей) MUST быть покрыта тестами.
- **FR-014**: Система MUST NOT реализовывать в F10 управление колонками (`/v2/kanban-status`, `project column *`), TUI move-диалог или интерактивные prompt’ы (F13 / TUI / F26).
- **FR-015**: Вызовы API kanban-task-status MUST идти через существующий адаптер/codegen-клиент, без ручных HTTP CRUD/DTO (constitution III).
- **FR-016**: При отсутствии токена или иной ошибке конфигурации команды kanban/move MUST завершаться кодом `2` до сетевого вызова (где применимо).

### Key Entities

- **Kanban task-status link**: Связь задачи с канбан-колонкой — идентификатор связи, ID задачи (`taskId`), ID колонки (`statusId`), порядок в колонке (`kanbanOrder`), признаки удаления/модификации при наличии в ответе API.
- **Kanban link list query**: Выборка связей с опциональными фильтрами по задаче и/или колонке.
- **Kanban link write input**: Поля создания/обновления из CLI (`--task`, `--column`, `--order`).
- **Task move intent**: Перемещение задачи в колонку — UX над create или update единственной активной связи задачи.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Пользователь может выполнить полный цикл «создать связь → получить → изменить колонку → переместить через move → удалить» по документации/`--help` с первой попытки без обращения к исходному коду.
- **SC-002**: 100% operations KanbanTaskStatusController_* из матрицы покрытия достижимы через поставку F10 (CLI + адаптер); управление колонками при этом отсутствует.
- **SC-003**: Для каждой KanbanTaskStatusController-операции есть автоматический unit-тест адаптера с мок-HTTP (happy path), проходящий без доступа к production API; сценарии move 0/1/>1 связей покрыты автоматическими тестами.
- **SC-004**: `singctl task kanban --help`, help пяти подкоманд и `singctl task move --help` перечисляют команды/флаги scope F10; ревьюер за ≤5 минут подтверждает соответствие ТЗ §6.1 (канбан-связь + move).
- **SC-005**: Типичные ошибки (нет токена, not found задачи/связи, misuse флага, неоднозначный move, сбой API) дают коды `2` / `3` / `1` / `1` / `1` соответственно и не смешивают текст ошибки с data-stdout (контракт F07).
- **SC-006**: Скрипт может выполнить `task kanban list --task <ID> --output json` и получить массив записей без ANSI при redirect; `task move -o json` / `kanban get -o json` — один объект связи (не массив), exit `0` при успехе.

## Assumptions

- Зависимости F08 (и транзитивно F01–F07) доступны: группа `task` (включая `task get` для pre-check), конфиг/токен, codegen, адаптер, ошибки/retry, рендер, scriptability.
- CLI-surface совпадает с ТЗ §6.1: `task kanban` list/get/create/update/delete + `task move`; фильтры list — `--task` / `--status`; create/update — `--column` (statusId) и опциональный `--order`.
- F13 (CRUD колонок) рекомендуется для удобства получения COLUMN_ID, но не блокирует F10: колонки задаются явным ID; валидация колонки — на стороне API.
- Pre-check задачи на `create`/`move` использует тот же путь, что получение задачи по ID (F08); лишний round-trip осознан ради явного exit `3`.
- `task kanban list` без pre-check задачи/колонки (clarified); фильтры опциональны.
- «Активная» связь для move — запись из list по `taskId` без включения удалённых (default API `includeRemoved=false`); F10 не передаёт `includeRemoved`.
- При >1 активной связи `move` не выбирает «лучшую» автоматически — пользователь разрешает через явный `kanban` CRUD.
- `kanban create` не upsert и не отказывает при уже существующих связях задачи: всегда POST create после pre-check задачи (clarified).
- `task move` при одной связи и уже совпадающей колонке: всё равно update (clarified); без short-circuit и без пустого stdout.
- `--order` на `move` отсутствует (clarified); порядок при create через move — default API; update через move меняет только колонку (`statusId`).
- Интерактивный ввод и confirm на delete отложены на F26; в F10 — только флаги/аргументы.
- TUI move (хоткей `m`) — вне F10; alias `singctl t` — F25.
- Интеграционные/live API-тесты — F33; в DoD F10 достаточно мок-unit-тестов адаптера + CLI-проверка help/поведения на моках или эквивалентном harness.
- Формат представления связи и колонок списка опирается на общий рендерер F06; разумный минимум колонок table (id, taskId, statusId, kanbanOrder) может уточняться в plan без расширения scope API.
