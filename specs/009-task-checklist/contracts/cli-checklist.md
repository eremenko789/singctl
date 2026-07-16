# Contract: CLI Checklist Commands (F09)

**Feature**: `009-task-checklist` | **Date**: 2026-07-16
**Package**: `internal/cli`

Наследует: [exit-codes-public](../../007-scriptability-exits/contracts/exit-codes-public.md), [stream-separation](../../007-scriptability-exits/contracts/stream-separation.md), [output-render](../../006-output-rendering/contracts/output-render.md), [checklist-output](./checklist-output.md).

Depends on F08: `GetTask` via `openAPISession` / session; task group registration.

---

## Command tree

```text
singctl task
└── checklist
    ├── list <TASK_ID>
    ├── get <CHECKLIST_ITEM_ID>
    ├── add <TASK_ID> --title TITLE [--done]
    ├── update <CHECKLIST_ITEM_ID> [--title ...] [--done | --undone]
    └── delete <CHECKLIST_ITEM_ID>
```

MUST NOT в F09: `--limit`, `--offset`, `--removed`, `--order`, `--parent` на update, TUI, `kanban`, `move`.

`task --help` MUST list `checklist`. Long text F08 «checklist недоступны» MUST быть обновлён.

---

## Flags

### list

| Arg / Flag | Maps to |
|------------|---------|
| `<TASK_ID>` | list parent + pre-check GetTask |

No other list flags.

### add

| Arg / Flag | Maps to |
|------------|---------|
| `<TASK_ID>` | create.parent + pre-check GetTask |
| `--title` | title (required, non-empty trim) |
| `--done` | done=true (optional) |

### update

| Flag | Maps to |
|------|---------|
| `--title` | title (optional; if set non-empty trim) |
| `--done` | done=true |
| `--undone` | done=false |

At least one write flag required. `--done` and `--undone` mutually exclusive.

### get / delete

| Arg | Meaning |
|-----|---------|
| `<CHECKLIST_ITEM_ID>` | path id |

Global: `-o/--output`, `--no-color`, `--token`, `--config`, `--debug` (existing root).

---

## Validation (before network)

| Case | Exit | stdout |
|------|------|--------|
| missing TASK_ID / item id | 1 | empty |
| empty/whitespace id | 1 | empty |
| add missing `--title` | 1 | empty |
| empty/whitespace `--title` (add or update) | 1 | empty |
| update no write flags | 1 | empty |
| `--done` + `--undone` | 1 | empty |
| no token / config factory | 2 | empty |

На add: title validation **до** GetTask.

---

## Pre-check (list / add)

| Step | Behavior |
|------|----------|
| 1 | `GetTask(TASK_ID)` |
| 2a | NotFound → exit 3, stderr message, no checklist call, empty stdout |
| 2b | Other API error → F05/F07 (typically exit 1) |
| 3 | Success → discard task body; call checklist list/create; render checklist result only |

---

## Success stdout

| Command | stdout |
|---------|--------|
| list | RecordSet of items (F06); json/yaml array |
| get / add / update | full item; json/yaml **one object** |
| delete | empty |

---

## Error exits (network)

| Case | Exit | stderr | stdout |
|------|------|--------|--------|
| task not found (pre-check) | 3 | yes | empty |
| item not found | 3 | yes | empty |
| other API / transport | 1 | yes | empty |

---

## Help

- `task checklist --help`: five subcommands.
- Each subcommand `--help`: args/flags in scope; no TUI/kanban/order/pagination promises.
