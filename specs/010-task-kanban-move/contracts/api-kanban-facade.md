# Contract: API Kanban Task-Status Facade (F10)

**Feature**: `010-task-kanban-move` | **Date**: 2026-07-17
**Package**: `internal/api` (поверх `internal/apiclient`)

Имена символов — ориентир для implement; семантика обязательна.

---

## Prerequisites

- `Session` из F04 (`NewFromSettings` / `NewSession`)
- `EnsureSuccess` + `Classify` (+ `WithEntityID`) из F04/F05
- `GetTask` из F08 — используется **CLI** для pre-check create/move, не внутри CRUD-методов ниже
- Codegen: `KanbanTaskStatusController*WithResponse`, create/update/list DTOs, `KanbanTaskStatusControllerListParams`

---

## Methods

```text
ListKanbanLinks(ctx, query KanbanLinkListQuery) ([]KanbanLink, error)
GetKanbanLink(ctx, id string) (KanbanLink, error)
CreateKanbanLink(ctx, in KanbanLinkWriteInput) (KanbanLink, error)
UpdateKanbanLink(ctx, id string, in KanbanLinkWriteInput) (KanbanLink, error)
DeleteKanbanLink(ctx, id string) error
MoveTaskToKanban(ctx, taskID, columnID string) (KanbanLink, error)
```

Типы — см. [data-model.md](../data-model.md).

CRUD-методы **не** вызывают GetTask. `MoveTaskToKanban` **не** вызывает GetTask (pre-check — CLI) и **не** задаёт kanbanOrder.

---

## Mapping

| Method | Codegen call | Body / params |
|--------|--------------|---------------|
| ListKanbanLinks | `KanbanTaskStatusControllerListWithResponse` | params: optional TaskId, StatusId only |
| GetKanbanLink | `KanbanTaskStatusControllerGetByIdWithResponse` | path id |
| CreateKanbanLink | `KanbanTaskStatusControllerCreateWithResponse` | CreateDto: taskId, statusId, optional kanbanOrder; no externalId |
| UpdateKanbanLink | `KanbanTaskStatusControllerUpdateWithResponse` | partial UpdateDto (taskId and/or statusId and/or kanbanOrder) |
| DeleteKanbanLink | `KanbanTaskStatusControllerDeleteWithResponse` | path id; success 204 |
| MoveTaskToKanban | List (TaskId) then Create or Update | update: StatusId only; create: taskId+statusId, no order |

---

## MoveTaskToKanban semantics

1. `ListKanbanLinks` with `taskId` only.
2. `len == 0` → `CreateKanbanLink` (taskId, statusId).
3. `len == 1` → `UpdateKanbanLink(id, statusId=columnID)` always (even if equal).
4. `len > 1` → return classified error (`KindValidation` or project-equivalent) with message that CLI prints to stderr; no write.

List response from step 1 MUST NOT leak to CLI stdout (CLI renders only returned KanbanLink).

---

## Success handling

1. Вызов `*WithResponse`.
2. `EnsureSuccess(status, body)` (delete: 204 без JSON ok).
3. Decode/map `KanbanTaskStatusResponseDto` → `KanbanLink` (list: `kanbanTaskStatuses`).
4. Return; on failure `Classify(err, WithEntityID(id))` when id known.

---

## Errors

| Condition | Kind (via Classify) | CLI ExitCode |
|-----------|---------------------|--------------|
| HTTP 404 | NotFound | 3 |
| HTTP 401/403/422/429/5xx | per F05 catalog | 1 |
| Transport | Transport | 1 |
| Move ambiguous (>1 links) | Validation (or Other with clear message) | 1 |
| Empty id before call | MAY return usage-style error without network; CLI also validates | 1 |

---

## Tests (DoD)

httptest + `test-token-…` для каждого из: list, create, get, update, delete (happy path).

Дополнительно:

- get или update/delete → 404 → KindNotFound
- list request: при фильтрах только taskId/statusId; **не** maxCount/offset/includeRemoved от фасада F10
- Move: 0 links → POST create; 1 link → PATCH statusId only (assert no kanbanOrder in body); >1 → error, no write
- Move same statusId still PATCHes

Токен MUST NOT логироваться; Authorization как в session/task tests.
