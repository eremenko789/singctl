# Feature Specification: Project Sections

**Feature Branch**: `012-project-sections`

**Created**: 2026-07-17

**Status**: Draft

**Input**: User description: "F12 project sections — CRUD секций проекта (`project section *`). Depends F11. Scope: CRUD /v2/task-group. Out of scope: kanban columns (F13). Acceptance: `project section *` реализован полностью. Inputs: ТЗ §6.2 (секции), coverage TaskGroupController_*, OpenAPI, constitution."

## Clarifications

### Session 2026-07-17

- Q: Write-флаги секции сверх `--title`? → A: `--title` + `--parent` на update (перенос в другой проект); create по-прежнему через `<PROJECT_ID>`; `--order` / `externalId` / `fake` вне F12.
- Q: Обязателен ли `<PROJECT_ID>` для `section list`? → A: Обязателен; без него exit `1` до сети (как ТЗ §6.2).
- Q: Пустой / whitespace-only `--title`? → A: Exit `1` до сети на create/update.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List and inspect project sections (Priority: P1)

Пользователь или скрипт получает список секций (групп задач) конкретного проекта и может посмотреть одну секцию по идентификатору. Вывод идёт в выбранном формате (table/json/yaml/csv) с разделением stdout/stderr и кодами выхода из F07.

**Why this priority**: Без list/get нельзя ни увидеть структуру проекта после create (API создаёт служебную секцию), ни автоматизировать сценарии вокруг секций; это минимальный read-путь к `TaskGroupController_list` / `TaskGroupController_getById`.

**Independent Test**: С мок-API: `project section list <PROJECT_ID>` передаёт фильтр родителя-проекта и рендерит результат; `project section get <SECTION_ID>` возвращает одну секцию; при «не найдено» — exit `3` и сообщение в stderr.

**Acceptance Scenarios**:

1. **Given** API возвращает набор секций проекта, **When** пользователь выполняет `singctl project section list <PROJECT_ID>`, **Then** в stdout — список в текущем формате вывода (F06), exit `0`, stderr на обычном пути пуст; запрос к API включает идентификатор родительского проекта.
2. **Given** пользователь задаёт опциональные `--removed`, `--limit`, `--offset` (по отдельности или в комбинации) вместе с `<PROJECT_ID>`, **When** выполняется `project section list`, **Then** запрос отражает соответствующие параметры list API (`parent`, `includeRemoved`, `maxCount`, `offset`); при `--limit` вне 1…1000 или отрицательном `--offset` — exit `1` до сети.
3. **Given** известный ID существующей секции, **When** выполняется `singctl project section get <SECTION_ID>` с `-o json` (или `yaml`), **Then** корень stdout — один JSON/YAML-объект секции (не массив); при `-o table`/`csv` — одна запись.
4. **Given** ID несуществующей секции (API «не найдено»), **When** выполняется `project section get` / `update` / `delete`, **Then** exit `3`, сообщение в stderr, stdout пуст.
5. **Given** API возвращает пустой список, **When** выполняется `project section list <PROJECT_ID>` с `-o json`, **Then** stdout — пустой массив `[]` (и валидные пустые представления для других форматов), exit `0`.
6. **Given** API возвращает набор секций, **When** выполняется `project section list <PROJECT_ID> -o json`, **Then** корень — массив объектов секций.
7. **Given** вызов `project section list` без `<PROJECT_ID>`, **When** команда запускается, **Then** ошибка использования (exit `1`) без сетевого вызова (clarification: project ID обязателен).

---

### User Story 2 - Create and update sections (Priority: P1)

Пользователь создаёт секцию в проекте с обязательным заголовком и обновляет существующую секцию (частично — только переданные поля). Поверхность write — ТЗ §6.2 (`--title`) плюс `--parent` на update для переноса секции в другой проект.

**Why this priority**: Вместе с list/get закрывает основной write-путь (`TaskGroupController_create` / `TaskGroupController_update`) и acceptance «`project section *` полностью».

**Independent Test**: Мок: `project section create <PROJECT_ID> --title "…"` создаёт секцию и показывает результат; `project section update <SECTION_ID> --title "…"` и/или `--parent "…"` патчит переданные поля; без `--title` / пустой title на create или update — ошибка использования (exit `1`), без сети.

**Acceptance Scenarios**:

