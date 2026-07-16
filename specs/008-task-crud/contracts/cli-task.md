# Contract: CLI Task Commands (F08)

**Feature**: `008-task-crud` | **Date**: 2026-07-16
**Package**: `internal/cli`

Наследует: [exit-codes-public](../../007-scriptability-exits/contracts/exit-codes-public.md), [stream-separation](../../007-scriptability-exits/contracts/stream-separation.md), [output-render](../../006-output-rendering/contracts/output-render.md), [task-output](./task-output.md).

---

## Command tree

```text
singctl task
├── list
├── get <ID>
├── create --title TITLE [flags]
├── update <ID> [flags]
├── delete <ID>
├── archive <ID> [--date DATE]
└── trash <ID> [--date DATE]
```

MUST NOT в F08: `checklist`, `kanban`, `move`, alias `t`.

---

## Flags

### list

| Flag | Maps to |
|------|---------|
| `--project` | projectId |
| `--parent` | parent |
| `--from` | startDateFrom (`YYYY-MM-DD`) |
| `--to` | startDateTo |
| `--archived` | includeArchived |
| `--removed` | includeRemoved |
| `--limit` | maxCount (1…1000) |
| `--offset` | offset (≥ 0) |
| `--all-recurrence` | includeAllRecurrence |

### create / update

| Flag | Maps to | create | update |
|------|---------|--------|--------|
| `--title` | title | required | optional |
| `--project` | projectId | optional | optional |
| `--parent` | parent | optional | optional |
| `--start` | start | optional | optional |
| `--note` | note (as-is; help: API may expect delta) | optional | optional |
| `--priority` | priority 0/1/2 | optional | optional |
| `--is-note` | isNote | optional | optional |
| `--archive-date` | journalDate | optional | optional |
| `--delete-date` | deleteDate (create: create+update) | optional | optional |

### archive / trash

| Flag | Meaning |
|------|---------|
| `--date` | `YYYY-MM-DD`; default TodayLocal |

Global: `-o/--output`, `--no-color`, `--token`, `--config`, `--debug` (existing root).

---

## Validation (before network)

| Case | Exit | stdout |
|------|------|--------|
| limit not in 1…1000 | 1 | empty |
| offset < 0 | 1 | empty |
| priority ∉ {0,1,2} | 1 | empty |
| create missing title | 1 | empty |
| update no write flags | 1 | empty |
| empty ID | 1 | empty |
| invalid date | 1 | empty |
| no token / config factory | 2 | empty |

---

## Success I/O

| Command | stdout | stderr |
|---------|--------|--------|
| list | RecordSet list (`SingleObject=false`) | empty |
| get / create / update / archive / trash | one Task (`SingleObject=true`) | empty |
| delete | **empty** | empty |

Errors: message on stderr via `Execute`; stdout empty; ExitCode per F05/F07.

---

## Help

- `singctl task --help` lists all seven subcommands.
- Each subcommand `--help` documents flags above.
- `--note` help mentions possible delta format.
- MUST NOT advertise checklist/kanban/move as available.

---

## Wiring

```text
LoadEffectiveSettings → NewFromSettings → facade → map Task(s) → output.Render
```

Reuse `Opts` format/color/date_format from root/config (F02/F06).
