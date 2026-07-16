# Quickstart: Task CRUD (F08)

**Feature**: `008-task-crud` | **Date**: 2026-07-16

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [api-task-facade.md](./contracts/api-task-facade.md), [cli-task.md](./contracts/cli-task.md), [task-output.md](./contracts/task-output.md). Модель: [data-model.md](./data-model.md).

---

## Prerequisites

- Go toolchain; `make test`
- F01–F07 в дереве (config, api session, output, exit codes)
- Токен в тестах только `test-token-…` (не `.env` для unit DoD)

---

## 1. Unit / harness

```bash
make test
```

Точечно:

```bash
go test ./internal/output/... ./internal/api/... ./internal/cli/... -count=1
```

**Expected** (exit 0):

| Check | Maps to |
|-------|---------|
| output SingleObject: json object vs list array | FR-008b, SC-006 |
| api List/Create/Get/Update/Delete httptest happy path | FR-009/012, SC-002/003 |
| api get/update/delete 404 → KindNotFound | FR-010, SC-005 |
| create + deleteDate → POST then PATCH (если реализовано в тесте) | research §2 |
| `task --help` + subcommand help содержат 7 команд / ключевые флаги | FR-011, SC-004 |
| list filters / limit validation без сети | FR-002 |
| get/create json shape; delete empty stdout | FR-008/008a/008b |
| no token → ExitCode 2; not found → 3; misuse → 1; streams | FR-010/015, SC-005 |

---

## 2. Manual smoke (optional, local mock or real API)

С мок-сервером или реальным API (не обязателен для merge DoD):

```bash
go build -o /tmp/singctl ./cmd/singctl
/tmp/singctl task --help
/tmp/singctl task list --help
# with token + base URL configured:
# /tmp/singctl task list -o json
# /tmp/singctl task create --title "smoke" -o json
# /tmp/singctl task get <ID> -o json
# /tmp/singctl task delete <ID>
```

**Expected**: help без checklist/move; list json — массив; get/create — объект; delete — пустой stdout, exit 0.

---

## 3. Out of scope checks

- Нет `singctl task checklist` / `kanban` / `move` в help tree F08.
- Нет live coverage gate (F35) и integration suite (F33) как блокеров F08.
