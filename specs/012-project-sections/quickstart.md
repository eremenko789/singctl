# Quickstart: Project Sections (F12)

**Feature**: `012-project-sections` | **Date**: 2026-07-17

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [api-section-facade.md](./contracts/api-section-facade.md), [cli-section.md](./contracts/cli-section.md), [section-output.md](./contracts/section-output.md). Модель: [data-model.md](./data-model.md). Research: [research.md](./research.md).

---

## Prerequisites

- Go toolchain; `make test`
- F01–F11 в дереве (config, api session, output SingleObject, exit codes, `project` group)
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
| list params include required `parent` | FR-002 |
| api get/update/delete 404 → KindNotFound | FR-010, SC-005 |
| `project section --help` + 5 subcommands / `--parent` on update | FR-011, SC-004 |
| list without PROJECT_ID → ExitCode 1, no network | FR-002, clarify |
| empty/whitespace `--title` → ExitCode 1 | FR-004/005, clarify |
| get/create/update json shape; delete empty stdout | FR-007/008 |
| update `--parent` only → PATCH parent | FR-005, clarify |
| no token → ExitCode 2; not found → 3; misuse → 1; streams | FR-010/015, SC-005 |

---

## 2. Manual smoke (optional)

```bash
go build -o /tmp/singctl ./cmd/singctl
/tmp/singctl project --help
/tmp/singctl project section --help
/tmp/singctl project section list --help
/tmp/singctl project section create --help
/tmp/singctl project section update --help
# with token + base URL:
# /tmp/singctl project section list <PROJECT_ID> -o json
# /tmp/singctl project section create <PROJECT_ID> --title "smoke" -o json
# /tmp/singctl project section get <SECTION_ID> -o json
# /tmp/singctl project section update <SECTION_ID> --title "renamed" -o json
# /tmp/singctl project section update <SECTION_ID> --parent <OTHER_PROJECT_ID> -o json
# /tmp/singctl project section delete <SECTION_ID>
```

**Expected**: `section` виден в `project --help`; list json — массив; get/create/update — объект; delete — пустой stdout, exit 0; нет `column` в help tree F12.

---

## 3. Out of scope checks

- Нет `singctl project column` в help tree F12.
- Нет `--order` / archive/trash для section.
- Нет live coverage gate (F35) и integration suite (F33) как блокеров F12.