1. **Given** валидный токен и API принимает создание, **When** пользователь выполняет `singctl project section create <PROJECT_ID> --title "Название"`, **Then** секция создаётся в указанном проекте, в stdout — полная секция в том же логическом наборе полей, что у `project section get` (формат F06), exit `0`.
2. **Given** команда `project section create` без `--title`, с пустым/whitespace-only `--title`, или без `<PROJECT_ID>`, **When** она запускается в неинтерактивном режиме, **Then** команда завершается с ошибкой использования (exit `1`), без вызова API; интерактивный prompt — вне scope (F26).
3. **Given** существующая секция, **When** выполняется `singctl project section update <SECTION_ID>` с `--title` и/или `--parent` (по отдельности или вместе; `--title` непустой), **Then** на API уходит частичное обновление только указанных полей, в stdout — полная обновлённая секция (как `get`), exit `0`.
4. **Given** `project section update <SECTION_ID>` без каких-либо изменяющих флагов или с пустым/whitespace-only `--title`, **When** команда выполняется, **Then** ошибка использования (exit `1`) без сетевого вызова.
5. **Given** существующая секция, **When** выполняется `project section update <SECTION_ID> --parent <OTHER_PROJECT_ID>`, **Then** секция переносится в указанный проект (поле `parent` API), в stdout — полная секция с обновлённым родителем, exit `0`.

---

### User Story 3 - Delete section (Priority: P1)

Пользователь безвозвратно удаляет секцию по идентификатору. Подтверждение и интерактивные формы — вне scope (F26); команда scriptable.

**Why this priority**: Закрывает полный lifecycle поверх `TaskGroupController_delete` и симметрию с `project delete` / `task delete`.

**Independent Test**: Мок: `project section delete <SECTION_ID>` вызывает удаление; успех — exit `0`, stdout пуст; not found — exit `3`.

**Acceptance Scenarios**:

1. **Given** существующая секция, **When** выполняется `singctl project section delete <SECTION_ID>`, **Then** секция безвозвратно удаляется, stdout пуст, exit `0`; без интерактивного подтверждения.
2. **Given** ID несуществующей секции, **When** выполняется `project section delete`, **Then** exit `3`, сообщение в stderr, stdout пуст.
3. **Given** пустой или whitespace-only ID, **When** delete (и get/update), **Then** ошибка использования (exit `1`) до сети.

---

### User Story 4 - Discoverable CLI help for section commands (Priority: P2)

Пользователь через `--help` узнаёт подгруппу `project section` и команды list/get/create/update/delete, их аргументы и флаги, согласованные с ТЗ §6.2 (секции), без обещания канбан-колонок.

**Why this priority**: Discoverability и единообразие с `project` / `task`; без help поверхность секций неудобна для скриптов и людей.

**Independent Test**: Вызвать `singctl project section --help` и `--help` каждой подкоманды; убедиться, что все пять команд и ключевые флаги описаны; `column` не обещан как реализованный в F12.

**Acceptance Scenarios**:

1. **Given** установленный CLI, **When** пользователь выполняет `singctl project section --help` (и/или видит `section` в `singctl project --help`), **Then** видны подкоманды `list`, `get`, `create`, `update`, `delete` (и краткое назначение каждой).
2. **Given** пользователь смотрит `singctl project section list --help` (и аналогично для остальных), **When** читает флаги/аргументы, **Then** описаны `<PROJECT_ID>` / `<SECTION_ID>`, `--title` где применимо, `--parent` на update (перенос), фильтры list (`--removed`, `--limit`, `--offset`); нет обещаний `project column` как реализованных в F12.
3. **Given** неизвестная подкоманда или флаг, **When** пользователь ошибается в вызове, **Then** exit `1`, сообщение в stderr (контракт F07).

---

### User Story 5 - Adapter coverage with unit tests (Priority: P1)

Слой поверх API обеспечивает вызовы всех операций секций (`list`, `create`, `getById`, `update`, `delete`) и покрывается unit-тестами с мок-HTTP (happy path и ключевые ошибки), без живого API в обязательном DoD.

**Why this priority**: Constitution III/IV/IX: CLI не должен содержать ручной HTTP; адаптер тестируем независимо от cobra-рендера; acceptance покрывает все TaskGroupController_*.

**Independent Test**: Для каждой из пяти operations TaskGroupController_* — unit-тест адаптера с мок-HTTP (успех); плюс хотя бы not-found или не-2xx на get/update/delete.

**Acceptance Scenarios**:

