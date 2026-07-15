# Data Model: singctl

Источник полей: `docs/api/openapi.yaml` (схемы `*ResponseDto` / `*CreateDto` / `*UpdateDto`). Ниже — доменная карта для CLI/TUI, не полный dump схемы.

## Config (локально)

| Поле | Тип | Описание |
|---|---|---|
| `api.base_url` | string | Default `https://api.singularity-app.com` |
| `api.token` | string | Bearer token без обязательного префикса `Bearer` |
| `api.timeout` | duration | HTTP timeout |
| `output.format` | enum | `table` \| `json` \| `csv` \| `yaml` |
| `output.color` | bool | |
| `output.date_format` | string | Go reference time layout |
| `tui.theme` | enum | `dark` \| `light` |
| `tui.vi_keys` | bool | |
| `tui.refresh_interval` | duration/int | `0` = без авто-refresh |

## Task

Ключевые атрибуты: `id`, `title`, `note`, `priority`, `projectId`, `parent`, `group`, `start`, `deadline`, `journalDate` (архив), `deleteDate` (корзина), `useTime`, `notifies`, `checked`/`complete`, `tags`, `modificatedDate`, `removed`.

Связи:

- Project через `projectId`
- ChecklistItem через `parent` = task id
- KanbanTaskStatus связывает task ↔ KanbanStatus
- TimeStat через `relatedTaskId`

CLI-операции archive/trash = PATCH соответствующих дат.

## Project

Атрибуты: `id`, `title`, `note`, `emoji`, `color`, `isNotebook`, `parent`, `parentOrder`, `journalDate`, `deleteDate`/`showInBasket`, `tags`, `sharedState` (read-only контекст).

Связи:

- TaskGroup (секции): `parent` = project id
- KanbanStatus: `projectId`

При create API может вернуть `ProjectCreateResponseDto` с `project` + `taskGroup`.

## TaskGroup (секция)

`id`, `title`, `parent` (project), `parentOrder`, `fake`, `removed`.

## KanbanStatus (колонка)

`id`, `name`, `projectId`, `kanbanOrder`, `numberOfColumns`, `removed`.

## KanbanTaskStatus (позиция задачи)

`id`, `taskId`, `statusId`, `kanbanOrder`. Создание/обновление = «переместить задачу».

## ChecklistItem

`id`, `parent` (task), `title`, `done`, `parentOrder`, `crypted`, `removed`.

## Habit

`id`, `title`, `description`, `color`, `order`, `status`, `externalId`, `removed`.

## HabitDailyProgress

`id`, `habit`, `date`, `progress`, `externalId`, `removed`.

Значения `progress` (из ТЗ/wiki): `0` нейтрально, `1` не выполнено, `2` выполнено — подтвердить при интеграционных тестах.

## Tag

`id`, `title`, `parent`, `parentOrder`, `hotkey`, `color`, `removed`. Иерархия через `parent`.

## TimeStat

`id`, `start`, `end`, `secondsPassed`, `quantity`, `relatedTaskId`, `source`, `outdated`, `removed`.

Bulk delete — отдельная операция list-level DELETE с фильтрами (см. OpenAPI), CLI может принимать список ID как UX над этой операцией.

## Validation rules (клиент)

- Обязательный `title` при create task/project/habit/tag/checklist (где требует API).
- Даты: ISO / конфиг `date_format`; ошибка формата до HTTP.
- Токен непустой перед сетевыми вызовами.
- Не отправлять поля recurrence на create, если API их не принимает для создания.

## Mapping CLI resource → OpenAPI path

| CLI | Paths |
|---|---|
| `task` | `/v2/task` |
| `task checklist` | `/v2/checklist-item` |
| `task move` | `/v2/kanban-task-status` |
| `project` | `/v2/project` |
| `project column` | `/v2/kanban-status` |
| `project section` | `/v2/task-group` |
| `habit` | `/v2/habit` |
| `habit track` | `/v2/habit-progress` |
| `tag` | `/v2/tag` |
| `time` | `/v2/time-stat` |
