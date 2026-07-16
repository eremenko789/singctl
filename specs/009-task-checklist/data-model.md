# Data Model: Task Checklist (F09)

**Feature**: `009-task-checklist` | **Date**: 2026-07-16

Логическая модель для CLI/фасада. Сериализация HTTP — codegen DTO (`ChecklistItemCreateDto`, `ChecklistItemUpdateDto`, `ChecklistItemResponseDto`); этот документ — view/query/write контракт F09.

---

## Entity: ChecklistItem (view)

Представление одного пункта для stdout (после map из `ChecklistItemResponseDto`).

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| id | string | yes | Path/identity |
| title | string | yes | |
| done | bool | yes | completion status |
| parent | string | yes | parent **task** id |
| parentOrder | number | no | read-only в F09 write; показывать в output |

**Relationships**: parent = task id (F08 Task). Reparent / non-task parent — out of scope F09.

**Identity**: `id` уникален в API; CLI не генерирует id.

**Out of view (F09)**: `crypted`, `removed`, `modificated`, `modificatedDate`.

---

## Entity: ChecklistListQuery

| Field | CLI | API param | Validation |
|-------|-----|-----------|------------|
| parent | positional `<TASK_ID>` | parent | required non-empty (trim); else exit 1 |

**Not in F09**: maxCount, offset, includeRemoved.

**Empty result**: валидный пустой список (json `[]`) **после** успешного pre-check задачи.

**Pre-check**: CLI вызывает GetTask(parent) before list; 404 → exit 3, no list call.

---

## Entity: ChecklistWriteInput

Поля create/update (частичный update — только явно заданные).

| Field | CLI flag | Create DTO | Update DTO | Notes |
|-------|----------|------------|------------|-------|
| parent | positional TASK_ID (add only) | required | **not set** | update MUST NOT change parent |
| title | `--title` | required (non-empty trim) | optional (if set: non-empty trim) | |
| done | `--done` / `--undone` | optional `--done` only | `--done`→true / `--undone`→false | mutually exclusive on update |
| parentOrder | — | **omit** | **omit** | clarify: no `--order` |
| crypted | — | omit | omit | out of scope |

**Add without title / empty title**: invalid (exit 1, no network, no pre-check).

**Update with zero fields set**: invalid (usage error, no network).

**Add pre-check**: GetTask(TASK_ID) before create; 404 → exit 3.

---

## Intent: Delete

| Intent | CLI | Effect |
|--------|-----|--------|
| Delete | `task checklist delete <id>` | ChecklistItemController_delete; empty stdout |

No confirm / `--force` in F09.

---

## Validation summary

| Rule | When |
|------|------|
| TASK_ID / item id non-empty (trim) | all commands with id |
| title non-empty (trim) | add always; update if `--title` set |
| at least one of title/done/undone | update |
| not both done and undone | update |
| parent task exists | list, add (via GetTask) |

---

## Mapping notes

- List response wrapper: `checklistItems` → `[]ChecklistItem`.
- Create required: `parent` + `title`.
- Update: only non-nil pointer fields in DTO.
- Delete success: HTTP 204, no body.
