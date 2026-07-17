# Feature Specification: Project CRUD

**Feature Branch**: `011-project-crud`

**Created**: 2026-07-17

**Status**: Draft

**Input**: User description: "F11 project CRUD — list/get/create/update/delete для проекта. Depends F07. Scope: CRUD /v2/project. Out of scope: sections (F12), columns (F13). Acceptance: list/get/create/update/delete. Inputs: ТЗ §6.2, coverage ProjectController_*, OpenAPI, constitution."

## Clarifications

### Session 2026-07-17

- Q: Archive/trash для проектов в F11? → A: Как у task: `project archive` / `project trash` поверх update дат (`journalDate` / `deleteDate`); hard `delete` остаётся.
- Q: Флаг `--parent` на create/update? → A: Добавить на create и update (`parent`); как у task.
- Q: Даты архива/корзины на create/update? → A: Только команды `archive`/`trash`; без `--archive-date`/`--delete-date` на create/update.
- Q: Формат `--emoji`? → A: Принимать unicode-символ и конвертировать в hex в CLI; hex as-is тоже допустим.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List and inspect projects (Priority: P1)

Пользователь или скрипт получает список проектов с фильтрами (архив, корзина/удалённые, пагинация) и может посмотреть один проект по идентификатору. Вывод идёт в выбранном формате (table/json/yaml/csv) с разделением stdout/stderr и кодами выхода из F07.

**Why this priority**: Без list/get нельзя ни ориентироваться в workspace, ни автоматизировать сценарии вокруг проектов; это минимальный read-путь к `ProjectController_list` / `ProjectController_getById`.

**Independent Test**: С мок-API: `project list` с набором фильтров передаёт ожидаемые параметры и рендерит результат; `project get <ID>` возвращает один проект; при «не найдено» — exit `3` и сообщение в stderr.

**Acceptance Scenarios**:

1. **Given** API возвращает набор проектов, **When** пользователь выполняет `singctl project list`, **Then** в stdout — список в текущем формате вывода (F06), exit `0`, stderr на обычном пути пуст.
2. **Given** пользователь задаёт фильтры `--archived`, `--removed`, `--limit`, `--offset` (по отдельности или в комбинации), **When** выполняется `project list`, **Then** запрос к API отражает соответствующие параметры списка из ТЗ §6.2 / OpenAPI (`includeArchived`, `includeRemoved`, `maxCount`, `offset`), а вывод содержит только результат этого запроса; при `--limit` вне 1…1000 или отрицательном `--offset` — exit `1` до сети.
3. **Given** известный ID существующего проекта, **When** выполняется `singctl project get <ID>` с `-o json` (или `yaml`), **Then** корень stdout — один JSON/YAML-объект проекта (не массив); при `-o table`/`csv` — одна запись.
4. **Given** ID несуществующего проекта (API «не найдено»), **When** выполняется `project get` / `update` / `delete` / `archive` / `trash`, **Then** exit `3`, сообщение в stderr, stdout пуст.
5. **Given** API возвращает пустой список, **When** выполняется `project list` с `-o json`, **Then** stdout — пустой массив `[]` (и валидные пустые представления для других форматов), exit `0`.
6. **Given** API возвращает набор проектов, **When** выполняется `project list -o json`, **Then** корень — массив объектов проектов.

---

### User Story 2 - Create and update projects (Priority: P1)

Пользователь создаёт проект с обязательным заголовком и опциональными полями из ТЗ §6.2 (`--note`, `--notebook`, `--emoji`, `--color`) плюс `--parent` для вложенности, а также обновляет существующий проект частичным набором тех же флагов (все опциональны при update).

**Why this priority**: Вместе с list/get закрывает основной write-путь (`ProjectController_create` / `ProjectController_update`) и acceptance «list/get/create/update/delete».

**Independent Test**: Мок: `project create --title "…"` создаёт проект и показывает результат; `project update <ID> --note "…"` патчит только переданные поля; без `--title` на create — ошибка использования (exit `1`), без сети.

