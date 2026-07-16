# Contract: Task Output (F08)

**Feature**: `008-task-crud` | **Date**: 2026-07-16
**Packages**: `internal/cli` (map Task→RecordSet), `internal/output` (Render)

Extends: [output-render.md](../../006-output-rendering/contracts/output-render.md).

---

## Columns (stable keys)

| Key | Table title |
|-----|-------------|
| id | ID |
| title | Title |
| projectId | Project |
| parent | Parent |
| priority | Priority |
| start | Start |
| journalDate | Archived |
| deleteDate | Trash |
| isNote | Note? |

Одинаковый набор для list и single-task команд (значения могут быть null/empty).

---

## List vs single

| Mode | Commands | json/yaml root | table/csv |
|------|----------|----------------|-----------|
| List | `task list` | array of objects (`[]` if empty) | header + N rows |
| Single | get, create, update, archive, trash | **one object** | header + 1 data row |
| None | delete success | (no render) | — |

---

## output.Render extension

```text
RenderOptions.SingleObject bool  // name MAY vary (e.g. RenderOne helper)
```

| SingleObject | Rows | json/yaml |
|--------------|------|-----------|
| false | any | always array (F06) |
| true | exactly 1 | encode that map as object |
| true | 0 or >1 | error (programmer misuse) |

table/csv: поведение F06 без изменения семантики строк.

**Tests**: unit в `internal/output` — single object encode; list still array; empty list `[]`.

---

## Date display

- Использовать `opts.DateLayout` / `output.FormatDate` когда значение — `time.Time`.
- API date strings, уже `YYYY-MM-DD`, MAY проходить как string без повторного форматирования.
- Отсутствующее поле → json/yaml `null`, table/csv empty cell.

---

## Color / pipe

Как F06/F07: non-TTY / `--no-color` / `NO_COLOR` → без ANSI в data stdout; json/yaml/csv без ANSI всегда.
