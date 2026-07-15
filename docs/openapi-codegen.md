# OpenAPI: обновление снимка и генерация Go-клиента

Повторяемая процедура для работы с REST API SingularityApp.

## Источники истины

| Что | URL |
|---|---|
| Wiki (человекочитаемо) | https://singularity-app.ru/wiki/api/ |
| Swagger UI | https://api.singularity-app.com/v2/api |
| OpenAPI JSON | https://api.singularity-app.com/v2/api-json |
| OpenAPI YAML | https://api.singularity-app.com/v2/api-yaml |

Локальные снимки в репозитории:

- `docs/api/openapi.yaml`
- `docs/api/openapi.json`
- `specs/001-singctl-cli-tui/contracts/openapi.yaml` (копия для Spec Kit contracts)

Base URL API (подтверждён wiki/Swagger): `https://api.singularity-app.com`  
Префикс путей: `/v2/...`  
Авторизация: `Authorization: Bearer <token>` (scheme `rest-token` в OpenAPI).

## 1. Обновить снимок OpenAPI

```bash
curl -fsSL "https://api.singularity-app.com/v2/api-json" -o docs/api/openapi.json
curl -fsSL "https://api.singularity-app.com/v2/api-yaml" -o docs/api/openapi.yaml
cp docs/api/openapi.yaml specs/001-singctl-cli-tui/contracts/openapi.yaml
```

Рекомендуется после обновления:

1. Просмотреть diff путей/схем.
2. Зафиксировать breaking changes в `specs/001-singctl-cli-tui/research.md` (или новой фиче).
3. Перегенерировать клиент (шаг 2).
4. Прогнать unit-тесты адаптеров.

## 2. Сгенерировать Go-клиент

Целевой генератор: [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) (client + models).

### Установка генератора

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

### Конфиг (ожидаемый)

Файл `api/oapi-codegen.yaml` (создаётся при реализации codegen-задач):

```yaml
package: apiclient
generate:
  models: true
  client: true
output: internal/apiclient/client.gen.go
```

### Команда генерации

```bash
# после появления go.mod и конфига
oapi-codegen -config api/oapi-codegen.yaml docs/api/openapi.yaml
```

Либо через `go generate`:

```go
//go:generate oapi-codegen -config ../../api/oapi-codegen.yaml ../../docs/api/openapi.yaml
```

в пакете `internal/apiclient`.

### Альтернатива

Допустим `openapi-generator-cli` (generator `go`), если команде удобнее его пайплайн. Выбор фиксируется в `research.md`; в любом случае **ручные DTO/HTTP-методы для CRUD сущностей запрещены** (см. constitution).

## 3. Слой поверх сгенерированного клиента

В `internal/api/` (или аналог) допускается только:

- подстановка base URL и Bearer-токена;
- timeout / retry (в т.ч. 429 exponential backoff);
- нормализация ошибок в типы CLI/TUI;
- фасады доменных операций (`ListTasks`, `ArchiveTask`, …), вызывающие сгенерированные методы.

Не копировать схемы вручную: при расхождении с API — обновить OpenAPI и перегенерировать.

## 4. Проверка без токена

```bash
# спека валидна и содержит ожидаемые пути
python3 -c "import json; p=json.load(open('docs/api/openapi.json')); print(sorted(p['paths']))"
```

С токеном (локально, не в CI без секретов):

```bash
curl -fsSL -H "Authorization: Bearer $SINGCTL_TOKEN" \
  "https://api.singularity-app.com/v2/project?maxCount=1"
```

## 5. Что коммитить

| Файл | Коммитить? |
|---|---|
| `docs/api/openapi.yaml` / `.json` | Да (воспроизводимый снимок) |
| `internal/apiclient/*.gen.go` | Да (чтобы сборка не требовала codegen) |
| Конфиг `api/oapi-codegen.yaml` | Да |
| Токены / ответы с ПДн | Нет |

## История снимка

Первый снимок: **2026-07-15**, OpenAPI `3.0.0`, API version `2.0`, title `Singularity`.