**Acceptance Scenarios**:

1. **Given** валидный токен и API принимает создание, **When** пользователь выполняет `singctl project create --title "Название"` (опционально с `--note`, `--notebook`, `--emoji`, `--color`, `--parent`), **Then** проект создаётся, в stdout — полный проект в том же логическом наборе полей, что у `project get` (формат F06), exit `0`.
2. **Given** команда `project create` без `--title`, **When** она запускается в неинтерактивном режиме, **Then** команда завершается с ошибкой использования (exit `1`), без вызова API; интерактивный prompt заголовка — вне scope (F26).
3. **Given** существующий проект, **When** выполняется `singctl project update <ID>` с одним или несколькими флагами из набора create (включая `--parent`; все опциональны), **Then** на API уходит частичное обновление только указанных полей, в stdout — полный обновлённый проект (как `get`), exit `0`.
4. **Given** `project update <ID>` без каких-либо изменяющих флагов, **When** команда выполняется, **Then** ошибка использования (exit `1`) без сетевого вызова.
5. **Given** API при создании также возвращает служебную секцию по умолчанию (side-effect бэкенда), **When** выполняется успешный `project create`, **Then** CLI показывает пользователю проект (как `get`); управление секциями в F11 не предлагается и не требуется.
6. **Given** create/update с `--emoji` в виде unicode-символа (напр. `💞`), **When** команда выполняется, **Then** в API уходит hex-код (напр. `1f49e`); при уже hex-строке — без изменения; при неоднозначном значении — exit `1` до сети.

---

### User Story 3 - Archive, trash, and permanent delete (Priority: P1)

Пользователь архивирует проект, перемещает в корзину или безвозвратно удаляет его. Archive и trash — удобные команды над обновлением дат (`journalDate` / `deleteDate`), по аналогии с `task archive` / `task trash`; delete вызывает безвозвратное удаление.

**Why this priority**: Симметрия с фильтрами `--archived` / `--removed` на list и с паттерном F08; закрывает полный lifecycle поверх `ProjectController_update` + `ProjectController_delete`.

**Independent Test**: Мок: `archive`/`trash` отправляют обновление с соответствующей датой; `delete` вызывает удаление; успех — exit `0`; not found — exit `3`.

**Acceptance Scenarios**:

1. **Given** существующий проект, **When** выполняется `singctl project archive <ID>` (опционально `--date DATE`), **Then** проект помечается архивным на указанную (или дату по умолчанию — «сегодня»), в stdout — полный проект (как `get`), exit `0`.
2. **Given** существующий проект, **When** выполняется `singctl project trash <ID>` (опционально `--date DATE`), **Then** проекту выставляется дата удаления/корзины на указанную (или «сегодня»), в stdout — полный проект (как `get`), exit `0`.
3. **Given** существующий проект, **When** выполняется `singctl project delete <ID>`, **Then** проект безвозвратно удаляется, stdout пуст, exit `0`; без интерактивного подтверждения (scriptability; интерактивные формы — F26).
4. **Given** неверный формат `--date`, **When** archive/trash, **Then** ошибка использования (exit `1`) до вызова API.
5. **Given** ID несуществующего проекта, **When** выполняется `project delete` / `archive` / `trash`, **Then** exit `3`, сообщение в stderr, stdout пуст.

---

### User Story 4 - Discoverable CLI help for project commands (Priority: P2)

Пользователь через `--help` узнаёт группу `project` и подкоманды list/get/create/update/delete/archive/trash, их аргументы и флаги, согласованные с ТЗ §6.2 плюс archive/trash (без секций и колонок).

**Why this priority**: Discoverability и единообразие с `task`/`config`; без help первая entity-группа проектов неудобна для скриптов и людей.

**Independent Test**: Вызвать `singctl project --help` и `--help` каждой подкоманды; убедиться, что все семь команд и ключевые флаги описаны; секции/колонки не обещаны как реализованные.

