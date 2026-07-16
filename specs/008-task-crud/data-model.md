# Data Model: Task CRUD (F08)

**Feature**: `008-task-crud` | **Date**: 2026-07-16

Логическая модель для CLI/фасада. Сериализация HTTP — codegen DTO (`TaskCreateDto`, `TaskUpdateDto`, `TaskResponseDto`); этот документ — view/query/write контракт F08.

---

## Entity: Task (view)

Представление одной задачи для stdout (после map из `TaskResponseDto`).

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| id | string | yes | Path/identity |
| title | string | yes | |
| projectId | string | no | пусто → null/empty в render |
| parent | string | no | parent task id |
| priority | number (0/1/2) | no | |
| start | date-string | no | API string / YYYY-MM-DD |
| journalDate | date-string | no | archive marker |
| deleteDate | date-string | no | trash marker |
| isNote | bool | no | |

**Relationships**: optional project (`projectId`), optional parent task (`parent`). Checklist/kanban links — out of scope F08.

**Identity**: `id` уникален в API; CLI не генерирует id.

---

## Entity: TaskListQuery

| Field | CLI flag | API param | Validation |
|-------|----------|-----------|------------|
| projectId | `--project` | projectId | optional string |
| parent | `--parent` | parent | optional string |
| startFrom | `--from` | startDateFrom | ParseDate if set |
| startTo | `--to` | startDateTo | ParseDate if set |
| includeArchived | `--archived` | includeArchived | bool |
| includeRemoved | `--removed` | includeRemoved | bool |
| maxCount | `--limit` | maxCount | if set: 1…1000 |
| offset | `--offset` | offset | if set: ≥ 0 |
| includeAllRecurrence | `--all-recurrence` | includeAllRecurrence | bool |

**Empty result**: валидный пустой список (json `[]`).

---

## Entity: TaskWriteInput

Поля create/update (частичный update — только явно заданные).

| Field | CLI flag | Create DTO | Update DTO | Notes |
|-------|----------|------------|------------|-------|
| title | `--title` | required | optional | create MUST |
| projectId | `--project` | optional | optional | clarify |
| parent | `--parent` | optional | optional | clarify |
| start | `--start` | optional | optional | ParseDate → string |
| note | `--note` | optional | optional | as-is (delta possible) |
| priority | `--priority` | optional | optional | 0/1/2 only |
| isNote | `--is-note` | optional | optional | |
| journalDate | `--archive-date` | optional | optional | ParseDate |
| deleteDate | `--delete-date` | **via follow-up Update** | optional | create: POST then PATCH |

**Update with zero fields set**: invalid (usage error, no network).

---

## Intent: Archive / Trash

| Intent | CLI | Effect |
|--------|-----|--------|
| Archive | `task archive <id> [--date]` | Update `journalDate` = date or TodayLocal |
| Trash | `task trash <id> [--date]` | Update `deleteDate` = date or TodayLocal |
| Delete | `task delete <id>` | Delete; no body; empty stdout |

**TodayLocal**: `YYYY-MM-DD` in `time.Local`.

---

## State / lifecycle (informational)

```text
active  --archive--> archived (journalDate set)
active  --trash----> trashed (deleteDate set)
*       --delete---> permanently removed (no get)
```

F08 не моделирует «un-archive» отдельной командой (можно через `update` с датами только если API это позволяет — не обещаем UX).

---

## Validation summary

- Dates: `YYYY-MM-DD` via `ParseDate`
- Priority: ∈ {0,1,2}
- Limit: 1…1000 when provided
- Offset: ≥ 0 when provided
- ID: non-empty trim
- Title on create: required non-empty (trim)
