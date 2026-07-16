# Research: Task CRUD (F08)

**Feature**: `008-task-crud` | **Date**: 2026-07-16

Все пункты Technical Context и clarify-сессии закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где живёт task facade

**Decision**: Методы на `*api.Session` в новом `internal/api/task.go` (например `ListTasks`, `GetTask`, `CreateTask`, `UpdateTask`, `DeleteTask`, плюс `ArchiveTask` / `TrashTask` как тонкие обёртки над Update). Внутри — только `Client().TaskController*WithResponse` → `EnsureSuccess` → `Classify` (`WithEntityID` для get/update/delete/archive/trash).

**Rationale**: Constitution III/IV; F04 отложил entity facades; CLI/TUI не должны знать codegen. Один пакет с Session/Classify/retry.

**Alternatives considered**:
- Отдельный `internal/api/task` пакет — лишний слой для первой сущности.
- Вызовы codegen прямо из `internal/cli` — дублирование и нарушение shared client.
- Ручной HTTP — запрещён constitution III.

---

## 2. `--delete-date` на `task create`

**Decision**: `TaskCreateDto` **не содержит** `deleteDate` (только `journalDate`). Если пользователь передал `--delete-date` при create: (1) POST create с остальными полями; (2) при успехе PATCH update с `deleteDate`; (3) вернуть задачу после update. `--archive-date` мапится в `journalDate` на create напрямую.

**Rationale**: Honest API Boundaries (VI); ТЗ §6.1 обещает флаг на create; двухшаговый путь закрывает UX без лжи в OpenAPI.

**Alternatives considered**:
- Отклонять `--delete-date` на create — ломает ТЗ.
- Класть `deleteDate` в create body вручную — обход codegen/схемы.
- Только update-команда для корзины — хуже UX create.

---

## 3. Дата «сегодня» для archive/trash

**Decision**: При отсутствии `--date` использовать **локальный** календарный день `time.Now().In(time.Local).Format("2006-01-02")`. Явный `--date` / `--from` / `--to` / `--start` / `--archive-date` / `--delete-date` — только через `api.ParseDate` (строгий `YYYY-MM-DD`, UTC midnight для хранения строки даты как `YYYY-MM-DD` в API).

**Rationale**: «Сегодня» для пользователя — локальный календарь; API принимает date strings; ParseDate уже есть и классифицируется как KindDate → exit 1.

**Alternatives considered**:
- Всегда UTC «сегодня» — сюрприз около полуночи для пользователя.
- RFC3339 с временем — ТЗ говорит date; OpenAPI filter — date range strings.

---

## 4. Single-object json/yaml (list vs get)

**Decision**: Расширить `internal/output`: флаг вроде `RenderOptions.SingleObject bool` (или `RenderOne`). При `SingleObject=true` и ровно одной строке: json/yaml кодируют **один объект**; table/csv — одна data-row как сейчас. `list` всегда `SingleObject=false` (корень-массив, пустой → `[]`). При 0 или >1 строк + SingleObject → ошибка программиста (тесты ловят).

**Rationale**: Clarify Q4; текущий `Render` всегда `[]`. Минимальное расширение F06 без ломки list-контракта.

**Alternatives considered**:
- Всегда массив из одного элемента — отвергнуто clarify.
- Обёртка `{ "task": … }` — отвергнуто clarify.
- Дублировать encode в CLI — нарушает shared renderer.

---

## 5. Колонки вывода Task

**Decision**: Стабильный набор ключей для **одной задачи** (get/create/update/archive/trash) и для **list** (один и тот же Key set, чтобы pipe/`jq` был предсказуем):

| Key | Title (table) | Источник |
|-----|---------------|----------|
| `id` | ID | TaskResponseDto.Id |
| `title` | Title | Title |
| `projectId` | Project | ProjectId |
| `parent` | Parent | Parent |
| `priority` | Priority | Priority (0/1/2) |
| `start` | Start | Start |
| `journalDate` | Archived | JournalDate |
| `deleteDate` | Trash | DeleteDate |
| `isNote` | Note? | IsNote |

Даты из API (часто string) → в RecordSet как string as-is **или** распарсенный `time.Time` если `YYYY-MM-DD`; иначе оставить string (без падения рендера). Поле `note` (тело) **не** в минимальном наборе list/get table F08 (объём/delta); при необходимости follow-up. Полный сырой DTO в stdout не требуется.

**Rationale**: Spec assumption «разумный минимум»; одинаковые ключи list/get для scriptability; note delta шумит в table.

**Alternatives considered**:
- Все поля ResponseDto — шум для table/csv.
- Разные ключи list vs get — ломает единообразие.

---

## 6. Валидация CLI до сети

**Decision**:

| Правило | Exit |
|---------|------|
| `--limit` не в 1…1000 (если флаг задан) | 1 |
| `--offset` < 0 | 1 |
| `--priority` не 0/1/2 | 1 |
| create без `--title` | 1 |
| update без ни одного write-флага | 1 |
| пустой/whitespace ID | 1 |
| невалидная дата (`ParseDate`) | 1 (KindDate) |
| нет токена / factory fail | 2 |
| HTTP 404 | 3 |
| прочий API/транспорт | 1 |

Булевы `--archived`/`--removed`/`--all-recurrence`/`--is-note`: передавать в API только когда флаг явно задан (pflag `Changed` / tri-state), чтобы default false OpenAPI не затирался неожиданно — для bool flags cobra default false обычно ок совпадает с API default.

**Rationale**: Clarify limit; F07 misuse→1; F05 taxonomy.

---

## 7. Archive / trash / delete

**Decision**:
- `ArchiveTask(id, date)` → Update с `JournalDate=date` (строка YYYY-MM-DD).
- `TrashTask(id, date)` → Update с `DeleteDate=date`.
- `DeleteTask(id)` → Delete; CLI при успехе **не** вызывает Render (пустой stdout).

**Rationale**: coverage.md; clarify stdout.

---

## 8. Тестовый harness

**Decision**: Как `config_validate_test.go`: httptest на `/v2/task` и `/v2/task/{id}`; temp config с `BaseURL=srv.URL`, `Token=test-token-…`; `executeForTest`; assert `ExitCode`, empty stdout on error / empty stderr on success; фасадные тесты без cobra.

**Rationale**: Уже принятый паттерн F04/F05; constitution VII.

**Alternatives considered**: Live API в DoD — F33.

---

## 9. Регистрация команд

**Decision**: `newTaskCmd()` → подкоманды list/get/create/update/delete/archive/trash; `root.AddCommand` в `newRootCmd`. Без alias `t` (F25). Help: Long/Example с пометкой delta для `--note`; без упоминания checklist/move как реализованных.

**Rationale**: Spec US4 / FR-013.
