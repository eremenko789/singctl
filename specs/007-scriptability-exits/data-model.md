# Data Model: Scriptability & Exit Codes (F07)

**Feature**: `007-scriptability-exits` | **Date**: 2026-07-16

Логические сущности контракта (не persistence). Числовой mapping реализуется существующим `cli.ExitCode` (F05).

---

## Exit Code Contract

Публичная семантика завершения процесса CLI.

| Code | Meaning | Typical signals |
|------|---------|-----------------|
| `0` | Успех | `err == nil` |
| `1` | Ошибка API / операции / транспорта / **использования CLI** / прочее | Classified API kinds (не NotFound/Config); transport; date validation; unknown flag/command; invalid `--output` |
| `2` | Ошибка конфигурации | Kind Config (нет токена, factory config) |
| `3` | Не найдено | Kind NotFound (HTTP 404 classified) |

**Rules**:
- Единственный SoT в runtime: `cli.ExitCode(err)` (F05).
- Misuse MUST NOT → `2` или `3`.
- Unknown/unclassified → `1`, never `0`.

**Relationships**: ClassifiedError.Kind (F05) → ExitCode; docs/help **document** the same table.

---

## Stream Separation Policy

Правила размещения байтов для DoD-команд (нестримящих).

| Outcome | stdout | stderr |
|---------|--------|--------|
| Success (normal, non-debug) | Полезные данные или success message | Empty (trim) |
| Failure | Empty (trim) | User-facing error (MAY have `Ошибка:` prefix) |
| `--help` / `--version` success | Help/version text | Empty |

**Out of scope**: verbose/debug diagnostics; future streaming partial writes.

**Relationships**: Applies to `config show`, `config validate`, misuse paths in F07 DoD; future entity commands MUST follow same policy (F08+).

---

## Pipe Scenario Contract

Формализация примера из ТЗ §10.

| Field | Description |
|-------|-------------|
| ID | Стабильный id (`json-redirect`, `list-jq-xargs`, `csv-awk`, `xargs-create`) |
| TZ example | Цитата/краткий bash из §10 |
| Required properties | format, no-ANSI on non-TTY, stream rules, exit semantics |
| Status | `verifiable_now` \| `contract_f08_plus` |
| Verified by | F06 fixture / F07 CLI tests / docs-only |

**Relationships**: Depends on Output Format + Color Policy (F06) and Exit Code Contract (F05/F07).

---

## Non-Interactive Context

Запуск без интерактивного TTY-ожидания ввода.

| Attribute | Rule |
|-----------|------|
| stdin piped/closed | Existing commands that ignore stdin MUST complete |
| Prompts | MUST NOT block on confirm for current commands |
| Color | stdout non-TTY → no ANSI (F06) |
| Exit | Same Exit Code Contract |

---

## Validation rules (cross-cutting)

1. Public docs table MUST list all four codes with meanings matching TZ §10.
2. Root `--help` MUST mention all four codes (brief).
3. DoD tests MUST cover: show success streams; validate exit matrix; ≥1 misuse → 1 + empty stdout.
4. Pipe scenarios document MUST cover all four TZ examples with status.
