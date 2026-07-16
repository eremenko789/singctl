# Quickstart: Task Checklist (F09)

**Feature**: `009-task-checklist` | **Date**: 2026-07-16

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [api-checklist-facade.md](./contracts/api-checklist-facade.md), [cli-checklist.md](./contracts/cli-checklist.md), [checklist-output.md](./contracts/checklist-output.md). Модель: [data-model.md](./data-model.md).

---

## Prerequisites

- Go toolchain; `make test`
- F01–F08 в дереве (config, api session, task facade/`GetTask`, output SingleObject, exit codes)
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
| api List/Create/Get/Update/Delete httptest happy path | FR-009/012, SC-002/003 |
| api get/update/delete 404 → KindNotFound | FR-010, SC-005 |
| list params: parent only | FR-002, clarify Q1 |
| `task checklist --help` + 5 subcommands; `task --help` mentions checklist | FR-011, SC-004 |
| list/add: task 404 → exit 3, no checklist HTTP | FR-002/004, SC-005 |
| empty title / done+undone / update no flags → exit 1 | FR-004/005 |
| get/add json shape (object); list array; delete empty stdout | FR-006/007/008 |
| no token → ExitCode 2; streams | FR-010/015 |

---

## 2. Manual smoke (optional)

```bash
go build -o /tmp/singctl ./cmd/singctl
/tmp/singctl task --help
/tmp/singctl task checklist --help
/tmp/singctl task checklist list --help
# with token + base URL:
# /tmp/singctl task checklist list <TASK_ID> -o json
# /tmp/singctl task checklist add <TASK_ID> --title "smoke" -o json
# /tmp/singctl task checklist get <ITEM_ID> -o json
# /tmp/singctl task checklist update <ITEM_ID> --done -o json
# /tmp/singctl task checklist delete <ITEM_ID>
```

**Expected**: help с пятью командами; list json — массив; get/add/update — объект; delete — пустой stdout; неизвестный task id → exit 3.

---

## Out of scope check

Help и команды **не** обещают TUI checklist, `--order`, pagination flags, kanban/`move`.
