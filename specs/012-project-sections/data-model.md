# Data Model: Project Sections (F12)

**Feature**: `012-project-sections` | **Date**: 2026-07-17

Логическая модель для CLI/фасада. Сериализация HTTP — codegen DTO (`TaskGroupCreateDto`, `TaskGroupUpdateDto`, `TaskGroupResponseDto`, `TaskGroupListResponseDto`); этот документ — view/query/write контракт F12.

---

## Entity: Section (view)

Представление одной секции (API: task group) для stdout после map из `TaskGroupResponseDto`.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| id | string | yes | Identity (часто `Q-…`) |
| title | string | yes | |
| parent | string | no | parent **project** id (часто `P-…`) |
| parentOrder | number | no | order within parent project |
| removed | bool | no | soft-removed flag from API |

**Relationships**: принадлежит одному проекту (`parent`). Tasks inside section — out of scope F12 (задачи — F08; привязка к секции через task fields при наличии). Columns — F13.

**Identity**: `id` уникален в API; CLI не генерирует id.

**API alias**: Section ≡ TaskGroup в OpenAPI.

---

## Entity: SectionListQuery

| Field | CLI | API param | Validation |
|-------|-----|-----------|------------|
| parent | positional `<PROJECT_ID>` | parent | **required** non-empty trim |
| includeRemoved | `--removed` | includeRemoved | bool; send if Changed |
| maxCount | `--limit` | maxCount | if set: 1…1000 |
| offset | `--offset` | offset | if set: ≥ 0 |

**Empty result**: валидный пустой список (json `[]`).

**Note**: List без parent («все секции») в F12 не поддерживается.

---

## Entity: SectionWriteInput

Поля create/update (частичный update — только явно заданные).

| Field | CLI | Create | Update | Notes |
|-------|-----|--------|--------|-------|
| title | `--title` | required non-empty trim | optional; if set → non-empty trim | clarify Q3 |
| parent | create: positional `<PROJECT_ID>`; update: `--parent` | required (DTO) | optional (move) | clarify Q1; create **no** `--parent` flag |

**Out of F12 write**: `parentOrder`, `externalId`, `fake`.

**Update with zero fields set**: invalid (usage error, no network).

---

## Intent: Delete

| Intent | CLI | Effect |
|--------|-----|--------|
| Delete | `project section delete <SECTION_ID>` | Delete API; no body; empty stdout |

Нет archive/trash команд для секций (нет journalDate/deleteDate в TaskGroupUpdateDto).

---

## State / lifecycle (informational)

```text
active  --delete--> permanently removed (no get)
active  (API may mark removed) --list --removed--> visible when includeRemoved
```

F12 не моделирует soft-delete write; только list filter `--removed`.

---

## Validation summary

- Limit: 1…1000 when provided
- Offset: ≥ 0 when provided
- Project ID (list/create): non-empty trim
- Section ID (get/update/delete): non-empty trim
- Title on create: required non-empty trim
- Title on update: if flag set → non-empty trim
- Parent on update: if flag set → non-empty trim
- Update: at least one of title/parent Changed
