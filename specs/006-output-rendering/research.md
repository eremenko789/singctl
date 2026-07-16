# Research: Output Rendering (F06)

**Feature**: `006-output-rendering` | **Date**: 2026-07-16

Все пункты Technical Context и clarify-сессии закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где жить shared renderer

**Decision**: Новый пакет `internal/output` (library). CLI (`internal/cli`) сохраняет `OutputFormat` как pflag-тип F01; effective resolve/render вызываются из будущего кода команд через `output`. Миграция `config show` — **не** в DoD.

**Rationale**: Constitution IV/V — presentation не дублировать в каждой команде; паттерн F05 (`internal/api` + тонкий CLI). Spec FR-009 / clarify Q5 — unit/harness only. `config show` рендерит Document (объект + masking), не RecordSet-массив — другая форма данных.

**Alternatives considered**:
- Класть рендер в `internal/cli` — раздувает CLI, хуже reuse для тестов без cobra.
- Класть в `internal/config` — смешивает IO/модель и presentation.
- Сразу рефакторить `config_show_render.go` на общий слой — желательно позже, не блокер acceptance.

---

## 2. Библиотеки форматов

**Decision**:
| Format | Library |
|--------|---------|
| table | `github.com/olekukonko/tablewriter` (уже direct; constitution) |
| json | `encoding/json` |
| yaml | `go.yaml.in/yaml/v3` (уже direct) |
| csv | `encoding/csv` |

**Rationale**: Уже в `go.mod` и частично в `config_show_render.go`; минимальный diff зависимостей.

**Alternatives considered**:
- `lipgloss` для table — TUI-стек; избыточно для CLI table DoD.
- Ручной CSV — риск сломать escaping (FR-011).

---

## 3. Модель RecordSet

**Decision**: Явные колонки + строки значений:

- `Column{Key string, Title string}` — `Key` стабилен для json/yaml/csv header/table; `Title` для table header (MAY = Key в fixture).
- Строка: `map[string]any` **или** `[]any` по индексу колонок; предпочтительно **значения по Key** + итерация колонок для порядка.
- Типы ячеек: `string`, числа, `bool`, `*time.Time` / optional date wrapper, `nil` (отсутствующее значение).
- Рендер нормализует: даты → `FormatDate`; `nil` date-field → json/yaml `null`, table/csv `""`.

**Rationale**: FR-001/001a/002; стабильный порядок колонок для CSV/table; одинаковые ключи во всех форматах.

**Alternatives considered**:
- Только `[]struct` + reflection — удобно позже, но для DoD fixture проще явная модель.
- Разные ключи table vs json — ломает «одни и те же данные».

---

## 4. ResolveFormat

**Decision**:

```text
if flagExplicitlySet → use flag value
else if configFormat non-empty & valid → use configFormat
else → table
```

Валидность enum: `table|json|yaml|csv` (как F01). Невалидный config format → fallback `table` (безопасный default; не паника).

**Rationale**: Spec FR-003 / US3. F01 уже валидирует флаг; config `Set` тоже валидирует format — fallback на всякий случай.

**Alternatives considered**:
- Всегда требовать флаг — противоречит конфигу ТЗ.
- Default yaml как у `config show` — только спецслучай show; для list/record-set default `table`.

---

## 5. Color policy

**Decision**: `ColorEnabled(isStdoutTTY, noColorFlag, noColorEnv, configColor) bool`:

1. Если `noColorFlag` **или** `noColorEnv != ""` → **false**
2. Иначе если `!isStdoutTTY` → **false**
3. Иначе если `!configColor` → **false**
4. Иначе → **true**

`isStdoutTTY` в production: `isatty` на stdout fd; в тестах — injectable bool.
`NO_COLOR`: любое **непустое** значение (clarify A); пустая строка = не задано.
`FORCE_COLOR` — вне scope.

ANSI: применять **только** в table-рендерере при `ColorEnabled==true`. JSON/YAML/CSV **никогда** не пишут escape-последовательности.

**Rationale**: Spec FR-004/005/005a/006; no-color.org; pipe-safe.

**Alternatives considered**:
- Игнорировать `NO_COLOR` — отклонено clarify.
- `FORCE_COLOR` в non-TTY — отклонено (вне scope).
- Красить json на TTY — ломает machine formats.

---

## 6. Date formatting

**Decision**: `NormalizeLayout(layout string) string` — если layout пуст или `time.Now().Format(layout)` / пробный parse reference некорректен → `2006-01-02` (config default). `FormatDate(t time.Time, layout string) string` использует нормализованный layout.

Отсутствующая дата (`nil`): не вызывать Format; сериализация per FR-007a.

**Rationale**: Clarify Q1 (строки везде) + FR-008 fallback; Go reference layout уже в ТЗ/конфиге.

**Alternatives considered**:
- ISO-8601 в json/yaml — отклонено clarify A.
- Падать на невалидном layout — отклонено FR-008.

**Note**: `api.ParseDate` — **входной** парсер `YYYY-MM-DD` (F05); display layout — зона `output.FormatDate`. Не смешивать ответственности.

---

## 7. JSON / YAML shape

**Decision**: `json.Marshal` / `yaml.Marshal` значения типа `[]map[string]any` (или эквивалент), построенного в порядке колонок. Пустой set → `[]` / `[]`.

**Rationale**: Clarify Q2; pipe `jq '.[].id'`.

**Alternatives considered**:
- Обёртка `{items:[]}` — отклонено.
- Один объект при N=1 — нестабильная форма, отклонено.

---

## 8. Table rendering

**Decision**: tablewriter с заголовками `Column.Title` (или Key); строки — уже отформатированные строковые ячейки (даты уже strings; null → ""). Цвет — только если ColorEnabled; иначе plain.

**Rationale**: Constitution tech table; единый путь stringification упрощает cross-format equality в тестах (нормализация).

**Alternatives considered**:
- Ручной `fmt` alignment — больше кода, хуже unicode width.

---

## 9. DoD / test harness

**Decision**: Пакетные `_test.go` с fixture ≥3 records, ≥1 date column, ≥1 nil date; table-driven по форматам; assert:
- decode json/yaml → same logical maps (null vs absent handled);
- csv parse → same cells;
- table contains expected literals (substring / normalized);
- bytes of json/yaml/csv never match ANSI regex;
- ColorEnabled matrix; ResolveFormat matrix; date_format change.

Нет CLI-команды, нет обязательного `config show` e2e.

**Rationale**: Clarify Q5 / FR-009 / SC-006.

**Alternatives considered**:
- Hidden `output demo` — отклонено.
- Wire через config show — другая data shape, не DoD.

---

## 10. CLI wiring in F06

**Decision**: **Не обязательна** для DoD. Опциональный follow-up (tasks MAY): helper в `cli` вызывающий `output.ResolveFormat` / `ColorEnabled` с `GlobalOptions` + loaded config — полезно F08, но SC закрываются unit-тестами `internal/output`.

**Rationale**: Spec FR-009; снижает scope creep.
