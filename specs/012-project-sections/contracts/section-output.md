# Contract: Section Output (F12)

**Feature**: `012-project-sections` | **Date**: 2026-07-17
**Packages**: `internal/cli` (map Section→RecordSet), `internal/output` (Render)

Extends: [output-render.md](../../006-output-rendering/contracts/output-render.md).

---

## Columns (stable keys)

| Key | Table title |
|-----|-------------|
| id | ID |
| title | Title |
| parent | Parent |
| parentOrder | Order |
| removed | Removed? |

Одинаковый набор для list и single-section команд (значения могут быть null/empty).

---

## List vs single

| Mode | Commands | json/yaml root | table/csv |
|------|----------|----------------|-----------|
| List | `project section list` | array of objects (`[]` if empty) | header + N rows |
| Single | get, create, update | **one object** | header + 1 data row |
| None | delete success | (no render) | — |

---

## output.Render

Reuse existing `RenderOptions.SingleObject` (F08/F11). No new output package API required for F12.

| SingleObject | Rows | json/yaml |
|--------------|------|-----------|
| false | any | always array (F06) |
| true | exactly 1 | encode that map as object |
| true | 0 or >1 | error (programmer misuse) |

---

## Field display

- Числовые `parentOrder`: как number в json/yaml; table/csv — десятичная запись без лишнего шума.
- Булев `removed`: json/yaml bool; table/csv — `true`/`false` или принятый проектный формат bool (как `isNotebook` у project).
- Отсутствующее поле → json/yaml `null`, table/csv empty cell.

---

## Color / pipe

Как F06/F07: non-TTY / `--no-color` / `NO_COLOR` → без ANSI в data stdout; json/yaml/csv без ANSI всегда.
