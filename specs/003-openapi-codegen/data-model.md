# Data Model: OpenAPI Codegen Pipeline (F03)

**Feature**: `003-openapi-codegen` | **Date**: 2026-07-16

Модель описывает артефакты пайплайна (не runtime CLI config из F02): снимок OpenAPI, конфиг генерации, ожидание coverage и сгенерированный клиент.

---
## Entities

### OpenAPI Snapshot

Локальная копия контракта SingularityApp REST API v2.

| Artifact | Path | Role |
|----------|------|------|
| JSON snapshot | `docs/api/openapi.json` | вход `api-coverage-check` (подсчёт operations) |
| YAML snapshot | `docs/api/openapi.yaml` | вход `make generate` |
| Upstream JSON URL | `$(API_BASE_URL)/v2/api-json` | источник `openapi-fetch` |
| Upstream YAML URL | `$(API_BASE_URL)/v2/api-yaml` | источник `openapi-fetch` |

**Invariants**:
- После успешного `openapi-fetch` JSON и YAML соответствуют одному и тому же upstream-состоянию (атомарная пара).
- Для офлайн-generate достаточно закоммиченного YAML (+ конфиг).

**Validation**:
- JSON парсится; `paths` — объект.
- Operation count = число HTTP-методов под `paths.*`, исключая ключи, начинающиеся с `x-`.

---
### Coverage Expectation

Зафиксированное ожидание полноты снимка относительно матрицы CLI (матрица сама по себе не парсится check’ом в F03).

| Field | Location | Type | Current value |
|-------|----------|------|---------------|
| `EXPECTED_API_OPS` | Makefile / `.env` override | int | `51` |
| Coverage matrix file | `docs/api/coverage.md` | file presence | must exist |

**Rules**:
- VR-C01: `operations_in_json == EXPECTED_API_OPS` → иначе `api-coverage-check` fail.
- VR-C02: отсутствие `coverage.md` → fail.
- VR-C03: устаревшее **содержимое** матрицы при верном числе ops → в F03 **не** fail (процесс/ревью).

---
### Codegen Config

Параметры генератора клиента.

| Field | Path / key | Value (F03) |
|-------|------------|-------------|
| Config file | `api/oapi-codegen.yaml` | required for `generate` |
| `package` | yaml | `apiclient` |
| `generate.models` | yaml | `true` |
| `generate.client` | yaml | `true` |
| `output` | yaml | `internal/apiclient/client.gen.go` |

**Rules**:
- VR-G01: отсутствие конфига → `make generate` exit ≠ 0 + понятное сообщение.
- VR-G02: отсутствие `oapi-codegen` в PATH → exit ≠ 0 + подсказка install.

---
### Generated API Client

Автоматически полученный Go-пакет.

| Field | Value |
|-------|-------|
| Package | `apiclient` |
| Path | `internal/apiclient/client.gen.go` (и при необходимости другие `*.gen.go`, если генератор создаст) |
| Source of truth in repo | committed files (DoD F03) |
| Manual edits | forbidden as behavior source; regenerate instead |
| Coverage | MAY exclude from `make test` coverage gate |

**Relationships**:
- Derived from OpenAPI Snapshot (YAML) + Codegen Config.
- Consumed later by F04 adapter (out of scope).

---
### Pipeline Environment

| Variable | Default | Used by |
|----------|---------|---------|
| `API_BASE_URL` | `https://api.singularity-app.com` | openapi-fetch URLs |
| `OPENAPI_JSON_URL` | `$(API_BASE_URL)/v2/api-json` | openapi-fetch |
| `OPENAPI_YAML_URL` | `$(API_BASE_URL)/v2/api-yaml` | openapi-fetch |
| `EXPECTED_API_OPS` | `51` | api-coverage-check |
| `SINGCTL_TOKEN` | empty | **not** required for F03 fetch/generate |

`.env` подключается Makefile’ом если существует; не коммитится.

---
## State transitions

```text
[no config] --(add oapi-codegen.yaml)--> [ready to generate]
[stale/missing snapshot] --(openapi-fetch OK)--> [snapshot current]
[snapshot current] --(api-coverage-check OK)--> [ops expectation met]
[ready to generate + YAML snapshot] --(make generate OK)--> [client.gen.go present]
[client.gen.go present] --(git commit DoD)--> [F03 delivered]
```

Независимые переходы: generate допускается при красном coverage-check; fetch не требует generate.

---
## Out of model (F04+)

- Auth header / Bearer wiring
- Retry, error mapping, CLI facades
- Drift/no-diff verification entity