**Acceptance Scenarios**:

1. **Given** установленный CLI, **When** пользователь выполняет `singctl project --help`, **Then** видны подкоманды `list`, `get`, `create`, `update`, `delete`, `archive`, `trash` (и краткое назначение каждой).
2. **Given** пользователь смотрит `singctl project list --help` (и аналогично для остальных), **When** читает флаги, **Then** описаны фильтры/поля scope F11 (ТЗ §6.2 + `--parent` на create/update; archive/trash — `--date`); для `--note` — пометка про возможный delta-формат API; для `--emoji` — примеры unicode и hex; нет обещаний `section` / `column` как реализованных в F11.
3. **Given** неизвестная подкоманда или флаг, **When** пользователь ошибается в вызове, **Then** exit `1`, сообщение в stderr (контракт F07).

---

### User Story 5 - Adapter coverage with unit tests (Priority: P1)

Слой поверх API обеспечивает вызовы всех операций проектов (`list`, `create`, `getById`, `update`, `delete`) и покрывается unit-тестами с мок-HTTP (happy path и ключевые ошибки), без живого API в обязательном DoD.

**Why this priority**: Constitution III/IV/IX: CLI не должен содержать ручной HTTP; адаптер тестируем независимо от cobra-рендера; acceptance покрывает все ProjectController_*.

**Independent Test**: Для каждой из пяти operations ProjectController_* — unit-тест адаптера с мок-HTTP (успех); плюс хотя бы not-found или не-2xx на get/update/delete.

**Acceptance Scenarios**:

1. **Given** мок отвечает успехом на list/create/get/update/delete, **When** адаптерный фасад проектов выполняет соответствующий вызов, **Then** результат смапплен в доменное/типизированное представление, пригодное для CLI.
2. **Given** мок возвращает «не найдено» на get/update/delete, **When** вызывается адаптер, **Then** ошибка классифицируется так, что CLI может завершиться кодом `3` (F05/F07).
3. **Given** набор тестов адаптера проектов, **When** они запускаются без сети к production API, **Then** все обязательные сценарии US5 проходят на моках; фикстуры токенов — только фиктивные (constitution VII).

---

### Edge Cases

