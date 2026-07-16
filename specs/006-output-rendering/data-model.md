# Data Model: Output Rendering (F06)

**Feature**: `006-output-rendering` | **Date**: 2026-07-16

Логические сущности presentation-слоя. Не схема БД.

---

## Format

Способ представления record set.

| Value | Notes |
|-------|-------|
| `table` | Human-readable; MAY use ANSI if ColorEnabled |
| `json` | Root **array** of objects; no ANSI |
| `yaml` | Root **sequence** of mappings; no ANSI |
| `csv` | Header row + data rows; RFC escaping; no ANSI |

**Default** (после resolve): `table`.

---

## Column

Определение поля record set.

| Field | Type | Rules |
|-------|------|-------|
| Key | string | Non-empty; стабильный идентификатор во всех форматах (json/yaml keys, csv header, logical field) |
| Title | string | Заголовок для table; MAY равняться Key |
| Kind | enum (optional) | `string` / `number` / `bool` / `date` — влияет на null/date formatting; если Kind=date и value nil → null/empty rules |

**Identity**: Key уникален в рамках одного RecordSet.

---

## CellValue

Значение ячейки до сериализации.

| Representation | Meaning |
|----------------|---------|
| `nil` / typed absent | Отсутствует; для Kind=date → json/yaml `null`, table/csv `""` |
| `time.Time` / `*time.Time` | Дата/время; сериализуется через DateFormatSetting → **string** во всех форматах |
| `string`, numeric, `bool` | Как есть (stringified for table/csv) |

После нормализации для текстовых форматов все ячейки становятся строками (кроме json/yaml typed null/bool/number — числа/bool MAY остаться JSON-native; даты всегда string или null).

**Recommendation for F06 fixture**: даты как `*time.Time`; прочие поля строки/числа; один nil date.

---

## RecordSet

Набор однотипных записей.

| Field | Type | Rules |
|-------|------|-------|
| Columns | []Column | Ordered; определяет порядок csv/table и набор ключей |
| Rows | []Row | Row = map[Key]CellValue (или эквивалент); отсутствующий ключ трактовать как nil |

**Invariants**:
- Число логических записей = len(Rows) во всех форматах.
- Пустой Rows → валидный empty output (json `[]`, yaml `[]`, csv headers only, empty/header-only table).
- Не добавлять data-rows при empty.

**Relationships**: RecordSet *uses* Columns; Render *consumes* RecordSet + Format + RenderOptions.

---

## DateFormatSetting

| Field | Type | Rules |
|-------|------|-------|
| Layout | string | Go reference layout; empty/invalid → `2006-01-02` (`config.DefaultOutputDateFormat`) |

**Behavior**: `FormatDate(t, Layout)` → string; одинаковая строка во всех форматах для одного `t`.

---

## ColorPolicy (effective)

Входы → булев «цвет разрешён».

| Input | Effect |
|-------|--------|
| `--no-color` true | Off |
| `NO_COLOR` non-empty | Off |
| stdout not TTY | Off |
| `output.color` false | Off |
| else (TTY + color allowed) | On |

Используется только table-рендерером. Machine formats игнорируют On.

---

## RenderOptions

| Field | Type | Source |
|-------|------|--------|
| Format | Format | ResolveFormat |
| Color | bool | ColorPolicy |
| DateLayout | string | NormalizeLayout(config.Output.DateFormat) |
| Writer | io.Writer | обычно stdout buffer в тестах / os.Stdout в командах |

---

## ResolveFormat inputs

| Input | Priority |
|-------|----------|
| Explicit CLI `--output`/`-o` (flag changed) | 1 |
| `config.Output.Format` if valid | 2 |
| Default `table` | 3 |

---

## Validation rules (summary)

| ID | Rule |
|----|------|
| VR-001 | Format ∈ {table,json,yaml,csv} |
| VR-002 | Column.Key unique & non-empty |
| VR-003 | Invalid/empty date layout → default `2006-01-02` |
| VR-004 | nil date → null (json/yaml) / empty (table/csv) |
| VR-005 | CSV must escape commas/quotes/newlines |
| VR-006 | json/yaml root is array; length = len(Rows) |

---

## Out of model (deferred)

- Entity-specific column catalogs (task/project/…) — F08+
- Single-object “show card” JSON shape — F08+
- `config show` Document flattening — follow-up migration
- Exit codes — F07
