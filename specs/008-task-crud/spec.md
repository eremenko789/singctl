# Feature Specification: Task CRUD

**Feature Branch**: `008-task-crud`

**Created**: 2026-07-16

**Status**: Draft

**Input**: User description: "F08 task CRUD — list/get/create/update/delete + archive/trash; фильтры для list. Depends F07. Out of scope: checklist (F09), kanban-link (F10). Acceptance: закрыты все operations TaskController_*; CLI help; unit-тесты адаптера. Inputs: ТЗ §6.1, coverage /v2/task, OpenAPI, constitution."

## Clarifications

### Session 2026-07-16

- Q: Флаги `--project` / `--parent` на create/update? → A: Добавить оба (`--project` → `projectId`, `--parent` → `parent`); полный dump OpenAPI write-полей — вне F08.
- Q: Stdout при успешном create/update/archive/trash/delete? → A: create/update/archive/trash → полная задача (как `get`); delete → пустой stdout.
- Q: Поведение `--limit` выше максимума API (1000)? → A: CLI отклоняет `--limit` > 1000 и ≤ 0 с exit `1` до вызова API.
- Q: Форма json/yaml для одной задачи vs list? → A: `list` → массив объектов; одна задача (`get`/create/update/archive/trash) → один объект без обёртки-массива.
- Q: Как обрабатывать `--note` (API: delta)? → A: Передавать строку как есть; в help указать, что API может ожидать delta; без клиентской конвертации в F08.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List and inspect tasks (Priority: P1)

Пользователь или скрипт получает список задач с фильтрами (проект, родитель, диапазон дат, архив/корзина, пагинация, экземпляры повторений) и может посмотреть одну задачу по идентификатору. Вывод идёт в выбранном формате (table/json/yaml/csv) с разделением stdout/stderr и кодами выхода из F07.

**Why this priority**: Без list/get нельзя ни автоматизировать пайплайны ТЗ §10, ни опираться на CRUD; это минимальный read-путь к `TaskController_list` / `TaskController_getById`.

**Independent Test**: С мок-API: `task list` с набором фильтров передаёт ожидаемые параметры и рендерит результат; `task get <ID>` возвращает одну задачу; при «не найдено» — exit `3` и сообщение в stderr.

**Acceptance Scenarios**:

1. **Given** API возвращает набор задач, **When** пользователь выполняет `singctl task list`, **Then** в stdout — список в текущем формате вывода (F06), exit `0`, stderr на обычном пути пуст.
2. **Given** пользователь задаёт фильтры `--project`, `--parent`, `--from`, `--to`, `--archived`, `--removed`, `--limit`, `--offset`, `--all-recurrence` (по отдельности или в комбинации), **When** выполняется `task list`, **Then** запрос к API отражает соответствующие параметры списка из ТЗ §6.1 / OpenAPI, а вывод содержит только результат этого запроса; при `--limit` вне 1…1000 или отрицательном `--offset` — exit `1` до сети.
3. **Given** известный ID существующей задачи, **When** выполняется `singctl task get <ID>` с `-o json` (или `yaml`), **Then** корень stdout — один JSON/YAML-объект задачи (не массив); при `-o table`/`csv` — одна запись.
4. **Given** ID несуществующей задачи (API «не найдено»), **When** выполняется `task get` / `update` / `delete` / `archive` / `trash`, **Then** exit `3`, сообщение в stderr, stdout пуст.
5. **Given** API возвращает пустой список, **When** выполняется `task list` с `-o json`, **Then** stdout — пустой массив `[]` (и валидные пустые представления для других форматов), exit `0`.
6. **Given** API возвращает набор задач, **When** выполняется `task list -o json`, **Then** корень — массив объектов задач.

---

### User Story 2 - Create and update tasks (Priority: P1)

Пользователь создаёт задачу с обязательным заголовком и опциональными полями (ТЗ §6.1 плюс `--project` / `--parent`), а также обновляет существующую задачу частичным набором тех же флагов (все опциональны при update).

**Why this priority**: Вместе с list/get закрывает основной write-путь (`TaskController_create` / `TaskController_update`) и acceptance «все operations TaskController_*».

**Independent Test**: Мок: `task create --title "…"` создаёт задачу и показывает результат; `task update <ID> --note "…"` патчит только переданные поля; без `--title` на create — ошибка использования (exit `1`), без сети.