- Нет токена / ошибка конфигурации: exit `2`, без вызова API (F02/F07).
- Ошибка API/транспорт (не not-found): exit `1`, сообщение в stderr, stdout пуст (F05/F07).
- `project list` с `--limit` > 1000 или ≤ 0: ошибка использования (exit `1`) до вызова API (не clamp, не «проброс» в API).
- Одновременные `--archived` / `--removed`: допустимы; оба фильтра передаются в API как есть.
- `delete` не требует `--force` и не спрашивает подтверждение в F11; при успехе stdout пуст.
- Успешные create/update/archive/trash пишут в stdout полный проект (тот же логический набор полей, что `get`).
- Archive/trash без `--date`: используется дата «сегодня» (календарный день по согласованной политике дат проекта, как у task).
- Create/update не принимают `--archive-date` / `--delete-date`; для архива/корзины — только `project archive` / `project trash`.
- Пустой или whitespace-only ID: ошибка использования (exit `1`) до сети.
- Совместные (shared) проекты API не возвращает в list — CLI MUST NOT обещать их в help/docs F11 (constitution VI).
- Секции (`project section …`) и колонки (`project column …`) в F11 отсутствуют (F12/F13); help не обещает их как реализованные.
- `--note` передаётся как есть (без delta-конвертера); прочие OpenAPI write-поля сверх поверхности ТЗ §6.2 + `--parent` + archive/trash дат (review, start/end на write, parentOrder, …) — вне F11.
- `--emoji`: unicode → hex в CLI; hex-строка → as-is; неоднозначное/невалидное значение — exit `1` до сети; `--color` — строка HEX as-is, без обязательной валидации палитры wiki в F11.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Система MUST предоставлять CLI-группу `project` с подкомандами `list`, `get`, `create`, `update`, `delete`, `archive`, `trash`.
- **FR-002**: `project list` MUST поддерживать фильтры ТЗ §6.2: `--archived`, `--removed`, `--limit`, `--offset`, с маппингом на параметры list API проектов (`includeArchived`, `includeRemoved`, `maxCount`, `offset`); `--limit` MUST быть в диапазоне 1…1000 включительно, иначе exit `1` без сети; `--offset` MUST быть ≥ 0, иначе exit `1` без сети.
- **FR-003**: `project get <ID>` MUST возвращать один проект по идентификатору.
- **FR-004**: `project create` MUST требовать `--title` и MAY принимать `--note`, `--notebook` (`isNotebook`), `--emoji`, `--color` (ТЗ §6.2) и `--parent` (`parent`; clarification).
- **FR-004a**: Значение `--note` MUST передаваться в API без клиентской конвертации в delta; help MUST честно отметить, что API может ожидать delta-формат.
- **FR-004b**: `--emoji` MUST принимать либо unicode-символ эмодзи (конвертация в hex-код перед вызовом API), либо уже hex-строку (передача as-is); help MUST привести оба примера; значение, которое CLI не может однозначно трактовать как эмодзи или hex, MUST давать ошибку использования (exit `1`) до сети.
- **FR-005**: `project update <ID>` MUST принимать тот же набор полей, что create (включая `--parent`), все опционально; MUST выполнять частичное обновление только переданных полей; при отсутствии изменяющих флагов MUST завершаться ошибкой использования (exit `1`) без вызова API.
- **FR-006**: `project archive <ID> [--date DATE]` MUST обновлять архивную дату проекта (`journalDate`); без `--date` — дата по умолчанию «сегодня».
- **FR-006a**: `project trash <ID> [--date DATE]` MUST обновлять дату удаления/корзины (`deleteDate`); без `--date` — «сегодня».
- **FR-006b**: `project delete <ID>` MUST безвозвратно удалять проект через операцию delete API; при успехе stdout MUST быть пустым.
- **FR-006c**: Create/update MUST NOT принимать флаги `--archive-date` / `--delete-date`; архив и корзина — только через `archive` / `trash`.
- **FR-007**: Успешные `create` / `update` / `archive` / `trash` MUST писать в stdout полный проект с тем же логическим набором полей, что успешный `get` (формат — F06); побочная секция из ответа create MUST NOT становиться предметом CLI-CRUD в F11.
- **FR-008**: В `json`/`yaml`: `project list` MUST давать корень-массив объектов; успешный `get` / `create` / `update` / `archive` / `trash` MUST давать один объект проекта (не массив из одного элемента). Для `table`/`csv`: list — много строк; один проект — одна запись/строка данных.
- **FR-009**: Все пять operations ProjectController_* (`list`, `create`, `getById`, `update`, `delete`) MUST быть достижимы через адаптерный слой и CLI F11 (`archive`/`trash` используют `update`).
- **FR-010**: Команды project MUST соблюдать контракты F06/F07: форматы вывода, отсутствие ANSI в pipe, stdout vs stderr, exit codes `0`/`1`/`2`/`3`.
- **FR-011**: `--help` для `project` и каждой подкоманды MUST документировать назначение, аргументы и флаги в scope F11.
- **FR-012**: Адаптер проектов MUST быть покрыт unit-тестами с мок-HTTP для каждой ProjectController-операции (минимум happy path) без обязательного live API в DoD.
- **FR-013**: Система MUST NOT реализовывать в F11 секции, колонки или интерактивные prompt’ы создания/подтверждения удаления (F12/F13/F26).
- **FR-014**: Вызовы API проектов MUST идти через существующий адаптер/codegen-клиент, без ручных HTTP CRUD/DTO (constitution III).
- **FR-015**: При отсутствии токена или иной ошибке конфигурации команды project MUST завершаться кодом `2` до сетевого вызова (где применимо).
- **FR-016**: Документация/help F11 MUST NOT обещать доступ к совместным (shared) проектам через list/get, если API их не отдаёт (constitution VI).

