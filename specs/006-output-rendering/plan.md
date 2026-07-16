# Implementation Plan: Output Rendering (F06)

**Branch**: `006-output-rendering` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-output-rendering/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F06. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F06 вводит общий presentation-слой **`internal/output`** для логического набора записей (record set):

- рендер в `table|json|yaml|csv` с одинаковым составом полей/значений;
- `json`/`yaml` — **корневой массив** объектов; даты — строки по `output.date_format` (не ISO-only); отсутствующая дата → `null` / пустая ячейка;
- политика цвета: stdout non-TTY, `--no-color`, непустой `NO_COLOR`, `output.color`; ANSI только в `table` (MAY);
- резолв формата: явный флаг → `output.format` → `table`.

DoD — **только** unit/harness на fixture (без новой CLI-команды и без обязательной миграции `config show`). Entity-колонки — F08+.

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: stdlib (`encoding/json`, `encoding/csv`, `io`, `os`, `time`); уже в модуле — `github.com/olekukonko/tablewriter`, `go.yaml.in/yaml/v3`; TTY — `github.com/mattn/go-isatty` (уже indirect, при необходимости поднять в direct). Без новых тяжёлых UI-библиотек.

**Storage**: N/A (только in-memory record set + конфиг-поля F02 как вход резолва).

**Testing**: TDD — failing unit-тесты render/format/color/date → реализация; fixture ≥3 записей с датой; симуляция non-TTY / `NO_COLOR` без реальной CLI-команды. `make test` + coverage; фиктивные данные без секретов.

**Target Platform**: CLI `singctl` (macOS/Linux); library используется будущими list-командами.

**Project Type**: shared library package внутри Go monorepo (+ опциональная тонкая CLI-проводка резолва, не обязательная для DoD).

**Performance Goals**: рендер типичных list-размеров (сотни/тысячи строк) без заметной задержки в интерактиве; unit-suite быстрый.

**Constraints**: constitution V/IX; FR-009 — без demo CLI; не ломать валидацию `--output` F01; `config show` ad-hoc MAY остаться; FORCE_COLOR вне scope.

**Scale/Scope**: новый `internal/output` + тесты; минимальная/нулевая правка `internal/cli` для DoD (резолв можно тестировать чисто в `output`); entity commands — нет.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go в том же модуле |
| G3 | OpenAPI-Generated API Client | PASS N/A | HTTP/codegen не затрагивается |
| G4 | Shared Client for CLI and TUI | PASS | presentation в `internal/output` (reuse list-командами; TUI не дублирует date/format policy) |
| G5 | Scriptability First | PASS | table/json/yaml/csv; pipe без ANSI; stdout/stderr разделение |
| G6 | Honest API Boundaries | PASS | без новых API-обещаний |
| G7 | Security of Credentials | PASS | fixture без токенов; рендерер не логирует секреты |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты до/вместе с кодом; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/006-output-rendering/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── output-render.md
└── tasks.md             # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/output/
├── doc.go                 # NEW: package overview
├── format.go              # NEW: Format enum + ResolveFormat(flagChanged, flag, config)
├── color.go               # NEW: ColorEnabled(tty, noColorFlag, noColorEnv, configColor)
├── date.go                # NEW: FormatDate / NormalizeDateLayout (default 2006-01-02)
├── model.go               # NEW: Column, RecordSet, cell value helpers (incl. optional dates)
├── render.go              # NEW: Render(w, format, set, opts) dispatcher
├── render_json.go         # NEW: root array; dates as formatted strings; null dates
├── render_yaml.go         # NEW: root list; same value rules as JSON
├── render_csv.go          # NEW: header + encoding/csv escaping
├── render_table.go        # NEW: tablewriter; color only if opts.Color
├── format_test.go         # NEW
├── color_test.go          # NEW
├── date_test.go           # NEW
├── render_test.go         # NEW: cross-format fixture harness (SC-001…SC-006)
└── …

internal/cli/
├── output.go              # EXISTS: OutputFormat pflag — KEEP; MAY thin-wrap Resolve later (not DoD)
└── config_show_render.go  # EXISTS ad-hoc — NO required migration in F06

internal/config/
└── types.go               # EXISTS: OutputConfig Format/Color/DateFormat — consume as inputs
```

**Structure Decision**: Shared record-set renderer живёт в **`internal/output`** (аналог `internal/api` для F05): CLI не раздувается presentation-логикой; будущие entity list-команды импортируют пакет. `config show` остаётся отдельным ad-hoc путём до follow-up. DoD закрывается unit-тестами пакета без новой команды.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
- пакет `internal/output` vs расширение `cli`;
- tablewriter + stdlib json/csv + yaml.v3;
- ColorEnabled: isatty(stdout) + `--no-color` + `NO_COLOR` + `output.color`;
- даты: `time.Time.Format(layout)` с валидацией layout / fallback default;
- JSON/YAML как `[]map` с порядком колонок из Column defs;
- DoD harness без CLI.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — RecordSet, Column, Format, ColorPolicy, DateFormatSetting
- `contracts/output-render.md` — Render / ResolveFormat / ColorEnabled / FormatDate
- `quickstart.md` — `make test` scenarios mapping to SC-*

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. Format + ResolveFormat tests → implementation;
2. ColorEnabled tests (TTY/`NO_COLOR`/`--no-color`/config) → implementation;
3. FormatDate / invalid layout fallback tests → implementation;
4. Cross-format Render fixture tests (incl. empty, null dates, CSV escape, no ANSI in machine formats) → renderers;
5. Table color on/off tests;
6. godoc; `make test` coverage.
