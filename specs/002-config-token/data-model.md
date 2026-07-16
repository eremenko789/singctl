# Data Model: Config & Token Storage (F02)

**Feature**: `002-config-token` | **Date**: 2026-07-16

Модель описывает: (1) структуру локального YAML-конфига, (2) резолвинг "effective config" из нескольких источников, (3) правила маскирования токена в выводе и (4) маппинг runtime-override `--token`.

---
## Entities

### GlobalOptions (из F01)

Сеансовые опции CLI, читаемые в контексте root-команды:

| Field | Type | Default | Validation | Notes |
|------|------|---------|------------|------|
| `ConfigPath` | string | empty | accepted, not resolved | `--config` |
| `Token` | string | empty | accepted | `--token` (runtime override) |
| `Output` | enum | `table` | MUST ∈ {`table`,`json`,`yaml`,`csv`} | `--output` / `-o` |
| `NoColor` | bool | false | — | флаг для рендера |
| `Debug` | bool | false | — | влияет на детали логов (F29) |

---
### ConfigDocument

Содержимое YAML-файла конфигурации (локальный документ пользователя).

| Section | Field | Type | Required | Default |
|---------|-------|------|----------|---------|
| `api` | `base_url` | string | MAY | `https://api.singularity-app.ru` |
| `api` | `token` | string | MAY | empty |
| `api` | `timeout` | string (duration) | MAY | `30s` |
| `output` | `format` | enum | MAY | `table` |
| `output` | `color` | bool | MAY | `true` |
| `output` | `date_format` | string | MAY | `2006-01-02` |
| `tui` | `theme` | enum | MAY | `dark` |
| `tui` | `vi_keys` | bool | MAY | `true` |
| `tui` | `refresh_interval` | int | MAY | `0` |

**Token storage rule**:
- `api.token` хранит **только "голый" токен** (без `Bearer`).

**Unknown keys**:
- чтение неизвестных ключей MAY игнорировать их с сохранением round-trip, если это практично в реализации.

---
### ResolvedConfigPath

Итоговый путь к файлу после резолвинга приоритета источников:
1. `--config`
2. `./.singctl.yaml`
3. `$XDG_CONFIG_HOME/singctl/config.yaml`
4. `~/.config/singctl/config.yaml`

---
### EffectiveSettings

Снимок настроек, используемых конкретной командой исполнения:

- берётся из прочитанного `ConfigDocument` по `ResolvedConfigPath`
- применяются runtime-override:
  - если `--token` задан, он подменяет `api.token` **для runtime-использования** (файл при этом не меняется; изменение файла делает `set-token`)

---
### API Token (маскируемое представление)

Два представления одного токена:

| Name | Source | Output rule |
|------|--------|--------------|
| `api.token` | ConfigDocument | хранит "голый" токен |
| masked token | EffectiveSettings | в `config show` отображается `первые 4 + **** + последние 4` (или только `****`, если токен короче 8) |

---
## Validation Rules (from requirements)

1. **VR-001 (Token format on set-token)**: `set-token` принимает ввод **без** префикса `Bearer` и отклоняет значения, начинающиеся с `Bearer` (с пробелом после).
2. **VR-002 (config set key scope)**: `config set` разрешает только ключи, входящие в схему `ConfigDocument` (dotted path) и валидирует значения по типам/enum.
3. **VR-003 (config show masking)**: в stdout/stderr `config show` не выводит полный секрет; используется masked token rule.
4. **VR-004 (config show output)**: `config show` уважает `--output` / `-o` и отображает один и тот же effective snapshot во всех поддерживаемых форматах.

---
## State Transitions

Минимальная семантика для фичи:

```text
[set-token]
    ├─ valid token → write ConfigDocument → success
    └─ invalid token format → reject (no write)

[set]
    ├─ valid key+value → write ConfigDocument → success
    └─ invalid key/value → reject (no write)

[show]
    ├─ config resolved + readable → print effective snapshot with masked token
    └─ config missing → print "config missing/empty" with safe messaging

[validate]
    ├─ token absent → fail with hint set-token
    └─ token present → stub local check (remote validation deferred) OR real API check после появления API слоя
```
