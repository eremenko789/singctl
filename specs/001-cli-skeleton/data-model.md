# Data Model: CLI Skeleton (F01)

**Feature**: `001-cli-skeleton` | **Date**: 2026-07-15

Персистентного хранилища в F01 нет. Модель описывает **сеансовые** сущности парсера и отображения версии.

---

## Entities

### GlobalOptions

Набор корневых параметров сеанса CLI, доступных всем будущим командам через persistent flags / Viper keys.

| Field | Type | Default | Validation | Notes |
|-------|------|---------|------------|-------|
| `ConfigPath` | string (path) | empty | не резолвится в F01 | флаг `--config` |
| `Token` | string | empty | не валидируется на формат в F01 | флаг `--token`; не логируется |
| `Output` | enum | `table` | MUST ∈ {`table`,`json`,`yaml`,`csv`} | флаги `--output` / `-o`; ошибка до help/version |
| `NoColor` | bool | `false` | — | флаг `--no-color`; эффект на рендер сущностей — F06+ |
| `Debug` | bool | `false` | — | флаг `--debug`; полный HTTP-лог — поздние фичи |

**Relationships**: принадлежит корневому сеансу (`RootCommand` / runtime context). Будущие команды читают те же опции; F02 мапит `ConfigPath`/`Token` в файл конфига.

**Viper keys (рекомендуемые)**: `config`, `token`, `output`, `no-color` / `color=false`, `debug` — согласовать с именами флагов при `BindPFlag`.

---

### VersionIdentity

Идентичность установленной сборки, печатная на stdout.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `Name` | string | yes | константа `singctl` |
| `Version` | string | yes | semver или `dev` |
| `Commit` | string | yes | git SHA или `unknown` |
| `Date` | string | yes | дата сборки или `unknown` |

**Relationships**: независима от `GlobalOptions` (кроме того, что невалидный `Output` блокирует печать). Источник — `internal/buildinfo` (+ ldflags).

**Display rule**: `singctl version` и `singctl --version` MUST выдавать одинаковое содержимое (имя, версия, commit и/или date).

---

## Validation Rules (from requirements)

1. **VR-001**: Любое значение `Output` вне `{table,json,yaml,csv}` → ошибка на русском, exit ≠ 0; help/version не показываются.
2. **VR-002**: Неизвестный глобальный флаг → ошибка, exit ≠ 0.
3. **VR-003**: Пустой argv после имени бинарника → ошибка (нет команды / TUI не реализован), exit ≠ 0.
4. **VR-004**: Неизвестная подкоманда → ошибка, exit ≠ 0, подсказка к `--help`.
5. **VR-005**: F01 не требует существования файла по `ConfigPath` и не требует непустого `Token`.

---

## State Transitions

Для F01 состояние минимально:

```text
[ParseArgs]
    ├─ invalid Output / unknown flag → Failed(exit≠0)
    ├─ --help (валидные флаги) → PrintedHelp(exit=0)
    ├─ --version | version (валидные флаги) → PrintedVersion(exit=0)
    ├─ no args → FailedNoCommand(exit≠0)
    └─ unknown command → FailedUnknown(exit≠0)
```

Персистентных state machine нет.
