# Contract: Output Render Library (F06)

**Feature**: `006-output-rendering` | **Package**: `internal/output`
**Date**: 2026-07-16

Контракт library API для unit/harness DoD. Имена функций — целевые; допускается эквивалентная группировка при сохранении семантики.

---

## 1. ResolveFormat

```text
ResolveFormat(flagSet bool, flagValue string, configFormat string) Format
```

| Case | Result |
|------|--------|
| `flagSet == true`, valid flag | = flag |
| `flagSet == true`, invalid flag | не ожидается на пути F01 (парсер уже отклонил); если вызвано — error или fallback table (implement: prefer error for programmer misuse in tests) |
| `flagSet == false`, valid configFormat | = configFormat |
| `flagSet == false`, empty/invalid config | `table` |

**Tests**: три сценария SC-005 (flag > config > default).

---

## 2. ColorEnabled

```text
ColorEnabled(isTTY bool, noColorFlag bool, noColorEnv string, configColor bool) bool
```

| Inputs | Result |
|--------|--------|
| noColorFlag=true | false |
| noColorEnv non-empty | false |
| isTTY=false | false |
| configColor=false | false |
| else | true |

**Env contract**: caller передаёт `os.Getenv("NO_COLOR")` (или test double); пустая строка = unset.

**Tests**: matrix covering FR-004/005/005a; ANSI absence when false.

---

## 3. FormatDate / NormalizeLayout

```text
NormalizeLayout(layout string) string   // empty/invalid → "2006-01-02"
FormatDate(t time.Time, layout string) string  // uses NormalizeLayout
```

| Input | Output |
|-------|--------|
| valid layout | `t.Format(layout)` |
| empty / invalid | `t.Format("2006-01-02")` |
| absent date (nil) | not formatted here — handled in Render |

**Tests**: two layouts → two literals (SC-004); invalid → default.

---

## 4. Render

```text
Render(w io.Writer, set RecordSet, opts RenderOptions) error
```

`opts.Format` ∈ table|json|yaml|csv.

### 4.1 Cross-format invariants

- len(records in output) == len(set.Rows)
- Field keys == Column.Key set
- Non-null dates: identical formatted strings across formats
- Null dates: json/yaml `null`; table/csv empty cell
- No ANSI in json/yaml/csv bytes ever
- No ANSI in table when `opts.Color == false`
- Empty Rows: json `[]`; yaml empty sequence; csv header only; table without data rows

### 4.2 JSON

- Root: JSON array
- Element: object with keys = Column.Key (all columns present; null for absent dates)
- Dates: JSON strings (formatted), not ISO auto by encoder
- Numbers/bools: JSON-native allowed

### 4.3 YAML

- Root: sequence of mappings
- Same logical values as JSON (null / strings / etc.)

### 4.4 CSV

- First row: Column.Key (or Title — **Decision**: use **Key** for machine stability; Title only for table)
- Subsequent rows: string cells in column order
- Use `encoding/csv` Writer (CRLF/LF per Writer; tests accept either or set Comma/UseCRLF explicitly)

### 4.5 Table

- Headers: Column.Title if non-empty else Key
- Cells: stringified; dates formatted; nil → empty
- Color/ANSI only if `opts.Color`

### Errors

- Unknown format → error (non-empty message); nothing useful on Writer or best-effort clear
- Writer failures → return error
- Invalid RecordSet (duplicate keys) → error preferred at Render or constructor

**Stderr**: library does not print; callers write errors to stderr. Tests use `bytes.Buffer` as `w`.

---

## 5. ANSI detection (tests)

Treat as ANSI if output matches common ESC sequences, e.g. contains `\x1b[` (CSI). Machine formats and color-off table must not match.

---

## 6. Non-goals (this contract)

- Cobra command surface / `singctl …` e2e
- Migrating `config show` Document renderer
- Entity column schemas
- `FORCE_COLOR`
- Exit codes (F07)
