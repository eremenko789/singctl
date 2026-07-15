# CLI Contract: singctl root (F01)

**Feature**: `001-cli-skeleton` | **Date**: 2026-07-15
**Audience**: пользователи CLI, тесты, последующие фичи (F02+)

Машинные имена — латиница; пользовательские тексты help/ошибок — русский (см. [spec.md](../spec.md) FR-014).

---

## Binary

| Item | Value |
|------|-------|
| Name | `singctl` |
| Build | `make build` → `bin/singctl` from `./cmd/singctl` |

---

## Invocation grammar

```text
singctl [--config <path>] [--token <string>] [--output|-o <format>] [--no-color] [--debug]
        [--help|-h] [--version]
        [<command> [command-flags...]]

<format> ::= table | json | yaml | csv

<command> ::= version
            | help          # стандартный cobra help, если включён
```

Запрещено регистрировать в F01: `task`, `project`, `habit`, `tag`, `time`, `tui`.

---

## Global flags

| Flag | Short | Type | Default | Required | Effect in F01 |
|------|-------|------|---------|----------|---------------|
| `--config` | — | path string | unset | no | accepted, not resolved |
| `--token` | — | string | unset | no | accepted, not persisted |
| `--output` | `-o` | enum | `table` | no | validated; formatting of entities deferred |
| `--no-color` | — | bool | false | no | accepted |
| `--debug` | — | bool | false | no | accepted |
| `--help` | `-h` | bool | — | no | root help (RU) |
| `--version` | — | bool | — | no | same payload as `version` |

Unknown global flags → non-zero exit + error message.

Invalid `--output` (any operation including `--help` / `version` / `--version`) → non-zero exit + RU validation error; **no** help/version body.

---

## Commands

### (none) — bare `singctl`

| | |
|--|--|
| Exit | ≠ 0 |
| stdout | empty (prefer) |
| stderr | RU error: command missing / TUI not implemented |
| Network | none |
| TUI | must not start |

### `singctl --help`

| | |
|--|--|
| Exit | 0 (если флаги валидны) |
| stdout | RU: product blurb, command list, global flags |
| Must include | `--config`, `--token`, `--output`/`-o`, `--no-color`, `--debug` |
| Must not include | entity/TUI command names |
| Network | none |

### `singctl version` / `singctl --version`

| | |
|--|--|
| Exit | 0 (если флаги валидны) |
| stdout | identical for both forms: CLI name `singctl`, version, build metadata (commit and/or date) |
| stderr | empty on success |
| Network | none |

Placeholders `dev` / `unknown` allowed when ldflags absent.

---

## Exit code summary

| Situation | Code |
|-----------|------|
| Help / version success | 0 |
| Bare invoke / unknown command | ≠ 0 |
| Invalid `--output` | ≠ 0 |
| Unknown flag | ≠ 0 |

Точные числовые коды (1 vs 2) уточняются в F07; F01 требует лишь **ненулевой** для ошибок.

---

## stdout / stderr

- Успешный version → **stdout**
- Справка → **stdout** (типично для Cobra)
- Ошибки парсера/валидации/нет команды → **stderr**

---

## Extensibility (non-breaking for later features)

Последующие фичи MAY:

1. Добавлять subcommands под root без изменения семантики глобальных флагов.
2. Читать `GlobalOptions` через Viper / shared context.
3. Реализовать загрузку `--config` и применение `--token` (F02).
4. Использовать `Output` / `NoColor` / `Debug` в рендере и HTTP (F04–F07).

Последующие фичи MUST NOT:

1. Переименовывать глобальные флаги без deprecate-цикла.
2. Расширять enum `--output` без обновления spec/contract.
3. Обещать в root help команды до их регистрации.
