# Data Model: Project CRUD (F11)

**Feature**: `011-project-crud` | **Date**: 2026-07-17

Логическая модель для CLI/фасада. Сериализация HTTP — codegen DTO (`ProjectCreateDto`, `ProjectUpdateDto`, `ProjectResponseDto`, `ProjectCreateResponseDto`); этот документ — view/query/write контракт F11.

---

## Entity: Project (view)

Представление одного проекта для stdout (после map из `ProjectResponseDto`).

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| id | string | yes | Path/identity (часто `P-…`) |
| title | string | yes | |
| emoji | string | no | hex в API |
| color | string | no | HEX string as stored |
| isNotebook | bool | no | |
| parent | string | no | parent project id |
| journalDate | date-string | no | archive marker |
| deleteDate | date-string | no | trash marker |

**Relationships**: optional parent project (`parent`). Tasks/sections/columns — out of scope F11 (задачи — F08; секции — F12; колонки — F13).

**Identity**: `id` уникален в API; CLI не генерирует id.

---

## Entity: ProjectListQuery

| Field | CLI flag | API param | Validation |
|-------|----------|-----------|------------|
| includeArchived | `--archived` | includeArchived | bool; send if Changed |
| includeRemoved | `--removed` | includeRemoved | bool; send if Changed |
| maxCount | `--limit` | maxCount | if set: 1…1000 |
| offset | `--offset` | offset | if set: ≥ 0 |

**Empty result**: валидный пустой список (json `[]`).

**Note**: Shared projects API не возвращает — не фильтр CLI.

---

## Entity: ProjectWriteInput

Поля create/update (частичный update — только явно заданные). **Без** journalDate/deleteDate на write-флагах (только archive/trash intents).

| Field | CLI flag | Create DTO | Update DTO | Notes |
|-------|----------|------------|------------|-------|
| title | `--title` | required | optional | create MUST |
| note | `--note` | optional | optional | as-is (delta possible) |
| isNotebook | `--notebook` | optional | optional | Changed → pointer |
| emoji | `--emoji` | optional | optional | after NormalizeProjectEmoji |
| color | `--color` | optional | optional | as-is |
| parent | `--parent` | optional | optional | clarify |

**Update with zero fields set**: invalid (usage error, no network).

**Create API side-effect**: response may include `taskGroup`; view model ignores it.

---

## Intent: Archive / Trash / Delete

| Intent | CLI | Effect |
|--------|-----|--------|
| Archive | `project archive <id> [--date]` | Update `journalDate` = date or TodayCalendarDate |
| Trash | `project trash <id> [--date]` | Update `deleteDate` = date or TodayCalendarDate |
| Delete | `project delete <id>` | Delete; no body; empty stdout |

**TodayCalendarDate**: `YYYY-MM-DD` in `time.Local` (`api.TodayCalendarDate`).

---

## Emoji input (pre-API)

| Input | Result |
|-------|--------|
| `1f49e` / `1F49E` | `1f49e` |
| `💞` (U+1F49E) | `1f49e` |
| empty / multi-rune / ASCII word | usage error |

---

## State / lifecycle (informational)

```text
active  --archive--> archived (journalDate set)
active  --trash----> trashed (deleteDate set)
*       --delete---> permanently removed (no get)
```

F11 не моделирует «un-archive» отдельной командой.

---

## Validation summary

- Dates (archive/trash `--date`): `YYYY-MM-DD` via `ParseDate`
- Limit: 1…1000 when provided
- Offset: ≥ 0 when provided
- ID: non-empty trim
- Title on create: required non-empty (trim)
- Emoji: NormalizeProjectEmoji (see research §4)
- Color: no client palette check
