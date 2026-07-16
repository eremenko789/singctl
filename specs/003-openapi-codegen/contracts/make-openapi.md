# Contract: Make OpenAPI Pipeline (F03)

**Feature**: `003-openapi-codegen` | **Date**: 2026-07-16

Контракт поведения Make-таргетов пайплайна. Не описывает CLI `singctl` (вне scope).

Рекомендуемый порядок (документация, **не** Make prerequisite):

```text
make openapi-fetch → make api-coverage-check → make generate
```

---
## Shared environment

| Name | Default | Notes |
|------|---------|-------|
| `API_BASE_URL` | `https://api.singularity-app.com` | из Makefile / `.env` |
| `OPENAPI_JSON_URL` | `$(API_BASE_URL)/v2/api-json` | |
| `OPENAPI_YAML_URL` | `$(API_BASE_URL)/v2/api-yaml` | |
| `EXPECTED_API_OPS` | `51` | переопределяется через `.env` при смене снимка |

`.env` не обязателен для публичного fetch и офлайн generate.

---
## Target: `openapi-fetch`

**Purpose**: обновить локальный снимок OpenAPI из upstream.

| | |
|--|--|
| **Inputs** | `OPENAPI_JSON_URL`, `OPENAPI_YAML_URL`; сеть |
| **Outputs** | `docs/api/openapi.json`, `docs/api/openapi.yaml` |
| **Exit 0** | оба файла успешно скачаны и атомарно установлены |
| **Exit ≠ 0** | любой сбой сети/HTTP/записи; целевая пара не должна остаться намеренно рассинхронизированной (см. research §4) |

**Не требует**: `SINGCTL_TOKEN`, `api/oapi-codegen.yaml`, gen-клиента.

---
## Target: `api-coverage-check`

**Purpose**: проверить число operations и наличие матрицы.

| | |
|--|--|
| **Inputs** | `docs/api/openapi.json`, `EXPECTED_API_OPS`, наличие `docs/api/coverage.md` |
| **Outputs** | stdout с `operations=N expected=E` и подтверждением наличия матрицы |
| **Exit 0** | `N == EXPECTED_API_OPS` **и** `coverage.md` существует |
| **Exit ≠ 0** | JSON нечитаем / `N != E` / нет `coverage.md` |

**MUST NOT** (F03): парсить строки матрицы; сверять operationId.

**Независимость**: не вызывается автоматически из `generate`.

---
## Target: `generate`

**Purpose**: сгенерировать Go-клиент из YAML-снимка.

| | |
|--|--|
| **Inputs** | `api/oapi-codegen.yaml`, `docs/api/openapi.yaml`, `oapi-codegen` в PATH |
| **Outputs** | `internal/apiclient/client.gen.go` (package `apiclient`, models+client) |
| **Exit 0** | генерация завершена; файл(ы) клиента на месте |
| **Exit ≠ 0** | нет конфига (понятное сообщение + ссылка на docs допустима); нет `oapi-codegen` в PATH (подсказка install); ошибка генератора / невалидный YAML |

**MUST NOT**: требовать успешный `api-coverage-check`; писать ручные HTTP CRUD; добавлять CLI-команды.

**DoD поставки**: успешный generate + коммит конфига, снимка, `coverage.md`, `*.gen.go`.

**Не в контракте F03**: проверка «повторный generate → empty git diff».

---
## Negative / edge contracts

| Scenario | Expected |
|----------|----------|
| Нет `api/oapi-codegen.yaml` | `generate` ≠ 0, понятная ошибка |
| Нет `oapi-codegen` | `generate` ≠ 0, install hint |
| `EXPECTED_API_OPS=999` при 51 ops | `api-coverage-check` ≠ 0 |
| Удалён `coverage.md` | `api-coverage-check` ≠ 0 |
| Upstream down | `openapi-fetch` ≠ 0 |
| Офлайн + есть YAML/конфиг/tool | `generate` = 0 |

---
## Documentation sync

После реализации тексты в `docs/openapi-codegen.md` и `docs/makefile.md` MUST отражать этот контракт (порядок, строгость check, атомарный fetch).
