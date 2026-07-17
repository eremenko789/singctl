# Research: Project CRUD (F11)

**Feature**: `011-project-crud` | **Date**: 2026-07-17

Все пункты Technical Context и clarify-сессии закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где живёт project facade

**Decision**: Методы на `*api.Session` в новом `internal/api/project.go` (`ListProjects`, `GetProject`, `CreateProject`, `UpdateProject`, `DeleteProject`, плюс `ArchiveProject` / `TrashProject` как тонкие обёртки над Update). Внутри — только `Client().ProjectController*WithResponse` → `EnsureSuccess` → `Classify` (`WithEntityID` для get/update/delete/archive/trash).

**Rationale**: Constitution III/IV; зеркало F08 task facade; CLI/TUI не должны знать codegen.

**Alternatives considered**:
- Вызовы codegen из `internal/cli` — нарушение shared client.
- Отдельный пакет `internal/api/project` — лишний слой.
- Ручной HTTP — запрещён constitution III.

---

## 2. Create response: project vs taskGroup

**Decision**: `ProjectControllerCreate` возвращает `ProjectCreateResponseDto` с полями `project` и `taskGroup`. Фасад `CreateProject` MUST возвращать только смапленный `Project` из `.Project`. `taskGroup` не экспонируется CLI F11 (секции — F12).

**Rationale**: Spec FR-007 / US2; honest boundary — side-effect API не превращается в CRUD секций.

**Alternatives considered**:
- Печатать оба объекта — раздувает stdout и путает SingleObject-контракт.
- Отдельный флаг `--include-section` — scope creep F12.

---

## 3. Archive / trash (без date flags на create/update)

**Decision**:
- `ArchiveProject(id, date)` → Update с `JournalDate=date` (строка `YYYY-MM-DD`).
- `TrashProject(id, date)` → Update с `DeleteDate=date`.
- `DeleteProject(id)` → Delete; CLI при успехе **не** вызывает Render.
- Без `--date` на archive/trash: `api.TodayCalendarDate()` (локальный календарь).
- Create/update **не** принимают `--archive-date` / `--delete-date` (clarify Q3).

**Rationale**: Clarify Q1/Q3; OpenAPI `ProjectUpdateDto` имеет оба поля; симметрия с list `--archived`/`--removed`.

**Alternatives considered**:
- Только hard delete — отвергнуто clarify.
- Флаги дат на create/update как у task — отвергнуто clarify (упрощает create; нет POST+PATCH).

---

## 4. Нормализация `--emoji`

**Decision**: Чистая функция (например `NormalizeProjectEmoji(s string) (string, error)`):

1. `strings.TrimSpace`; пусто → error (usage).
2. Если вся строка матчит `^[0-9A-Fa-f]{4,8}$` → вернуть **lowercase** hex as-is (pass-through).
3. Иначе декодировать UTF-8: ровно **один** Unicode scalar (один `rune`); закодировать в lowercase hex через `strconv.FormatInt(int64(r), 16)` (пример: `💞` → `1f49e`).
4. Несколько rune / ZWJ-последовательности / смешанный «hex+символ» → error (неоднозначно).
5. ASCII-слова вроде `heart` → error (не hex и не один emoji-скаляр в смысле «один rune, но мы принимаем любой один non-empty rune кроме уже покрытого hex» — фактически шаг 3 принимает любой один rune; `h` → `68`. Чтобы не путать опечатки: **если** после trim длина в runes = 1 и это ASCII letter/digit → error; иначе один rune → hex. Уточнение:

**Уточнённая эвристика**:
- Hex pattern → pass lowercase.
- Ровно один rune **и** `r > 0x7F` (non-ASCII) → hex.
- Иначе → error.

Help: примеры `--emoji 💞` и `--emoji 1f49e`.

**Rationale**: Clarify Q4; без новых зависимостей; API wiki ожидает hex; reject ZWJ снижает сюрпризы.

**Alternatives considered**:
- As-is без конвертации — отвергнуто clarify.
- Только hex с exit 1 — отвергнуто clarify.
- `x/text` grapheme / полная emoji-lib — избыточно для F11.

---

## 5. Колонки вывода Project

**Decision**: Стабильный набор ключей для list и single-project команд:

| Key | Title (table) | Источник |
|-----|---------------|----------|
| `id` | ID | ProjectResponseDto.Id |
| `title` | Title | Title |
| `emoji` | Emoji | Emoji (hex as stored) |
| `color` | Color | Color |
| `isNotebook` | Notebook? | IsNotebook |
| `parent` | Parent | Parent |
| `journalDate` | Archived | JournalDate |
| `deleteDate` | Trash | DeleteDate |

Поле `note` (тело/delta) **не** в минимальном table/list наборе F11 (шум); json/yaml single object использует те же ключи RecordSet (не сырой DTO). Полный dump всех ResponseDto-полей — вне F11.

**Rationale**: Spec assumption «разумный минимум»; одинаковые ключи list/get для scriptability.

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
| create без `--title` (или empty trim) | 1 |
| update без ни одного write-флага | 1 |
| пустой/whitespace ID | 1 |
| невалидная дата archive/trash (`ParseDate`) | 1 |
| невалидный `--emoji` (Normalize error) | 1 |
| нет токена / factory fail | 2 |
| HTTP 404 | 3 |
| прочий API/транспорт | 1 |

Булевы `--archived`/`--removed`/`--notebook`: передавать в API только когда флаг явно задан (`Flags().Changed`), как `--is-note` у task.

`--color`: as-is string; без клиентской палитры wiki.

`--parent`: string as-is при Changed.

**Rationale**: Spec FR-002/004b/015; F07 misuse→1; F05 taxonomy.

---

## 7. Single-object json/yaml

**Decision**: Переиспользовать существующий `output.RenderOptions.SingleObject` (F08). Новых изменений `internal/output` не требуется.

**Rationale**: Уже покрыто тестами; project CLI вызывает тот же путь, что task.

---

## 8. Тестовый harness

**Decision**: Как `task_*_test.go`: httptest на `/v2/project` и `/v2/project/{id}`; temp config `BaseURL=srv.URL`, `Token=test-token-…`; `executeForTest`; assert ExitCode / streams; фасадные тесты без cobra. Create mock body: `{"project":{...},"taskGroup":{...}}` — фасад читает только project.

**Rationale**: Принятый паттерн; constitution VII.

**Alternatives considered**: Live API в DoD — F33.

---

## 9. Регистрация команд

**Decision**: `newProjectCmd()` → подкоманды list/get/create/update/delete/archive/trash; `root.AddCommand` в `newRootCmd`. Без alias. Help: Long/Example с delta для `--note`, unicode+hex для `--emoji`; MUST NOT обещать `section`/`column`/shared projects как реализованные; MAY кратко сказать «sections/columns — later» только если не выглядит как доступные подкоманды (предпочтительно просто не упоминать как команды).

**Rationale**: Spec US4 / FR-013 / FR-016.
