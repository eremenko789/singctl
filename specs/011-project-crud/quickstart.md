# Quickstart: Project CRUD (F11)

**Feature**: `011-project-crud` | **Date**: 2026-07-17

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [api-project-facade.md](./contracts/api-project-facade.md), [cli-project.md](./contracts/cli-project.md), [project-output.md](./contracts/project-output.md). Модель: [data-model.md](./data-model.md). Research: [research.md](./research.md).

---

## Prerequisites

- Go toolchain; `make test`
- F01–F08 в дереве (config, api session, output SingleObject, exit codes, task pattern)
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
| NormalizeProjectEmoji: hex / unicode / reject | FR-004b |
| api List/Create/Get/Update/Delete httptest happy path | FR-009/012, SC-002/003 |
| create unwraps `project`, ignores `taskGroup` | FR-007 |
| api get/update/delete 404 → KindNotFound | FR-010, SC-005 |
| Archive/Trash PATCH journalDate/deleteDate | FR-006/006a |
| `project --help` + subcommand help: 7 команд / ключевые флаги | FR-011, SC-004 |
| list filters / limit validation без сети | FR-002 |
| get/create json shape; delete empty stdout | FR-007/008 |
| no token → ExitCode 2; not found → 3; misuse → 1; streams | FR-010/015, SC-005 |

---

## 2. Manual smoke (optional)

```bash
go build -o /tmp/singctl ./cmd/singctl
/tmp/singctl project --help
/tmp/singctl project list --help
/tmp/singctl project create --help
# with token + base URL:
# /tmp/singctl project list -o json
# /tmp/singctl project create --title "smoke" --emoji 💞 -o json
# /tmp/singctl project get <ID> -o json
# /tmp/singctl project archive <ID> -o json
# /tmp/singctl project delete <ID>
```

**Expected**: help без section/column; list json — массив; get/create — объект; emoji в API как hex; delete — пустой stdout, exit 0.

---

## 3. Out of scope checks

- Нет `singctl project section` / `column` в help tree F11.
- Нет live coverage gate (F35) и integration suite (F33) как блокеров F11.
