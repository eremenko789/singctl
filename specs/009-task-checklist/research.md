# Research: Task Checklist (F09)

**Feature**: `009-task-checklist` | **Date**: 2026-07-16

Все пункты Technical Context и clarify-сессии закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где живёт checklist facade

**Decision**: Методы на `*api.Session` в новом `internal/api/checklist.go`: `ListChecklistItems`, `GetChecklistItem`, `CreateChecklistItem`, `UpdateChecklistItem`, `DeleteChecklistItem`. Внутри — только `Client().ChecklistItemController*WithResponse` → `EnsureSuccess` → `Classify` (`WithEntityID` для get/update/delete).

**Rationale**: Constitution III/IV; тот же паттерн, что F08 `task.go`; CLI/TUI не знают codegen.

**Alternatives considered**:
- Отдельный пакет `internal/api/checklist` — лишний слой.
- Вызовы codegen из CLI — нарушение shared client.
- Ручной HTTP — запрещён constitution III.

---

## 2. Pre-check parent (list / add)

**Decision**: Pre-check выполняется в **CLI** через существующий `Session.GetTask(ctx, taskID)` **до** вызова checklist facade. При `KindNotFound` → exit `3`, checklist HTTP не вызывается. Успешный GetTask **не** рендерится (stdout только checklist result). `get` / `update` / `delete` пункта — без GetTask.

**Rationale**: Clarify Q3; переиспользует F08 без дублирования task HTTP в checklist facade; facade остаётся чистым ChecklistItemController_* (проще unit-тесты фасада).

**Alternatives considered**:
- Pre-check внутри facade (`ListForTask` вызывает GetTask) — связывает сущности; усложняет изолированные тесты checklist.
- Без pre-check (только API checklist) — отвергнуто clarify.
- Отдельный HEAD/existence endpoint — его нет в API.

---

## 3. List query params

**Decision**: В `ChecklistItemControllerListParams` выставлять **только** `Parent` (= TASK_ID). `MaxCount`, `Offset`, `IncludeRemoved` **не задавать** (API defaults). CLI не экспонирует `--limit` / `--offset` / `--removed`.

**Rationale**: Clarify Q1; ТЗ §6.1; honest surface.

**Alternatives considered**:
- Прокинуть pagination «на будущее» — scope creep F09.
- Жёстко слать maxCount=1000 — неоправданное поведение без UX-флага.

---

## 4. Done / Undone flags

**Decision**:
- `add`: опциональный `--done` (bool flag); если не задан — в CreateDto `Done` omit или `false` (default API).
- `update`: `--done` → `Done: true`; `--undone` → `Done: false`; оба сразу → usage exit `1`. Частичный update: в DTO только явно заданные поля (`Title` / `Done` pointers).

**Rationale**: ТЗ `[--done]`; clarify `--undone` для снятия; pointer-поля UpdateDto позволяют partial PATCH.

**Alternatives considered**:
- `--done=true|false` один флаг — хуже discoverability в help.
- Только `--done` без undoing — нельзя снять статус из CLI.

---

## 5. Колонки вывода ChecklistItem

**Decision**: Стабильный набор ключей для list и single (get/add/update):

| Key | Title (table) | Источник |
|-----|---------------|----------|
| `id` | ID | Id |
| `title` | Title | Title |
| `done` | Done | Done (bool → true/false string or bool in json) |
| `parent` | Parent | Parent (task id) |
| `parentOrder` | Order | ParentOrder (read-only в F09) |

`crypted`, `removed`, `modificated*` — не в минимальном наборе. `parentOrder` в **выводе** допустим (API всегда отдаёт), но **write** `--order` нет (clarify Q2).

**Rationale**: Spec assumption «разумный минимум»; одинаковые ключи list/single для scriptability.

**Alternatives considered**:
- Все поля ResponseDto — шум.
- Без parentOrder в table — хуже видеть порядок; read-only безопасно.

---

## 6. Валидация CLI до сети

**Decision**:

| Правило | Exit |
|---------|------|
| add без `--title` | 1 |
| `--title` whitespace-only (trim empty) на add/update | 1 |
| update без `--title`/`--done`/`--undone` | 1 |
| `--done` и `--undone` вместе | 1 |
| пустой/whitespace TASK_ID или item ID | 1 |
| нет токена / factory fail | 2 |
| GetTask / checklist item 404 | 3 |

Порядок на add: validate args/title → open session → GetTask → CreateChecklistItem.

**Rationale**: Clarify Q4; F07; согласовано с F08 validation style.

**Alternatives considered**:
- Проброс пустого title в API — отвергнуто clarify.

---

## 7. Тестовый паттерн

**Decision**: Фасад — httptest на `/v2/checklist-item` (+ `/{id}`), как `task_test.go`. CLI list/add — httptest с **двумя** маршрутами: `GET /v2/task/{id}` (pre-check) и checklist endpoint; assert checklist не вызывается при 404 task. Help tests обновляют ожидания `task --help` (появляется `checklist`). Фикстуры: `test-token-…`.

**Rationale**: F08 harness; clarify pre-check.

**Alternatives considered**:
- Мок Session interface — больше абстракции, чем принято в проекте.

---

## 8. Регистрация CLI

**Decision**: `newTaskChecklistCmd()` с подкомандами; `task_cmd.go` добавляет её и обновляет Long/RunE (убрать «checklist недоступны»). Имена: `list`, `get`, `add`, `update`, `delete` (не `create` — как ТЗ).

**Rationale**: ТЗ §6.1; discoverability FR-011.

**Alternatives considered**:
- Плоские `task checklist-list` — хуже UX.
- Alias `create` для add — вне ТЗ; можно позже.
