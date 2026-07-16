# Research: Task Kanban Link & Move (F10)

**Feature**: `010-task-kanban-move` | **Date**: 2026-07-17

Все пункты Technical Context и clarify-сессии закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где живёт kanban facade и move

**Decision**: Методы на `*api.Session` в `internal/api/kanban_task_status.go`:

- `ListKanbanLinks`, `GetKanbanLink`, `CreateKanbanLink`, `UpdateKanbanLink`, `DeleteKanbanLink`
- `MoveTaskToKanban(ctx, taskID, columnID)` — оркестрация list → create | update statusId | ambiguous error

Внутри CRUD — только `Client().KanbanTaskStatusController*WithResponse` → `EnsureSuccess` → `Classify` (`WithEntityID` для get/update/delete).

**Rationale**: Constitution III/IV; тот же паттерн, что F08/F09; логика move переиспользуема будущим TUI (хоткей `m`) без дублирования в cobra.

**Alternatives considered**:
- Move только в CLI (list+create/update руками) — дублирование для TUI (нарушение IV spirit).
- Отдельный пакет `internal/api/kanban` — лишний слой.
- Ручной HTTP — запрещён constitution III.

---

## 2. Pre-check задачи (create / move only)

**Decision**: Pre-check в **CLI** через `Session.GetTask(ctx, taskID)` **до** `CreateKanbanLink` / `MoveTaskToKanban`. При `KindNotFound` → exit `3`, kanban write/move HTTP не вызывается (для move — и list внутри Move не должен стартовать). Успешный GetTask **не** рендерится.

`task kanban list` — **без** GetTask (clarify Q4), даже с `--task`.

`get` / `update` / `delete` связи — без GetTask.

**Rationale**: Clarify Q4; согласовано с F09 (pre-check на write/parent-required paths); list фильтр опционален → пустой список при «левом» taskId — валидный UX.

**Alternatives considered**:
- Pre-check на list при `--task` — отвергнуто clarify.
- Pre-check колонки через KanbanStatus get — F13 out of scope; полагаемся на API.

---

## 3. List query params

**Decision**: В `KanbanTaskStatusControllerListParams` выставлять только заданные фильтры:

| CLI | API |
|-----|-----|
| `--task` | `TaskId` |
| `--status` | `StatusId` |

`MaxCount`, `Offset`, `IncludeRemoved` **не задавать**. CLI не экспонирует `--limit` / `--offset` / `--removed`.

Для `MoveTaskToKanban` внутренний list: только `TaskId`; `IncludeRemoved` unset (default false).

**Rationale**: Spec FR-002; ТЗ §6.1; honest surface.

**Alternatives considered**:
- Прокинуть pagination «на будущее» — scope creep F10.
- Жёстко maxCount=1000 на move — скрытое поведение; при >1 связи и так exit 1; при огромном числе связей без фильтра — edge API, не CLI-флаг.

---

## 4. Create vs uniqueness vs move

**Decision**:

- `CreateKanbanLink` / `kanban create`: всегда POST create после pre-check задачи; **нет** list-проверки «уже есть связь» (clarify Q1).
- `MoveTaskToKanban`: list по taskId → len 0 create; len 1 update только `StatusId`; len >1 → ошибка с `KindValidation` (или эквивалент, мапящийся в exit `1`) и сообщением указать `task kanban list` / `update`; **без** create/update.

**Rationale**: Clarify Q1 + FR-007; явный CRUD отдельно от UX move.

**Alternatives considered**:
- Create отказывает при ≥1 связи — отвергнуто clarify.
- Create upsert как move — смешивает поверхности; отвергнуто.

---

## 5. Move: same column и `--order`

**Decision**:

- При ровно одной связи: **всегда** PATCH update с `StatusId = columnID`, даже если уже совпадает (clarify Q3). Не short-circuit, не пустой stdout.
- `MoveTaskToKanban` / `task move` **не** принимают / не передают `KanbanOrder` (clarify Q2). Create-ветка: omit order (API default). Update-ветка: только `StatusId` pointer, `KanbanOrder` omit.

**Rationale**: Проще тесты и TUI; ТЗ §6.1 для move без `--order`.

**Alternatives considered**:
- No-op get при same column — отвергнуто clarify.
- `--order` на move — отвергнуто clarify.

---

## 6. Колонки вывода KanbanLink

**Decision**: Стабильный набор ключей для list и single (get/create/update/move):

| Key | Title (table) | Источник |
|-----|---------------|----------|
| `id` | ID | Id |
| `taskId` | Task | TaskId |
| `statusId` | Column | StatusId |
| `kanbanOrder` | Order | KanbanOrder |

`removed`, `modificated*`, `modificatedDate`, `externalId` — не в минимальном наборе F10.

**Rationale**: Spec assumption; scriptability; одинаковые ключи list/single.

**Alternatives considered**:
- Ключ `column` вместо `statusId` — расхождение с API/json shape; лучше статусный id как в API, table title «Column».
- Все поля ResponseDto — шум.

---

## 7. Валидация CLI до сети

**Decision**:

| Правило | Exit |
|---------|------|
| create без `--task` или `--column` | 1 |
| move без TASK_ID или `--column` | 1 |
| пустой/whitespace task/column/link id | 1 |
| update без `--task`/`--column`/`--order` | 1 |
| `--order` отрицательный или не число | 1 |
| нет токена / factory fail | 2 |
| GetTask 404 (create/move) | 3 |
| link 404 (get/update/delete) | 3 |
| move >1 связей | 1 |

Порядок на create: validate flags → open session → GetTask → CreateKanbanLink.
Порядок на move: validate → session → GetTask → MoveTaskToKanban.

`--order` на create/update: CLI `float32`/`float64` ≥ 0 (codegen `*float32`); передавать pointer только если флаг задан.

**Rationale**: Spec edge cases + F07; согласовано с F08/F09.

**Alternatives considered**:
- Integer-only order — OpenAPI number/float32; принимать неотрицательное число.

---

## 8. Тестовый паттерн

**Decision**:

- Фасад CRUD: httptest `/v2/kanban-task-status` (+ `/{id}`), как checklist/task tests.
- `MoveTaskToKanban`: httptest с list + create или update; кейсы 0 / 1 / >1; assert update body содержит statusId и **не** требует kanbanOrder.
- CLI create/move: httptest с `GET /v2/task/{id}` + kanban endpoints; assert kanban не вызывается при 404 task.
- CLI list: один kanban list endpoint; assert **нет** вызова task get.
- Help: `task --help` упоминает `kanban` и `move`; убрать текст «kanban и move недоступны».
- Фикстуры: `test-token-…`.

**Rationale**: F08/F09 harness; clarify pre-check / list.

**Alternatives considered**:
- Мок Session interface — больше абстракции, чем принято.

---

## 9. Регистрация CLI

**Decision**:

- `newTaskKanbanCmd()` с подкомандами `list` / `get` / `create` / `update` / `delete` (имена как ТЗ, не `add`).
- `newTaskMoveCmd()` — сосед `task move` (не под `kanban`).
- `task_cmd.go`: AddCommand обеих; обновить Long/RunE.

**Rationale**: ТЗ §6.1; coverage map (`move` / `kanban create`).

**Alternatives considered**:
- `task kanban move` вместо `task move` — расходится с ТЗ.
- Alias `add` для create — вне ТЗ.