**Acceptance Scenarios**:

1. **Given** валидный токен и API принимает создание, **When** пользователь выполняет `singctl task create --title "Название"` (опционально с `--project`, `--parent`, `--start`, `--note`, `--priority`, `--is-note`, `--archive-date`, `--delete-date`), **Then** задача создаётся, в stdout — полная задача в том же логическом наборе полей, что у `task get` (формат F06), exit `0`.
2. **Given** команда `task create` без `--title`, **When** она запускается в неинтерактивном режиме, **Then** команда завершается с ошибкой использования (exit `1`), без вызова API; интерактивный prompt заголовка — вне scope (F26).
3. **Given** существующая задача, **When** выполняется `singctl task update <ID>` с одним или несколькими флагами из набора create (включая `--project` / `--parent`; все опциональны), **Then** на API уходит частичное обновление только указанных полей, в stdout — полная обновлённая задача (как `get`), exit `0`.
4. **Given** `task update <ID>` без каких-либо изменяющих флагов, **When** команда выполняется, **Then** ошибка использования (exit `1`) без сетевого вызова.
5. **Given** неверное значение `--priority` (не 0/1/2), **When** create/update, **Then** ошибка использования (exit `1`) до вызова API.

---

### User Story 3 - Archive, trash, and permanent delete (Priority: P1)

Пользователь архивирует задачу, перемещает в корзину или безвозвратно удаляет её. Archive и trash — удобные команды над обновлением дат (`journalDate` / `deleteDate`), а не отдельные API-ресурсы; delete вызывает безвозвратное удаление.

**Why this priority**: Явный scope F08 и покрытие `TaskController_update` (через archive/trash) + `TaskController_delete` из coverage.

**Independent Test**: Мок: `archive`/`trash` отправляют обновление с соответствующей датой; `delete` вызывает удаление; успех — exit `0`; not found — exit `3`.

**Acceptance Scenarios**:

1. **Given** существующая задача, **When** выполняется `singctl task archive <ID>` (опционально `--date DATE`), **Then** задача помечается архивной на указанную (или дату по умолчанию — «сегодня»), в stdout — полная задача (как `get`), exit `0`.
2. **Given** существующая задача, **When** выполняется `singctl task trash <ID>` (опционально `--date DATE`), **Then** задаче выставляется дата удаления/корзины на указанную (или «сегодня»), в stdout — полная задача (как `get`), exit `0`.
3. **Given** существующая задача, **When** выполняется `singctl task delete <ID>`, **Then** задача безвозвратно удаляется, stdout пуст, exit `0`; без интерактивного подтверждения (scriptability; интерактивные формы — F26).
4. **Given** неверный формат `--date`, **When** archive/trash, **Then** ошибка использования (exit `1`) до вызова API.

---

### User Story 4 - Discoverable CLI help for task commands (Priority: P2)

Пользователь через `--help` узнаёт группу `task` и подкоманды list/get/create/update/delete/archive/trash, их аргументы и флаги, согласованные с ТЗ §6.1.

**Why this priority**: Acceptance F08 («CLI help»); без discoverability первая entity-группа неудобна и ломает единообразие с config-командами.

**Independent Test**: Вызвать `singctl task --help` и `--help` каждой подкоманды; убедиться, что все семь команд и ключевые флаги описаны.

**Acceptance Scenarios**:

1. **Given** установленный CLI, **When** пользователь выполняет `singctl task --help`, **Then** видны подкоманды `list`, `get`, `create`, `update`, `delete`, `archive`, `trash` (и краткое назначение каждой).
2. **Given** пользователь смотрит `singctl task list --help` (и аналогично для остальных), **When** читает флаги, **Then** описаны фильтры/поля scope F08 (ТЗ §6.1 + `--project`/`--parent` на create/update); для `--note` — пометка про возможный delta-формат API; нет обещаний checklist/kanban/`move`.
3. **Given** неизвестная подкоманда или флаг, **When** пользователь ошибается в вызове, **Then** exit `1`, сообщение в stderr (контракт F07).

---

### User Story 5 - Adapter coverage with unit tests (Priority: P1)

Слой поверх API обеспечивает вызовы всех операций задач (`list`, `create`, `getById`, `update`, `delete`) и покрывается unit-тестами с мок-HTTP (happy path и ключевые ошибки), без живого API в обязательном DoD.

