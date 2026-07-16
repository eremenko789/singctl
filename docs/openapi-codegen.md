# OpenAPI: обновление снимка и генерация Go-клиента

Повторяемые шаги запускаются через **Makefile** (constitution §VIII). Параметры — из `.env` (см. `.env.example`).

## Источники

| Что | URL |
|---|---|
| Wiki | https://singularity-app.ru/wiki/api/ |
| Swagger UI | https://api.singularity-app.com/v2/api |
| OpenAPI JSON | https://api.singularity-app.com/v2/api-json |
| OpenAPI YAML | https://api.singularity-app.com/v2/api-yaml |

Локально: `docs/api/openapi.yaml`, `docs/api/openapi.json`.
Матрица покрытия CLI: `docs/api/coverage.md`.

Base URL: `https://api.singularity-app.com`
Auth: `Authorization: Bearer <token>` (OpenAPI scheme `rest-token`).

## Команды

```bash
cp .env.example .env          # при необходимости задать API_BASE_URL / SINGCTL_TOKEN
make openapi-fetch            # обновить снимок
make api-coverage-check       # 51 operations + наличие coverage.md
make generate                 # после появления api/oapi-codegen.yaml
make smoke                    # нужен SINGCTL_TOKEN
```

Эквивалент вручную (не предпочтительно):

```bash
curl -fsSL "https://api.singularity-app.com/v2/api-json" -o docs/api/openapi.json
curl -fsSL "https://api.singularity-app.com/v2/api-yaml" -o docs/api/openapi.yaml
```

## Генерация Go-клиента

Генератор: [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen).

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

Ожидаемый конфиг `api/oapi-codegen.yaml` (создаётся при реализации):

```yaml
package: apiclient
generate:
  models: true
  client: true
output: internal/apiclient/client.gen.go
```

Затем `make generate`.

Ручные DTO/HTTP CRUD запрещены (constitution). Допускаются только адаптеры над codegen.

## Что коммитить

| Файл | Коммитить? |
|---|---|
| `docs/api/openapi.yaml` / `.json` | Да |
| `docs/api/coverage.md` | Да (обновлять вместе со снимком) |
| `internal/apiclient/*.gen.go` | Да (когда появятся) |
| `api/oapi-codegen.yaml` | Да |
| `.env` / токены | Нет |

## История снимка

Первый снимок: **2026-07-15**, OpenAPI `3.0.0`, API version `2.0`, **51** HTTP operation.