### Key Entities

- **Project**: Контейнер задач пользователя SingularityApp — идентификатор, заголовок, заметка, оформление (emoji, color), признак notebook, родительский проект (при наличии), даты архива/корзины (`journalDate` / `deleteDate`) и прочие поля ответа API, релевантные для отображения.
- **Project list query**: Набор фильтров и пагинации для выборки проектов (включение архивных/удалённых, limit/offset).
- **Project write input**: Набор полей создания/обновления из CLI (ТЗ §6.2 + `--parent`); для archive/trash — целевая дата.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Пользователь может выполнить полный цикл «создать → получить → изменить → архивировать/в корзину → удалить» по документации/`--help` с первой попытки без обращения к исходному коду.
- **SC-002**: 100% operations ProjectController_* из матрицы покрытия достижимы через поставку F11 (CLI + адаптер); секции и колонки при этом отсутствуют.
- **SC-003**: Для каждой ProjectController-операции есть автоматический unit-тест адаптера с мок-HTTP (happy path), проходящий без доступа к production API.
- **SC-004**: `singctl project --help` и help каждой из семи подкоманд перечисляют команды/флаги scope F11; ревьюер за ≤5 минут подтверждает соответствие ТЗ §6.2 + `--parent` + archive/trash (без section/column).
- **SC-005**: Типичные ошибки (нет токена, not found, misuse флага, сбой API) дают коды `2` / `3` / `1` / `1` соответственно и не смешивают текст ошибки с data-stdout (контракт F07).
- **SC-006**: Скрипт может выполнить `project list --output json` и получить массив записей без ANSI при redirect; `project get -o json` — один объект проекта (не массив), exit `0` при успехе.

## Assumptions

- Зависимости F07 (и транзитивно F01–F06: CLI-скелет, конфиг/токен, codegen, адаптер, ошибки/retry, рендер, scriptability) доступны как база для команд project; паттерны F08 (task CRUD) — ориентир UX/контрактов вывода.
- Базовая поверхность create/update — таблица ТЗ §6.2 (`title`, `note`, `isNotebook`, `emoji`, `color`) плюс `--parent` (clarification); `--note` — as-is без delta-конвертера; прочие OpenAPI write-поля (start/end, review, parentOrder, …) вне F11, кроме дат archive/trash через dedicated-команды.
- `project archive` / `project trash` — clarification сверх минимального backlog acceptance; опираются на `journalDate` / `deleteDate` в ProjectUpdateDto (как F08 для task); флаги дат на create/update в F11 не добавляются.
- Дата по умолчанию для archive/trash без `--date` — «сегодня» в согласованном формате дат проекта (как у task).
- Ответ create API может включать автоматически созданную секцию (`taskGroup`); в F11 пользователь видит только проект; CRUD секций — F12.
- Совместные проекты недоступны через REST list — это честная граница API, не баг CLI.
- Интерактивный ввод `--title` и confirm на delete отложены на F26; в F11 — только флаги/аргументы.
- `project section` / `project column` — F12/F13; alias коротких команд — F25.
- Интеграционные/live API-тесты — F33; в DoD F11 достаточно мок-unit-тестов адаптера + CLI-проверка help/поведения на моках или эквивалентном harness.
- Формат представления одного проекта и колонок списка опирается на общий рендерер F06; конкретный набор колонок table (id, title, emoji/color/notebook) выбирается разумным минимумом и может уточняться в plan без расширения scope API.
- Валидация `--color` против wiki-палитры — не обязательна в F11; неверный формат, отвергнутый API, обрабатывается как ошибка API (exit `1`).
- Конвертация `--emoji` unicode→hex — часть CLI F11 (clarification); детали эвристики hex-детекции уточняются в plan/research без расширения scope API.
