# Contract: CLI Project Commands (F11)

**Feature**: `011-project-crud` | **Date**: 2026-07-17
**Package**: `internal/cli`

Наследует: [exit-codes-public](../../007-scriptability-exits/contracts/exit-codes-public.md), [stream-separation](../../007-scriptability-exits/contracts/stream-separation.md), [output-render](../../006-output-rendering/contracts/output-render.md), [project-output](./project-output.md).

---

## Command tree

```text
singctl project
├── list
├── get <ID>
├── create --title TITLE [flags]
├── update <ID> [flags]
├── delete <ID>
├── archive <ID> [--date DATE]
└── trash <ID> [--date DATE]
```

MUST NOT в F11: `section`, `column`, alias короткого имени.

---

## Flags

### list

| Flag | Maps to |
|------|---------|
| `--archived` | includeArchived |
| `--removed` | includeRemoved |
| `--limit` | maxCount (1…1000) |
| `--offset` | offset (≥ 0) |

### create / update

| Flag | Maps to | create | update |
|------|---------|--------|--------|
| `--title` | title | required | optional |
| `--note` | note (as-is; help: API may expect delta) | optional | optional |
| `--notebook` | isNotebook | optional | optional |
| `--emoji` | emoji (after NormalizeProjectEmoji) | optional | optional |
| `--color` | color (as-is) | optional | optional |
| `--parent` | parent | optional | optional |

MUST NOT: `--archive-date`, `--delete-date` на create/update.

### archive / trash

| Flag | Meaning |
|------|---------|
| `--date` | `YYYY-MM-DD`; default TodayCalendarDate |

Global: `-o/--output`, `--no-color`, `--token`, `--config`, `--debug` (existing root).

---

## Validation (before network)

| Case | Exit | stdout |
|------|------|--------|
| limit not in 1…1000 | 1 | empty |
| offset < 0 | 1 | empty |
| create missing/empty title | 1 | empty |
| update no write flags | 1 | empty |
| empty ID | 1 | empty |
| invalid date | 1 | empty |
| invalid emoji | 1 | empty |
| no token / config factory | 2 | empty |

---

## Success I/O

| Command | stdout | stderr |
|---------|--------|--------|
| list | RecordSet list (`SingleObject=false`) | empty |
| get / create / update / archive / trash | one Project (`SingleObject=true`) | empty |
| delete | **empty** | empty |

Errors: message on stderr via `Execute`; stdout empty; ExitCode per F05/F07.

---

## Help

- `singctl project --help` lists all seven subcommands.
- Each subcommand `--help` documents flags above.
- `--note` help mentions possible delta format.
- `--emoji` help shows unicode and hex examples.
- MUST NOT advertise `section` / `column` as available commands.
- MUST NOT promise shared/collaborative projects in list results.

---

## Wiring

```text
LoadEffectiveSettings → NewFromSettings → facade → map Project(s) → output.Render
```

Reuse `Opts` format/color/date_format from root/config (F02/F06).
