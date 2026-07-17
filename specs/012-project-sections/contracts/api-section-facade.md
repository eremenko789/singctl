# Contract: API Section Facade (F12)

**Feature**: `012-project-sections` | **Date**: 2026-07-17
**Package**: `internal/api` (поверх `internal/apiclient`)

Имена символов — ориентир для implement; семантика обязательна.
CLI-термин: **section**. Codegen: **TaskGroupController_***.

---

## Prerequisites

- `Session` из F04 (`NewFromSettings` / `NewSession`)
- `EnsureSuccess` + `Classify` (+ `WithEntityID`) из F04/F05
- Codegen: `TaskGroupController*WithResponse`, `TaskGroupCreateDto`, `TaskGroupUpdateDto`, `TaskGroupControllerListParams`
- Depends F11 project facade (не обязателен для вызовов section, но CLI группа `project` уже существует)

---

## Methods

```text
ListSections(ctx, query SectionListQuery) ([]Section, error)
GetSection(ctx, id string) (Section, error)
CreateSection(ctx, in SectionWriteInput) (Section, error)
UpdateSection(ctx, id string, in SectionWriteInput) (Section, error)
DeleteSection(ctx, id string) error
```

`Section` / `SectionListQuery` / `SectionWriteInput` — см. [data-model.md](../data-model.md).

Create `SectionWriteInput` MUST включать non-empty `Parent` и `Title`.
Update MAY включать `Title` и/или `Parent` (partial).

---

## Mapping

| Method | Codegen call | Body / params |
|--------|--------------|---------------|
| ListSections | `TaskGroupControllerListWithResponse` | `TaskGroupControllerListParams` from query (`Parent` required) |
| GetSection | `TaskGroupControllerGetByIdWithResponse` | path id |
| CreateSection | `TaskGroupControllerCreateWithResponse` | `TaskGroupCreateDto` (`Title`, `Parent`) |
| UpdateSection | `TaskGroupControllerUpdateWithResponse` | partial `TaskGroupUpdateDto` (only set fields) |
| DeleteSection | `TaskGroupControllerDeleteWithResponse` | path id; success 204 |

List response: map `JSON200.TaskGroups` → `[]Section`.
Create/Get/Update: map `JSON200` (`TaskGroupResponseDto`) → `Section`.

---

## Success handling

1. Вызов `*WithResponse`.
2. `EnsureSuccess(status, body)` (delete: 204 без JSON ok).
3. Decode/map DTO → `Section` / `[]Section`.
4. Return; on failure `Classify(err, WithEntityID(id))` when id known.

---

## Errors

| Condition | Kind (via Classify) | CLI ExitCode |
|-----------|---------------------|--------------|
| HTTP 404 | NotFound | 3 |
| HTTP 401/403/422/429/5xx | per F05 catalog | 1 (429 after transport retries) |
| Transport | Transport | 1 |
| Empty id / empty required parent/title | usage-style before network | 1 |

---

## Tests (DoD)

httptest + `test-token-…` для каждого из: list, create, get, update, delete (happy path).
Дополнительно: get или update/delete → 404 → KindNotFound.
List: assert query содержит `parent`; mock body `{"taskGroups":[...]}`.
Create: assert body `title` + `parent`.
Update: partial body (title only / parent only / both).

Токен MUST NOT логироваться; Authorization как в session/project tests.
