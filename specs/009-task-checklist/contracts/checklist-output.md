# Contract: Checklist Output (F09)

**Feature**: `009-task-checklist` | **Date**: 2026-07-16
**Packages**: `internal/cli` (map ChecklistItem→RecordSet), `internal/output` (Render)

Extends: [output-render.md](../../006-output-rendering/contracts/output-render.md), F08 SingleObject behavior ([task-output.md](../../008-task-crud/contracts/task-output.md)).

---

## Columns (stable keys)

| Key | Table title |
|-----|-------------|
| id | ID |
| title | Title |
| done | Done |
| parent | Parent |
| parentOrder | Order |

Одинаковый набор для list и single-item команд (значения могут быть null/empty где применимо).

`done` в json/yaml — boolean; в table/csv — стабильное текстовое представление (`true`/`false` или `yes`/`no` — выбрать одно в implement и зафиксировать тестом; рекомендация: `true`/`false`).

---

## List vs single

| Mode | Commands | json/yaml root | table/csv |
|------|----------|----------------|-----------|
| List | `task checklist list` | array of objects (`[]` if empty) | header + N rows |
| Single | get, add, update | **one object** | header + 1 data row |
| None | delete success | (no render) | — |

---

## Render options

Reuse F08:

```text
RenderOptions.SingleObject = true  // get, add, update
RenderOptions.SingleObject = false // list
```

No new output package APIs required for F09 if SingleObject already exists.

---

## Streams

- Data → stdout only.
- Errors / usage → stderr.
- No ANSI when stdout not a TTY (F06/F07).
- Pre-check task body MUST NOT appear on stdout.
