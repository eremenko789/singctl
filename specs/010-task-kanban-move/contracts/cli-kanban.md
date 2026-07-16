# Contract: CLI Kanban & Move Commands (F10)

**Feature**: `010-task-kanban-move` | **Date**: 2026-07-17
**Package**: `internal/cli`

Наследует: [exit-codes-public](../../007-scriptability-exits/contracts/exit-codes-public.md), [stream-separation](../../007-scriptability-exits/contracts/stream-separation.md), [output-render](../../006-output-rendering/contracts/output-render.md), [kanban-output](./kanban-output.md).

Depends on F08: `GetTask` via session; task group registration.

---

## Command tree

```text
singctl task
├── kanban
│   ├── list [--task ID] [--status COLUMN_ID]
│   ├── get <LINK_ID>
│   ├── create --task ID --column COLUMN_ID [--order N]
│   ├── update <LINK_ID> [--task ...] [--column ...] [--order ...]
│   └── delete <LINK_ID>
└── move <TASK_ID> --column COLUMN_ID
```

MUST NOT в F10: `--limit`, `--offset`, `--removed` на list; `--order` на move; `project column *`; TUI move.

`task --help` MUST list `kanban` и `move`. Long text «kanban и move недоступны» MUST быть обновлён.

---

## Flags

### list

| Flag | Maps to |
|------|---------|
| `--task` | list taskId (optional) |
| `--status` | list statusId (optional) |

No pre-check. No other list flags.

### create

| Flag | Maps to |
|------|---------|
| `--task` | create.taskId + pre-check GetTask (required) |
| `--column` | create.statusId (required) |
| `--order` | create.kanbanOrder (optional, ≥ 0) |

### update

| Flag | Maps to |
|------|---------|
| `--task` | update.taskId |
| `--column` | update.statusId |
| `--order` | update.kanbanOrder (≥ 0 if set) |

At least one write flag required.

### move

| Arg / Flag | Maps to |
|------------|---------|
| `<TASK_ID>` | pre-check GetTask + MoveTaskToKanban taskId |
| `--column` | MoveTaskToKanban columnID (required) |

MUST NOT accept `--order`.

### get / delete

| Arg | Meaning |
|-----|---------|
| `<LINK_ID>` | path id |

Global: `-o/--output`, `--no-color`, `--token`, `--config`, `--debug` (existing root).

---

## Validation (before network)

| Case | Exit | stdout |
|------|------|--------|
| missing / empty ids or required flags | 1 | empty |
| create missing `--task` or `--column` | 1 | empty |
| move missing TASK_ID or `--column` | 1 | empty |
| update no write flags | 1 | empty |
| `--order` invalid / negative | 1 | empty |
| no token / config factory | 2 | empty |

На create/move: flag validation **до** GetTask.

---

## Pre-check (create / move only)

| Step | Behavior |
|------|----------|
| 1 | `GetTask(TASK_ID)` |
| 2a | NotFound → exit 3, stderr, no kanban write/move, empty stdout |
| 2b | Other API error → F05/F07 (typically exit 1) |
| 3 | Success → discard task body; call CreateKanbanLink or MoveTaskToKanban; render link only |

List: skip this table entirely.

---

## Success stdout

| Command | stdout |
|---------|--------|
| list | RecordSet of links (F06); json/yaml array |
| get / create / update / move | full link; json/yaml **one object** |
| delete | empty |

Intermediate move list MUST NOT appear on stdout.

---

## Error exits (network)

| Case | Exit | stderr | stdout |
|------|------|--------|--------|
| task not found (pre-check create/move) | 3 | yes | empty |
| link not found | 3 | yes | empty |
| move ambiguous (>1 links) | 1 | yes (hint kanban list/update) | empty |
| other API / transport | 1 | yes | empty |

---

## Help

- `task kanban --help`: five subcommands.
- `task move --help`: `--column`; no `--order`.
- Each subcommand `--help`: args/flags in scope; no TUI / column CRUD / pagination promises.
