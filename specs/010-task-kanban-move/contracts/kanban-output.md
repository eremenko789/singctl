# Contract: Kanban Link Output (F10)

**Feature**: `010-task-kanban-move` | **Date**: 2026-07-17
**Packages**: `internal/cli` (map KanbanLink→RecordSet), `internal/output` (Render)

Extends: [output-render.md](../../006-output-rendering/contracts/output-render.md), F08 SingleObject behavior ([task-output.md](../../008-task-crud/contracts/task-output.md)).

---

## Columns (stable keys)

| Key | Table title |
|-----|-------------|
| id | ID |
| taskId | Task |
| statusId | Column |
| kanbanOrder | Order |

Одинаковый набор для list и single-item команд (значения могут быть null/empty где применимо).

`kanbanOrder` в json/yaml — number; в table/csv — стабильное текстовое представление числа.

---

## List vs single

| Mode | Commands | json/yaml root | table/csv |
|------|----------|----------------|-----------|
| List | `task kanban list` | array of objects (`[]` if empty) | header + N rows |
| Single | get, create, update, move | **one object** | header + 1 data row |
| None | delete success | (no render) | — |

---

## Render options

Reuse F08:

```text
RenderOptions.SingleObject = true  // get, create, update, move
RenderOptions.SingleObject = false // list
```

No new output package APIs required for F10 if SingleObject already exists.

---

## Streams

- Data → stdout only.
- Errors / usage → stderr.
- No ANSI when stdout not a TTY (F06/F07).
- Pre-check task body MUST NOT appear on stdout.
- Move intermediate list MUST NOT appear on stdout.
