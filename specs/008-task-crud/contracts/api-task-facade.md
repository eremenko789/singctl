# Contract: API Task Facade (F08)

**Feature**: `008-task-crud` | **Date**: 2026-07-16
**Package**: `internal/api` (поверх `internal/apiclient`)

Имена символов — ориентир для implement; семантика обязательна.

---

## Prerequisites

- `Session` из F04 (`NewFromSettings` / `NewSession`)
- `EnsureSuccess` + `Classify` (+ `WithEntityID`) из F04/F05
- Codegen: `TaskController*WithResponse`, `TaskCreateDto`, `TaskUpdateDto`, `TaskControllerListParams`

---

## Methods

```text
ListTasks(ctx, query TaskListQuery) ([]Task, error)
GetTask(ctx, id string) (Task, error)
CreateTask(ctx, in TaskWriteInput) (Task, error)
UpdateTask(ctx, id string, in TaskWriteInput) (Task, error)
DeleteTask(ctx, id string) error
ArchiveTask(ctx, id string, dateYYYYMMDD string) (Task, error)
TrashTask(ctx, id string, dateYYYYMMDD string) (Task, error)
```

`Task` / `TaskListQuery` / `TaskWriteInput` — см. [data-model.md](../data-model.md) (MAY быть локальными struct в `api` или type aliases).

---

## Mapping

| Method | Codegen call | Body / params |
|--------|--------------|---------------|
| ListTasks | `TaskControllerListWithResponse` | `TaskControllerListParams` from query |
| GetTask | `TaskControllerGetByIdWithResponse` | path id |
| CreateTask | `TaskControllerCreateWithResponse` | `TaskCreateDto` (no deleteDate) |
| CreateTask + deleteDate | затем `TaskControllerUpdateWithResponse` | `{ deleteDate }` only |
| UpdateTask | `TaskControllerUpdateWithResponse` | partial `TaskUpdateDto` (only set fields) |
| DeleteTask | `TaskControllerDeleteWithResponse` | path id; success 204 |
| ArchiveTask | Update | `{ journalDate: date }` |
| TrashTask | Update | `{ deleteDate: date }` |

---

## Success handling

1. Вызов `*WithResponse`.
2. `EnsureSuccess(status, body)` (delete: 204 без JSON ok).
3. Decode/map `TaskResponseDto` → `Task` (list: `tasks` array).
4. Return; on failure `Classify(err, WithEntityID(id))` when id known.

---

## Errors

| Condition | Kind (via Classify) | CLI ExitCode |
|-----------|---------------------|--------------|
| HTTP 404 | NotFound | 3 |
| HTTP 401/403/422/429/5xx | per F05 catalog | 1 (429 after transport retries) |
| Transport | Transport | 1 |
| Empty id before call | MAY return usage-style error without network; CLI also validates | 1 |

---

## Tests (DoD)

httptest + `test-token-…` для каждого из: list, create, get, update, delete (happy path).
Дополнительно: get или update/delete → 404 → KindNotFound.
Create with deleteDate: assert **два** запроса (POST затем PATCH) при необходимости.

Токен MUST NOT логироваться; Authorization проверяется как в session_test.
