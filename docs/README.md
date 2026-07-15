# Документация sa-cli / singctl

## Быстрый указатель

| Документ | Назначение |
|---|---|
| [tz/singularityapp-cli-tui-tz.md](./tz/singularityapp-cli-tui-tz.md) | Исходное продуктовое ТЗ |
| [api/coverage.md](./api/coverage.md) | Матрица покрытия всех 51 REST operations |
| [api/openapi.yaml](./api/openapi.yaml) | Снимок OpenAPI 3.0 |
| [api/wiki-rest-api.md](./api/wiki-rest-api.md) | Текстовый снимок wiki REST API |
| [openapi-codegen.md](./openapi-codegen.md) | Обновление OpenAPI и codegen |
| [makefile.md](./makefile.md) | Make-таргеты и `.env` |
| [spec-kit.md](./spec-kit.md) | Отправные файлы vs результаты Spec Kit |

## Spec Kit

Методология: [GitHub Spec Kit](https://github.com/github/spec-kit).

- Конституция: [`.specify/memory/constitution.md`](../.specify/memory/constitution.md)
- Артефакты `specs/` появятся после `/speckit.specify` и следующих команд — см. [spec-kit.md](./spec-kit.md)

## Автоматизация

```bash
cp .env.example .env
make help
make openapi-fetch
make api-coverage-check
```

## Внешние источники API

- Wiki: https://singularity-app.ru/wiki/api/
- Swagger UI: https://api.singularity-app.com/v2/api
- OpenAPI JSON: https://api.singularity-app.com/v2/api-json
- OpenAPI YAML: https://api.singularity-app.com/v2/api-yaml
- Личный кабинет (токены): https://me.singularity-app.com