**Why this priority**: Acceptance F08 («unit-тесты адаптера») и constitution III/IV/IX: CLI не должен содержать ручной HTTP; адаптер тестируем независимо от cobra-рендера.

**Independent Test**: Для каждой из пяти operations TaskController_* — unit-тест адаптера с мок-HTTP (успех); плюс хотя бы not-found или не-2xx на get/update/delete.

**Acceptance Scenarios**:

1. **Given** мок отвечает успехом на list/create/get/update/delete, **When** адаптерный фасад задач выполняет соответствующий вызов, **Then** результат смапплен в доменное/типизированное представление, пригодное для CLI.
2. **Given** мок возвращает «не найдено» на get/update/delete, **When** вызывается адаптер, **Then** ошибка классифицируется так, что CLI может завершиться кодом `3` (F05/F07).
3. **Given** набор тестов адаптера задач, **When** они запускаются без сети к production API, **Then** все обязательные сценарии US5 проходят на моках; фикстуры токенов — только фиктивные (constitution VII).

---

### Edge Cases

- Нет токена / ошибка конфигурации: exit `2`, без вызова API (F02/F07).
- Ошибка API/транспорт (не not-found): exit `1`, сообщение в stderr, stdout пуст (F05/F07).
- `task list` с `--limit` > 1000 или ≤ 0: ошибка использования (exit `1`) до вызова API (не clamp, не «проброс» в API).
- Одновременные `--archived` / `--removed`: допустимы; оба фильтра передаются в API как есть.
- Archive/trash без `--date`: используется дата «сегодня» (календарный день по согласованной политике дат проекта).
- `delete` не требует `--force` и не спрашивает подтверждение в F08; при успехе stdout пуст.
- Успешные create/update/archive/trash пишут в stdout полную задачу (тот же логический набор полей, что `get`).
- Пустой или whitespace-only ID: ошибка использования (exit `1`) до сети.
- Checklist- и kanban-подкоманды / `task move` в F08 отсутствуют (F09/F10); help не обещает их как реализованные.
- `--note` передаётся как есть (без delta-конвертера); прочие OpenAPI write-поля сверх clarified набора — вне F08.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Система MUST предоставлять CLI-группу `task` с подкомандами `list`, `get`, `create`, `update`, `delete`, `archive`, `trash`.
- **FR-002**: `task list` MUST поддерживать фильтры ТЗ §6.1: `--project`, `--parent`, `--from`, `--to`, `--archived`, `--removed`, `--limit`, `--offset`, `--all-recurrence`, с маппингом на параметры list API задач; `--limit` MUST быть в диапазоне 1…1000 включительно, иначе exit `1` без сети; `--offset` MUST быть ≥ 0, иначе exit `1` без сети.
- **FR-003**: `task get <ID>` MUST возвращать одну задачу по идентификатору.
- **FR-004**: `task create` MUST требовать `--title` и MAY принимать `--project` (`projectId`), `--parent` (`parent`), `--start`, `--note`, `--priority` (0/1/2), `--is-note`, `--archive-date`, `--delete-date` (ТЗ §6.1 + clarification: project/parent).
- **FR-004a**: Значение `--note` MUST передаваться в API без клиентской конвертации в delta; help MUST честно отметить, что API может ожидать delta-формат.
- **FR-005**: `task update <ID>` MUST принимать тот же набор полей, что create, все опционально; MUST выполнять частичное обновление только переданных полей; при отсутствии изменяющих флагов MUST завершаться ошибкой использования (exit `1`) без вызова API.
- **FR-006**: `task archive <ID> [--date DATE]` MUST обновлять архивную дату задачи (`journalDate`); без `--date` — дата по умолчанию «сегодня».
- **FR-007**: `task trash <ID> [--date DATE]` MUST обновлять дату удаления/корзины (`deleteDate`); без `--date` — «сегодня».
- **FR-008**: `task delete <ID>` MUST безвозвратно удалять задачу через операцию delete API; при успехе stdout MUST быть пустым.
- **FR-008a**: Успешные `create` / `update` / `archive` / `trash` MUST писать в stdout полную задачу с тем же логическим набором полей, что успешный `get` (формат — F06).
- **FR-008b**: В `json`/`yaml`: `task list` MUST давать корень-массив объектов; успешный `get` / `create` / `update` / `archive` / `trash` MUST давать один объект задачи (не массив из одного элемента). Для `table`/`csv`: list — много строк; одна задача — одна запись/строка данных.
- **FR-009**: Все пять operations TaskController_* (`list`, `create`, `getById`, `update`, `delete`) MUST быть достижимы через адаптерный слой и CLI F08 (archive/trash используют `update`).
- **FR-010**: Команды task MUST соблюдать контракты F06/F07: форматы вывода, отсутствие ANSI в pipe, stdout vs stderr, exit codes `0`/`1`/`2`/`3`.
- **FR-011**: `--help` для `task` и каждой подкоманды MUST документировать назначение, аргументы и флаги в scope F08.
- **FR-012**: Адаптер задач MUST быть покрыт unit-тестами с мок-HTTP для каждой TaskController-операции (минимум happy path) без обязательного live API в DoD.
- **FR-013**: Система MUST NOT реализовывать в F08 checklist, kanban-link, `task move` или интерактивные prompt’ы создания (F09/F10/F26).
- **FR-014**: Вызовы API задач MUST идти через существующий адаптер/codegen-клиент, без ручных HTTP CRUD/DTO (constitution III).
- **FR-015**: При отсутствии токена или иной ошибке конфигурации команды task MUST завершаться кодом `2` до сетевого вызова (где применимо).

