# Contract: CLI Project Section Commands (F12)

**Feature**: `012-project-sections` | **Date**: 2026-07-17
**Package**: `internal/cli`

Наследует: [exit-codes-public](../../007-scriptability-exits/contracts/exit-codes-public.md), [stream-separation](../../007-scriptability-exits/contracts/stream-separation.md), [output-render](../../006-output-rendering/contracts/output-render.md), [section-output](./section-output.md).

---

## Command tree

```text
singctl project
└── section
    ├── list <PROJECT_ID>
    ├── get <SECTION_ID>
    ├── create <PROJECT_ID> --title TITLE
    ├── update <SECTION_ID> [--title TITLE] [--parent PROJECT_ID]
    └── delete <SECTION_ID>
```

MUST NOT в F12: `project column`, archive/trash для section, `--order`, alias короткого имени.

`singctl project --help` MUST показывать подкоманду `section`.

---

## Flags

### list

| Arg / Flag | Maps to |
|------------|---------|
| `<PROJECT_ID>` (required) | parent |
| `--removed` | includeRemoved |
| `--limit` | maxCount (1…1000) |
| `--offset` | offset (≥ 0) |

### create

| Arg / Flag | Maps to |
|------------|---------|
| `<PROJECT_ID>` (required) | parent (DTO) |
| `--title` (required, non-empty trim) | title |

MUST NOT: `--parent` flag on create.

### update

| Flag | Maps to |
|------|---------|
| `--title` | title (non-empty if set) |
| `--parent` | parent (move; non-empty if set) |

At least one of `--title` / `--parent` MUST be Changed.

### delete / get

Positional `<SECTION_ID>` only.

Global: `-o/--output`, `--no-color`, `--token`, `--config`, `--debug` (existing root).

---

## Validation (before network)

| Case | Exit | stdout |
|------|------|--------|
| list missing/empty PROJECT_ID | 1 | empty |
| limit not in 1…1000 | 1 | empty |
| offset < 0 | 1 | empty |
| create missing/empty title or PROJECT_ID | 1 | empty |
| update no write flags | 1 | empty |
| update empty title or parent when flag set | 1 | empty |
| empty SECTION_ID | 1 | empty |
| no token / config factory | 2 | empty |

---

## Success I/O

| Command | stdout | stderr |
|---------|--------|--------|
| list | RecordSet list (`SingleObject=false`) | empty |
| get / create / update | one Section (`SingleObject=true`) | empty |
| delete | **empty** | empty |

Errors: message on stderr via `Execute`; stdout empty; ExitCode per F05/F07.

---

## Help

- `singctl project section --help` lists all five subcommands.
- Each subcommand `--help` documents args/flags above.
- `--parent` on update: описать как перенос секции в другой проект.
- MUST use user term «секция» / section; MAY mention API task-group only in Long technical note if needed (prefer not).
- MUST NOT advertise `column` as available command.

---

## Wiring

```text
LoadEffectiveSettings → NewFromSettings → ListSections/GetSection/… → map Section(s) → output.Render
```

Reuse `Opts` format/color/date_format from root/config (F02/F06).
Паттерн файлов: `project_section_*.go` (как `task_checklist_*.go`).
