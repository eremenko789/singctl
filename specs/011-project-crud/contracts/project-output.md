# Contract: Project Output (F11)

**Feature**: `011-project-crud` | **Date**: 2026-07-17
**Packages**: `internal/cli` (map Project→RecordSet), `internal/output` (Render)

Extends: [output-render.md](../../006-output-rendering/contracts/output-render.md).

---

## Columns (stable keys)

| Key | Table title |
|-----|-------------|
| id | ID |
| title | Title |
| emoji | Emoji |
| color | Color |
| isNotebook | Notebook? |
| parent | Parent |
| journalDate | Archived |
| deleteDate | Trash |

Одинаковый набор для list и single-project команд (значения могут быть null/empty).

---

## List vs single

| Mode | Commands | json/yaml root | table/csv |
|------|----------|----------------|-----------|
| List | `project list` | array of objects (`[]` if empty) | header + N rows |
| Single | get, create, update, archive, trash | **one object** | header + 1 data row |
| None | delete success | (no render) | — |

---

## output.Render

Reuse existing `RenderOptions.SingleObject` (F08). No new output package API required for F11.

| SingleObject | Rows | json/yaml |
|--------------|------|-----------|
| false | any | always array (F06) |
| true | exactly 1 | encode that map as object |
| true | 0 or >1 | error (programmer misuse) |

---

## Date display

- Использовать `opts.DateLayout` / `output.FormatDate` когда значение — `time.Time`.
- API date strings `YYYY-MM-DD` MAY проходить как string.
- Отсутствующее поле → json/yaml `null`, table/csv empty cell.

---

## Color / pipe

Как F06/F07: non-TTY / `--no-color` / `NO_COLOR` → без ANSI в data stdout; json/yaml/csv без ANSI всегда.
