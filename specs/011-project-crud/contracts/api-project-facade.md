# Contract: API Project Facade (F11)

**Feature**: `011-project-crud` | **Date**: 2026-07-17
**Package**: `internal/api` (поверх `internal/apiclient`)

Имена символов — ориентир для implement; семантика обязательна.

---

## Prerequisites

- `Session` из F04 (`NewFromSettings` / `NewSession`)
- `EnsureSuccess` + `Classify` (+ `WithEntityID`) из F04/F05
- `TodayCalendarDate` / `ParseDate` из `internal/api`
- Codegen: `ProjectController*WithResponse`, `ProjectCreateDto`, `ProjectUpdateDto`, `ProjectControllerListParams`

---

## Methods

```text
ListProjects(ctx, query ProjectListQuery) ([]Project, error)
GetProject(ctx, id string) (Project, error)
CreateProject(ctx, in ProjectWriteInput) (Project, error)
UpdateProject(ctx, id string, in ProjectWriteInput) (Project, error)
DeleteProject(ctx, id string) error
ArchiveProject(ctx, id string, dateYYYYMMDD string) (Project, error)
TrashProject(ctx, id string, dateYYYYMMDD string) (Project, error)
NormalizeProjectEmoji(raw string) (hex string, error)  // MAY live in same package
```

`Project` / `ProjectListQuery` / `ProjectWriteInput` — см. [data-model.md](../data-model.md).

---

## Mapping

| Method | Codegen call | Body / params |
|--------|--------------|---------------|
| ListProjects | `ProjectControllerListWithResponse` | `ProjectControllerListParams` from query |
| GetProject | `ProjectControllerGetByIdWithResponse` | path id |
| CreateProject | `ProjectControllerCreateWithResponse` | `ProjectCreateDto`; map **`.JSON200.Project`** only |
| UpdateProject | `ProjectControllerUpdateWithResponse` | partial `ProjectUpdateDto` (only set fields) |
| DeleteProject | `ProjectControllerDeleteWithResponse` | path id; success 204 |
| ArchiveProject | Update | `{ journalDate: date }` |
| TrashProject | Update | `{ deleteDate: date }` |

Emoji MUST быть уже нормализован до вызова Create/Update (CLI вызывает `NormalizeProjectEmoji`).

---

## Success handling

1. Вызов `*WithResponse`.
2. `EnsureSuccess(status, body)` (delete: 204 без JSON ok).
3. Decode/map `ProjectResponseDto` → `Project` (list: `projects` array; create: nested `project`).
4. Return; on failure `Classify(err, WithEntityID(id))` when id known.

---

## Errors

| Condition | Kind (via Classify) | CLI ExitCode |
|-----------|---------------------|--------------|
| HTTP 404 | NotFound | 3 |
| HTTP 401/403/422/429/5xx | per F05 catalog | 1 (429 after transport retries) |
| Transport | Transport | 1 |
| Empty id / emoji normalize fail | usage-style before network | 1 |

---

## Tests (DoD)

httptest + `test-token-…` для каждого из: list, create, get, update, delete (happy path).
Дополнительно: get или update/delete → 404 → KindNotFound.
Create: mock JSON с `project` + `taskGroup` — assert возвращён только project; taskGroup не протекает в view.
Archive/Trash: PATCH body содержит только соответствующую дату.
`NormalizeProjectEmoji`: hex pass-through, unicode→hex, reject cases.

Токен MUST NOT логироваться; Authorization как в session/task tests.
