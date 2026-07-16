# Data Model: Task Kanban Link & Move (F10)

**Feature**: `010-task-kanban-move` | **Date**: 2026-07-17

Логическая модель для CLI/фасада. Сериализация HTTP — codegen DTO (`KanbanTaskStatusCreateDto`, `KanbanTaskStatusUpdateDto`, `KanbanTaskStatusResponseDto`); этот документ — view/query/write контракт F10.

---

## Entity: KanbanLink (view)

Представление одной связи задача↔колонка для stdout (после map из `KanbanTaskStatusResponseDto`).

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| id | string | yes | Link id (KTS-…) |
| taskId | string | yes | Task id (T-…) |
| statusId | string | yes | Column / kanban-status id (KS-…) |
| kanbanOrder | number | yes (from API) | порядок в колонке; на move write не задаём |

**Relationships**: taskId → Task (F08); statusId → KanbanStatus / column (F13 out of scope for CRUD).

**Identity**: `id` уникален в API; CLI не генерирует id.

**Out of view (F10)**: `removed`, `modificated`, `modificatedDate`, `externalId`.

---

## Entity: KanbanLinkListQuery

| Field | CLI | API param | Validation |
|-------|-----|-----------|------------|
| taskId | `--task` | taskId | optional; if set non-empty trim else exit 1 |
| statusId | `--status` | statusId | optional; if set non-empty trim else exit 1 |

**Not in F10**: maxCount, offset, includeRemoved.

**Empty result**: валидный пустой список (json `[]`) без pre-check задачи.

**Pre-check**: none on list (clarify).

---

## Entity: KanbanLinkWriteInput

Поля create/update (частичный update — только явно заданные).

| Field | CLI flag | Create DTO | Update DTO | Notes |
|-------|----------|------------|------------|-------|
| taskId | `--task` | required | optional | create: after GetTask pre-check |
| statusId | `--column` | required | optional | CLI name `--column`; API statusId |
| kanbanOrder | `--order` | optional | optional | ≥ 0 if set; omit if unset |
| externalId | — | **omit** | **omit** | out of scope |

**Create without task/column**: invalid (exit 1, no network, no pre-check).

**Create uniqueness**: not enforced client-side (clarify); always POST.

**Update with zero fields set**: invalid (usage error, no network).

**Create pre-check**: GetTask(taskId) before create; 404 → exit 3.

---

## Intent: MoveTaskToKanban

| Field | Source | Notes |
|-------|--------|-------|
| taskId | positional `<TASK_ID>` | pre-check GetTask in CLI |
| statusId | `--column` | required; no `--order` |

**Algorithm (facade)**:

1. List links with `taskId` only (`includeRemoved` unset).
2. len == 0 → Create(taskId, statusId) without kanbanOrder.
3. len == 1 → Update(link.id, statusId only) — even if statusId unchanged.
4. len > 1 → error (KindValidation / equivalent) → CLI exit 1; no write.

**Stdout**: only resulting KanbanLink after create/update; list response never printed.

---

## Intent: Delete

| Intent | CLI | Effect |
|--------|-----|--------|
| Delete | `task kanban delete <LINK_ID>` | KanbanTaskStatusController_delete; empty stdout |

No confirm / `--force` in F10.

---

## Validation summary

| Rule | When |
|------|------|
| ids non-empty (trim) | all commands with id/flags |
| `--task` + `--column` required | create |
| TASK_ID + `--column` required | move |
| at least one of task/column/order | update |
| `--order` ≥ 0 numeric | create/update if flag set |
| parent task exists | create, move (via GetTask) |
| not both ambiguous multi-link | move (facade) |

---

## Mapping notes

- List response wrapper: `kanbanTaskStatuses` → `[]KanbanLink`.
- Create required: `taskId` + `statusId`.
- Update: only non-nil pointer fields in DTO.
- Delete success: HTTP 204, no body.
- Codegen `KanbanOrder` / order fields: `float32`.