### Key Entities

- **Task**: Рабочая единица пользователя SingularityApp — идентификатор, заголовок, заметка, приоритет, даты (start, archive/journal, delete/trash), признаки (isNote и др. из ответа), связь с проектом/родителем при наличии в ответе API.
- **Task list query**: Набор фильтров и пагинации для выборки задач (проект, родитель, диапазон start, включение архивных/удалённых, recurrence, limit/offset).
- **Task write input**: Набор полей создания/обновления из CLI (ТЗ §6.1 + `--project` / `--parent`); для archive/trash — целевая дата.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Пользователь может выполнить полный цикл «создать → получить → изменить → архивировать/в корзину → удалить» по документации/`--help` с первой попытки без обращения к исходному коду.
- **SC-002**: 100% operations TaskController_* из матрицы покрытия достижимы через поставку F08 (CLI + адаптер); checklist/kanban при этом отсутствуют.
- **SC-003**: Для каждой TaskController-операции есть автоматический unit-тест адаптера с мок-HTTP (happy path), проходящий без доступа к production API.
- **SC-004**: `singctl task --help` и help каждой из семи подкоманд перечисляют команды/флаги scope F08; ревьюер за ≤5 минут подтверждает соответствие ТЗ §6.1 (без checklist/kanban).
- **SC-005**: Типичные ошибки (нет токена, not found, misuse флага, сбой API) дают коды `2` / `3` / `1` / `1` соответственно и не смешивают текст ошибки с data-stdout (контракт F07).
- **SC-006**: Скрипт может выполнить `task list --output json` и получить массив записей без ANSI при redirect; `task get -o json` — один объект задачи (не массив), exit `0` при успехе.

## Assumptions

- Зависимости F07 (и транзитивно F01–F06: CLI-скелет, конфиг/токен, codegen, адаптер, ошибки/retry, рендер, scriptability) доступны как база для команд task.
- Базовая поверхность create/update — таблица ТЗ §6.1 плюс `--project` / `--parent`; `--note` — as-is без delta-конвертера; прочие OpenAPI write-поля (tags, deadline, …) вне F08.
- Дата по умолчанию для archive/trash без `--date` — «сегодня» в согласованном формате дат проекта (как в существующих date-хелперах).
- Интерактивный ввод `--title` и confirm на delete отложены на F26; в F08 — только флаги/аргументы.
- `task move`, checklist и kanban-команды — F09/F10; alias `singctl t` — F25.
- Интеграционные/live API-тесты — F33; в DoD F08 достаточно мок-unit-тестов адаптера + CLI-проверка help/поведения на моках или эквивалентном harness.
- Формат представления одной задачи и колонок списка опирается на общий рендерер F06; конкретный набор колонок table для task выбирается разумным минимумом (id, title, ключевые даты/статус) и может уточняться в plan без расширения scope API.
