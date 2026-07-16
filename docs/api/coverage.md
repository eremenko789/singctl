# Покрытие SingularityApp REST API v2

Источник: `docs/api/openapi.yaml` (снимок с `https://api.singularity-app.com/v2/api-json`).

Требование constitution: клиент и CLI/TUI MUST закрывать **все** операции ниже.
Это входной документ для `/speckit.specify` / plan — не замена артефактов Spec Kit.

**Всего операций: 51.** Схемы в `components.schemas` без путей (например `AuthenticateRequest`) в публичном REST не экспонируются и в CLI не нужны.

| Method | Path | operationId | CLI (целевой) | Примечание |
|---|---|---|---|---|
| GET | `/v2/task` | TaskController_list | `singctl task list` | фильтры OpenAPI |
| POST | `/v2/task` | TaskController_create | `singctl task create` | |
| GET | `/v2/task/{id}` | TaskController_getById | `singctl task get` | |
| PATCH | `/v2/task/{id}` | TaskController_update | `singctl task update` / `archive` / `trash` | archive/trash = PATCH дат |
| DELETE | `/v2/task/{id}` | TaskController_delete | `singctl task delete` | |
| GET | `/v2/checklist-item` | ChecklistItemController_list | `singctl task checklist list` | `parent` = task id |
| POST | `/v2/checklist-item` | ChecklistItemController_create | `singctl task checklist add` | |
| GET | `/v2/checklist-item/{id}` | ChecklistItemController_getById | `singctl task checklist get` | |
| PATCH | `/v2/checklist-item/{id}` | ChecklistItemController_update | `singctl task checklist update` | |
| DELETE | `/v2/checklist-item/{id}` | ChecklistItemController_delete | `singctl task checklist delete` | |
| GET | `/v2/project` | ProjectController_list | `singctl project list` | |
| POST | `/v2/project` | ProjectController_create | `singctl project create` | |
| GET | `/v2/project/{id}` | ProjectController_getById | `singctl project get` | |
| PATCH | `/v2/project/{id}` | ProjectController_update | `singctl project update` | |
| DELETE | `/v2/project/{id}` | ProjectController_delete | `singctl project delete` | |
| GET | `/v2/task-group` | TaskGroupController_list | `singctl project section list` | секции |
| POST | `/v2/task-group` | TaskGroupController_create | `singctl project section create` | |
| GET | `/v2/task-group/{id}` | TaskGroupController_getById | `singctl project section get` | |
| PATCH | `/v2/task-group/{id}` | TaskGroupController_update | `singctl project section update` | |
| DELETE | `/v2/task-group/{id}` | TaskGroupController_delete | `singctl project section delete` | |
| GET | `/v2/kanban-status` | KanbanStatusController_list | `singctl project column list` | колонки |
| POST | `/v2/kanban-status` | KanbanStatusController_create | `singctl project column create` | |
| GET | `/v2/kanban-status/{id}` | KanbanStatusController_getById | `singctl project column get` | |
| PATCH | `/v2/kanban-status/{id}` | KanbanStatusController_update | `singctl project column update` | |
| DELETE | `/v2/kanban-status/{id}` | KanbanStatusController_delete | `singctl project column delete` | |
| GET | `/v2/kanban-task-status` | KanbanTaskStatusController_list | `singctl task kanban list` | связи task↔column |
| POST | `/v2/kanban-task-status` | KanbanTaskStatusController_create | `singctl task move` / `kanban create` | |
| GET | `/v2/kanban-task-status/{id}` | KanbanTaskStatusController_getById | `singctl task kanban get` | |
| PATCH | `/v2/kanban-task-status/{id}` | KanbanTaskStatusController_update | `singctl task kanban update` / `move` | |
| DELETE | `/v2/kanban-task-status/{id}` | KanbanTaskStatusController_delete | `singctl task kanban delete` | |
| GET | `/v2/habit` | HabitController_list | `singctl habit list` | |
| POST | `/v2/habit` | HabitController_create | `singctl habit create` | |
| GET | `/v2/habit/{id}` | HabitController_getById | `singctl habit get` | |
| PATCH | `/v2/habit/{id}` | HabitController_update | `singctl habit update` | |
| DELETE | `/v2/habit/{id}` | HabitController_delete | `singctl habit delete` | |
| GET | `/v2/habit-progress` | HabitDailyProgressController_list | `singctl habit progress list` | |
| POST | `/v2/habit-progress` | HabitDailyProgressController_create | `singctl habit track` / `progress create` | |
| GET | `/v2/habit-progress/{id}` | HabitDailyProgressController_getById | `singctl habit progress get` | |
| PATCH | `/v2/habit-progress/{id}` | HabitDailyProgressController_update | `singctl habit progress update` | |
| DELETE | `/v2/habit-progress/{id}` | HabitDailyProgressController_delete | `singctl habit progress delete` | |
| GET | `/v2/tag` | TagController_list | `singctl tag list` | |
| POST | `/v2/tag` | TagController_create | `singctl tag create` | |
| GET | `/v2/tag/{id}` | TagController_getById | `singctl tag get` | |
| PATCH | `/v2/tag/{id}` | TagController_update | `singctl tag update` | |
| DELETE | `/v2/tag/{id}` | TagController_delete | `singctl tag delete` | |
| GET | `/v2/time-stat` | TimeStatController_list | `singctl time list` | |
| POST | `/v2/time-stat` | TimeStatController_create | `singctl time add` | |
| GET | `/v2/time-stat/{id}` | TimeStatController_getById | `singctl time get` | |
| PATCH | `/v2/time-stat/{id}` | TimeStatController_update | `singctl time update` | |
| DELETE | `/v2/time-stat/{id}` | TimeStatController_delete | `singctl time delete` | |
| DELETE | `/v2/time-stat` | TimeStatController_deleteBulk | `singctl time delete-bulk` | фильтры date/task, не отдельный path `/delete-bulk` |

## Пробелы исходного ТЗ (закрыты этой матрицей)

Черновик `docs/tz/…` описывал `habit track`, но не полный CRUD `/v2/habit-progress`.
Также не были явно названы `get` для checklist / column / section / kanban-link и полный CRUD `kanban-task-status` (кроме move). Матрица выше — канон покрытия.

## Как обновлять

```bash
make openapi-fetch
make api-coverage-check   # сверяет число operations в openapi.json с ожиданием
```

После изменения upstream OpenAPI обновите эту таблицу и прогоните Spec Kit clarify/plan при необходимости.
