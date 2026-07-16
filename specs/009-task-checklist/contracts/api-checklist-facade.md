# Contract: API Checklist Facade (F09)

**Feature**: `009-task-checklist` | **Date**: 2026-07-16
**Package**: `internal/api` (поверх `internal/apiclient`)

Имена символов — ориентир для implement; семантика обязательна.

---

## Prerequisites

- `Session` из F04 (`NewFromSettings` / `NewSession`)
- `EnsureSuccess` + `Classify` (+ `WithEntityID`) из F04/F05
- `GetTask` из F08 — используется **CLI** для pre-check, не обязательно внутри этих методов
- Codegen: `ChecklistItemController*WithResponse`, create/update/list DTOs, `ChecklistItemControllerListParams`

---

## Methods

```text
ListChecklistItems(ctx, query ChecklistListQuery) ([]ChecklistItem, error)
GetChecklistItem(ctx, id string) (ChecklistItem, error)
CreateChecklistItem(ctx, in ChecklistWriteInput) (ChecklistItem, error)
UpdateChecklistItem(ctx, id string, in ChecklistWriteInput) (ChecklistItem, error)
DeleteChecklistItem(ctx, id string) error
```

`ChecklistItem` / `ChecklistListQuery` / `ChecklistWriteInput` — см. [data-model.md](../data-model.md).

Фасад **не** вызывает GetTask; pre-check — ответственность CLI.

---

## Mapping

| Method | Codegen call | Body / params |
|--------|--------------|---------------|
| ListChecklistItems | `ChecklistItemControllerListWithResponse` | params: `Parent` only |
| GetChecklistItem | `ChecklistItemControllerGetByIdWithResponse` | path id |
| CreateChecklistItem | `ChecklistItemControllerCreateWithResponse` | `ChecklistItemCreateDto` (parent, title, optional done; no parentOrder/crypted) |
| UpdateChecklistItem | `ChecklistItemControllerUpdateWithResponse` | partial `ChecklistItemUpdateDto` (title and/or done only) |
| DeleteChecklistItem | `ChecklistItemControllerDeleteWithResponse` | path id; success 204 |

---

## Success handling

1. Вызов `*WithResponse`.
2. `EnsureSuccess(status, body)` (delete: 204 без JSON ok).
3. Decode/map `ChecklistItemResponseDto` → `ChecklistItem` (list: `checklistItems` array).
4. Return; on failure `Classify(err, WithEntityID(id))` when id known.

---

## Errors

| Condition | Kind (via Classify) | CLI ExitCode |
|-----------|---------------------|--------------|
| HTTP 404 | NotFound | 3 |
| HTTP 401/403/422/429/5xx | per F05 catalog | 1 |
| Transport | Transport | 1 |
| Empty id before call | MAY return usage-style error without network; CLI also validates | 1 |

---

## Tests (DoD)

httptest + `test-token-…` для каждого из: list, create, get, update, delete (happy path).
Дополнительно: get или update/delete → 404 → KindNotFound.

Assert list request query содержит `parent` и **не** требует maxCount/offset/includeRemoved от CLI.

Токен MUST NOT логироваться; Authorization как в session/task tests.
