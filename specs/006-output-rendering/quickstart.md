# Quickstart: Output Rendering (F06)

**Feature**: `006-output-rendering` | **Date**: 2026-07-16

Офлайн-проверка DoD после `/speckit-implement`. Контракт: [output-render.md](./contracts/output-render.md). Модель: [data-model.md](./data-model.md).

---

## Prerequisites

- Go toolchain; `make test`
- F01/F02 available (flags/config types exist; F06 library может тестироваться изолированно)

---

## 1. Unit / harness

```bash
make test
```

Или точечно:

```bash
go test ./internal/output/... -count=1
```

**Expected** (exit 0):

| Check | Maps to |
|-------|---------|
| Fixture ≥3 rows, ≥1 date, all 4 formats agree on fields/values | SC-001, FR-001/002 |
| json/yaml root arrays; empty → `[]` | FR-001a, FR-012 |
| csv header + escaped fields | FR-011 |
| non-TTY / Color=false → no `\x1b[` in outputs | SC-002, FR-004/006 |
| `--no-color` / non-empty `NO_COLOR` path → ColorEnabled false; table without ANSI | SC-003, FR-005/005a |
| two `date_format` layouts → two literals in all formats; nil date → null/empty | SC-004, FR-007/007a/008 |
| ResolveFormat: flag > config > table | SC-005, FR-003 |

---

## 2. Manual CLI (optional, not DoD)

F06 **не** добавляет demo-команду. Ручная проверка форматов появится с list-командами F08+.

`config show -o json` проверяет **другой** рендерер (Document); не считать его доказательством RecordSet-контракта F06.

---

## Definition of done (checklist)

- [x] `internal/output` реализует ResolveFormat, ColorEnabled, FormatDate, Render
- [x] Unit harness покрывает SC-001…SC-006 без новой CLI-команды
- [x] Machine formats never contain ANSI
- [x] Dates are `date_format` strings in all formats; null/empty per clarify
- [x] `make test` green; coverage not regressing
- [x] Exported API has godoc; no package stutter
- [x] `config show` migration **not** required
