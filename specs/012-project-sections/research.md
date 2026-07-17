# Research: Project Sections (F12)

**Feature**: `012-project-sections` | **Date**: 2026-07-17

Все пункты Technical Context и clarify-сессии закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где живёт section facade

**Decision**: Методы на `*api.Session` в новом `internal/api/section.go`:
`ListSections`, `GetSection`, `CreateSection`, `UpdateSection`, `DeleteSection`.
Внутри — только `Client().TaskGroupController*WithResponse` → `EnsureSuccess` → `Classify` (`WithEntityID` для get/update/delete).

**Rationale**: Constitution III/IV; зеркало F08/F11; CLI/TUI не должны знать codegen; UX-имя Section (FR-016).

**Alternatives considered**:
- Имена `ListTaskGroups` в публичном фасаде — путает CLI/TUI с API jargon.
- Вызовы codegen из `internal/cli` — нарушение shared client.
- Ручной HTTP — запрещён constitution III.

---

## 2. Section vs task-group naming

**Decision**:
- CLI / help / user docs: **section** (`project section …`).
- OpenAPI / codegen / HTTP path: **task-group** (`/v2/task-group`, TaskGroupController_*).
- Domain view type в `internal/api`: `Section` (map из `TaskGroupResponseDto`).
- Contracts: явно таблица Section ↔ TaskGroup.

**Rationale**: Spec FR-016; honest boundary без переименования API.

**Alternatives considered**:
- Команды `project task-group` — расходится с ТЗ §6.2.
- Alias обоих — scope F25.

---

## 3. List: required parent project

**Decision**: CLI `project section list <PROJECT_ID>` всегда передаёт `parent=<PROJECT_ID>` в `TaskGroupControllerListParams`. Без аргумента — usage exit `1` до сети. OpenAPI допускает list без `parent`; F12 **не** экспонирует «все секции аккаунта» (clarify Q2).

Фасад `ListSections(ctx, query)` требует non-empty `query.Parent` (или CLI гарантирует до вызова — предпочтительно валидация на CLI + assert в фасаде при пустом parent → usage-style error).

**Rationale**: Clarify Q2; ТЗ §6.2; снижает риск огромных списков.

**Alternatives considered**:
- Опциональный project / `--project` — отвергнуто clarify.
- List без parent в фасаде «на будущее» без CLI — не нужно в F12.

---

## 4. Write surface (create / update)

**Decision**:

| Surface | Create | Update |
|---------|--------|--------|
| title | required non-empty (trim) | optional; if set → non-empty trim |
| parent | from positional `<PROJECT_ID>` only | optional `--parent` (move) |
| parentOrder / externalId / fake | out of scope | out of scope |

Update with zero changing flags → exit `1`. Update with only `--parent` (non-empty) → valid. Create MUST NOT expose `--parent` flag (FR-005a).

Create DTO: `Title` + `Parent` required by OpenAPI. Update DTO: partial fields only when Changed.

**Rationale**: Clarify Q1/Q3; OpenAPI TaskGroupCreateDto/UpdateDto; симметрия с project `--parent` на update.

**Alternatives considered**:
- `--order` / full write dump — отвергнуто clarify.
- `--parent` на create как флаг — отвергнуто (позиционный ID яснее и совпадает с ТЗ).

---

## 5. Колонки вывода Section

**Decision**: Стабильный набор ключей для list и single-section команд:

| Key | Title (table) | Источник |
|-----|---------------|----------|
| `id` | ID | TaskGroupResponseDto.Id |
| `title` | Title | Title |
| `parent` | Parent | Parent |
| `parentOrder` | Order | ParentOrder |
| `removed` | Removed? | Removed |

Поля `fake`, `externalId`, `modificatedDate`, `modificated` **не** в минимальном table/list наборе F12 (шум / служебные). json/yaml single object использует те же ключи RecordSet (не сырой DTO).

**Rationale**: Spec assumption «разумный минимум»; `parentOrder` полезен для ориентира даже без write `--order`; одинаковые ключи list/get.

**Alternatives considered**:
- Все поля ResponseDto — шум для table/csv.
- Без parentOrder — хуже для сортировки глазами в table.

---

## 6. Валидация CLI до сети

**Decision**:

| Правило | Exit |
|---------|------|
| list без PROJECT_ID / empty trim | 1 |
| `--limit` не в 1…1000 (если флаг задан) | 1 |
| `--offset` < 0 | 1 |
| create без `--title` или empty trim | 1 |
| create без / empty PROJECT_ID | 1 |
| update без ни одного write-флага | 1 |
| update `--title` empty trim (если флаг задан) | 1 |
| update `--parent` empty trim (если флаг задан) | 1 |
| пустой/whitespace section ID (get/update/delete) | 1 |
| нет токена / factory fail | 2 |
| HTTP 404 | 3 |
| прочий API/транспорт | 1 |

Булев `--removed`: передавать в API только когда флаг явно задан (`Flags().Changed`), как `--archived` у project.

**Rationale**: Spec FR-002/004/005/015; F07 misuse→1; F05 taxonomy; clarify Q3.

---

## 7. Response unwrap

**Decision**:
- List: `JSON200.TaskGroups` → `[]Section`.
- Get / Create / Update: `JSON200` как `TaskGroupResponseDto` → `Section` (create **не** вложен в wrapper, в отличие от ProjectCreateResponseDto).
- Delete: 204, без body; CLI не Render.

**Rationale**: OpenAPI schemas; отличие от F11 create unwrap.

**Alternatives considered**: N/A — формат API фиксирован.

---

## 8. Single-object json/yaml

**Decision**: Переиспользовать `output.RenderOptions.SingleObject` (F08/F11). Новых изменений `internal/output` не требуется.

**Rationale**: Уже покрыто тестами; section CLI вызывает тот же путь.

---

## 9. Тестовый harness

**Decision**: Как `project_*_test.go` / `task_checklist_*_test.go`: httptest на `/v2/task-group` и `/v2/task-group/{id}`; temp config `BaseURL=srv.URL`, `Token=test-token-…`; `executeForTest`; assert ExitCode / streams; фасадные тесты без cobra. List mock: `{"taskGroups":[...]}`.

**Rationale**: Принятый паттерн; constitution VII.

**Alternatives considered**: Live API в DoD — F33.

---

## 10. Регистрация команд

**Decision**: `newProjectSectionCmd()` с подкомандами list/get/create/update/delete; `newProjectCmd()` добавляет `section`. Обновить `project` Long/RunE hint: упомянуть `section` среди подкоманд. Без alias. Help MUST NOT обещать `column` как реализованный; MAY не упоминать column вовсе.

**Rationale**: Spec US4 / FR-011 / FR-013; паттерн `task checklist`.