1. **Given** мок отвечает успехом на list/create/get/update/delete, **When** адаптерный фасад секций выполняет соответствующий вызов, **Then** результат смапплен в доменное/типизированное представление, пригодное для CLI.
2. **Given** мок возвращает «не найдено» на get/update/delete, **When** вызывается адаптер, **Then** ошибка классифицируется так, что CLI может завершиться кодом `3` (F05/F07).
3. **Given** набор тестов адаптера секций, **When** они запускаются без сети к production API, **Then** все обязательные сценарии US5 проходят на моках; фикстуры токенов — только фиктивные (constitution VII).

---

### Edge Cases

- Нет токена / ошибка конфигурации: exit `2`, без вызова API (F02/F07).
- Ошибка API/транспорт (не not-found): exit `1`, сообщение в stderr, stdout пуст (F05/F07).
- `project section list` с `--limit` > 1000 или ≤ 0: ошибка использования (exit `1`) до вызова API (не clamp).
- `delete` не требует `--force` и не спрашивает подтверждение в F12; при успехе stdout пуст.
- Успешные create/update пишут в stdout полную секцию (тот же логический набор полей, что `get`).
- Пустой или whitespace-only project/section ID: ошибка использования (exit `1`) до сети.
- Пустой или whitespace-only `--title` на create/update: ошибка использования (exit `1`) до сети (clarification).
- Канбан-колонки (`project column …`) в F12 отсутствуют (F13); help не обещает их как реализованные.
- Прочие OpenAPI write-поля секции сверх `--title` и `--parent` на update (например `parentOrder`, `externalId`, `fake`) — вне F12; list MAY фильтровать по `includeRemoved` через `--removed`.
- `--parent` на update: пустой/whitespace ID — exit `1` до сети; несуществующий проект-родитель — ошибка API (exit `1`) или not-found (exit `3`) по классификации адаптера F05/F07.
- Soft-delete / archive / trash для секций в API write-DTO не представлены так же, как у project/task — dedicated archive/trash команд для секций в F12 нет; `--removed` только на list.
- Удаление последней/служебной секции проекта: поведение определяется API; CLI честно пробрасывает ошибку API (exit `1`), не изобретая локальную политику.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Система MUST предоставлять CLI-подгруппу `project section` с подкомандами `list`, `get`, `create`, `update`, `delete`.
- **FR-002**: `project section list <PROJECT_ID>` MUST передавать идентификатор проекта как фильтр родителя list API (`parent`) и MUST поддерживать опциональные `--removed` (`includeRemoved`), `--limit` (`maxCount`), `--offset`; `--limit` MUST быть в диапазоне 1…1000 включительно, иначе exit `1` без сети; `--offset` MUST быть ≥ 0, иначе exit `1` без сети; без `<PROJECT_ID>` MUST завершаться ошибкой использования (exit `1`) без сети (clarification).
- **FR-003**: `project section get <SECTION_ID>` MUST возвращать одну секцию по идентификатору.
- **FR-004**: `project section create <PROJECT_ID> --title TITLE` MUST требовать project ID и непустой `--title` (после trim whitespace); MUST создавать секцию с родителем-проектом и заголовком (ТЗ §6.2); пустой/whitespace-only `--title` MUST давать exit `1` без сети (clarification).
- **FR-005**: `project section update <SECTION_ID>` MUST принимать `--title` и/или `--parent` (перенос в другой проект; clarification); MUST выполнять частичное обновление только переданных полей; при отсутствии изменяющих флагов MUST завершаться ошибкой использования (exit `1`) без вызова API; если передан `--title`, значение MUST быть непустым после trim, иначе exit `1` без сети (clarification).
- **FR-005a**: Create MUST NOT принимать отдельный флаг `--parent`: родитель задаётся только позиционным `<PROJECT_ID>`. `--order` (`parentOrder`), `externalId`, `fake` MUST NOT входить в CLI F12.
- **FR-006**: `project section delete <SECTION_ID>` MUST безвозвратно удалять секцию через операцию delete API; при успехе stdout MUST быть пустым.
- **FR-007**: Успешные `create` / `update` MUST писать в stdout полную секцию с тем же логическим набором полей, что успешный `get` (формат — F06).
- **FR-008**: В `json`/`yaml`: `project section list` MUST давать корень-массив объектов; успешный `get` / `create` / `update` MUST давать один объект секции (не массив из одного элемента). Для `table`/`csv`: list — много строк; одна секция — одна запись/строка данных.
- **FR-009**: Все пять operations TaskGroupController_* (`list`, `create`, `getById`, `update`, `delete`) MUST быть достижимы через адаптерный слой и CLI F12.
- **FR-010**: Команды section MUST соблюдать контракты F06/F07: форматы вывода, отсутствие ANSI в pipe, stdout vs stderr, exit codes `0`/`1`/`2`/`3`.
- **FR-011**: `--help` для `project section` и каждой подкоманды MUST документировать назначение, аргументы и флаги в scope F12 (включая `--parent` на update); `project --help` MUST упоминать подгруппу `section` как доступную.
- **FR-012**: Адаптер секций MUST быть покрыт unit-тестами с мок-HTTP для каждой TaskGroupController-операции (минимум happy path) без обязательного live API в DoD.
- **FR-013**: Система MUST NOT реализовывать в F12 канбан-колонки или интерактивные prompt’ы создания/подтверждения удаления (F13/F26).
- **FR-014**: Вызовы API секций MUST идти через существующий адаптер/codegen-клиент, без ручных HTTP CRUD/DTO (constitution III).
- **FR-015**: При отсутствии токена или иной ошибке конфигурации команды section MUST завершаться кодом `2` до сетевого вызова (где применимо).
- **FR-016**: Документация/help F12 MUST использовать термин «секция» (section) для пользователя; внутреннее имя API «task group» MAY фигурировать только в технических контрактах/адаптере, не как основное UX-название команд.

