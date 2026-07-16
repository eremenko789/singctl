# Quickstart: Task Kanban Link & Move (F10)

**Feature**: `010-task-kanban-move` | **Date**: 2026-07-17

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [api-kanban-facade.md](./contracts/api-kanban-facade.md), [cli-kanban.md](./contracts/cli-kanban.md), [kanban-output.md](./contracts/kanban-output.md). Модель: [data-model.md](./data-model.md).

---

## Prerequisites

- Go toolchain; `make test`
- F01–F08 в дереве (config, api session, task facade/`GetTask`, output SingleObject, exit codes)
- F09 checklist в дереве соседний (не блокирует F10)
- Токен в тестах только `test-token-…` (не `.env` для unit DoD)

---

## 1. Unit / harness

```bash
make test
```

Точечно:

```bash
go test ./internal/api/... ./internal/cli/... -count=1
```

**Expected** (exit 0):

| Check | Maps to |
|-------|---------|
| api List/Create/Get/Update/Delete httptest happy path | FR-010/013, SC-002/003 |
| api get/update/delete 404 → KindNotFound | FR-011, SC-005 |
| list params: only taskId/statusId when set | FR-002 |
| Move 0→create, 1→update statusId only, >1→error | FR-007, SC-003 |
| Move same column still updates | clarify Q3 |
| `task kanban --help` + 5 subcommands; `task move --help`; `task --help` mentions both | FR-012, SC-004 |
| create/move: task 404 → exit 3, no kanban write | FR-004/007, SC-005 |
| list: no task get even with `--task` | clarify Q4 |
| create always POST (no client uniqueness) | clarify Q1 |
| move rejects `--order` / no order flag | clarify Q2 |
| get/create/update/move json object; list array; delete empty stdout | FR-006/008/009 |
| no token → ExitCode 2; streams | FR-011/016 |

---

## 2. Manual smoke (optional)

```bash
go build -o /tmp/singctl ./cmd/singctl
/tmp/singctl task --help
/tmp/singctl task kanban --help
/tmp/singctl task move --help
# with token + base URL:
# /tmp/singctl task kanban list --task <TASK_ID> -o json
# /tmp/singctl task kanban create --task <TASK_ID> --column <COLUMN_ID> -o json
# /tmp/singctl task move <TASK_ID> --column <COLUMN_ID> -o json
# /tmp/singctl task kanban get <LINK_ID> -o json
# /tmp/singctl task kanban delete <LINK_ID>
```

**Expected**: help с kanban (5 команд) и move; list json — массив; get/create/move — объект; delete — пустой stdout; неизвестный task на move/create → exit 3.

---

## Out of scope check

Help и команды **не** обещают `project column`, TUI move, `--order` на move, pagination/`--removed` на kanban list.