### Key Entities

- **Section (task group)**: Группа задач внутри проекта SingularityApp — идентификатор, заголовок, родительский проект, порядок в родителе, признаки removed/fake и прочие поля ответа API, релевантные для отображения.
- **Section list query**: Набор фильтров и пагинации для выборки секций (родительский проект, включение удалённых, limit/offset).
- **Section write input**: Поля создания/обновления из CLI (ТЗ §6.2: заголовок; для create — родительский проект через `<PROJECT_ID>`; для update — опционально `--parent` для переноса).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Пользователь может выполнить полный цикл «создать секцию → получить → изменить → удалить» по документации/`--help` с первой попытки без обращения к исходному коду.
- **SC-002**: 100% operations TaskGroupController_* из матрицы покрытия достижимы через поставку F12 (CLI + адаптер); канбан-колонки при этом отсутствуют.
- **SC-003**: Для каждой TaskGroupController-операции есть автоматический unit-тест адаптера с мок-HTTP (happy path), проходящий без доступа к production API.
- **SC-004**: `singctl project section --help` и help каждой из пяти подкоманд перечисляют команды/флаги scope F12 (включая `--parent` на update); ревьюер за ≤5 минут подтверждает соответствие ТЗ §6.2 (секции) + clarification без column.
- **SC-005**: Типичные ошибки (нет токена, not found, misuse флага/аргумента, сбой API) дают коды `2` / `3` / `1` / `1` соответственно и не смешивают текст ошибки с data-stdout (контракт F07).
- **SC-006**: Скрипт может выполнить `project section list <PROJECT_ID> --output json` и получить массив записей без ANSI при redirect; `project section get -o json` — один объект секции (не массив), exit `0` при успехе.

## Assumptions

- Зависимость F11 (project CRUD) и транзитивно F01–F07 доступны как база; паттерны F08/F11 (entity CRUD, выход, help) — ориентир UX/контрактов.
- Базовая поверхность create/update — таблица ТЗ §6.2 для секций (`--title`); create получает `<PROJECT_ID>` как parent. На update дополнительно `--parent` для переноса (clarification). Пустой/whitespace-only `--title` — exit `1` до сети (clarification). Прочие OpenAPI write-поля (`parentOrder`, `externalId`, `fake`) вне F12.
- List в ТЗ показывает только `<PROJECT_ID>`; `<PROJECT_ID>` обязателен (clarification); `--limit` / `--offset` / `--removed` добавлены по аналогии с `project list` / параметрами OpenAPI list task-group — без расширения write-scope. List без фильтра «все секции аккаунта» в F12 не поддерживается.
- Archive/trash для секций не входят в F12: в TaskGroup update DTO нет дат архива/корзины как у project/task; soft-removed виден через `--removed` на list, если API так отдаёт.
- Термин CLI — `section`; API resource — task-group / TaskGroup*; маппинг имён — в plan/contracts.
- Интерактивный ввод и confirm на delete отложены на F26.
- `project column` — F13; короткие alias-команды — F25.
- Интеграционные/live API-тесты — F33; в DoD F12 достаточно мок-unit-тестов адаптера + CLI-проверка help/поведения на моках или эквивалентном harness.
- Формат представления одной секции и колонок списка опирается на общий рендерер F06; конкретный набор колонок table (id, title, parent, …) выбирается разумным минимумом и может уточняться в plan без расширения scope API.
- Поведение API при удалении последней/служебной секции проекта не дублируется локальной политикой CLI.
