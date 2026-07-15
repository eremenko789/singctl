# Документация sa-cli / singctl

## Быстрый указатель

| Документ | Назначение |
|---|---|
| [tz/singularityapp-cli-tui-tz.md](./tz/singularityapp-cli-tui-tz.md) | Исходное продуктовое ТЗ |
| [api/openapi.yaml](./api/openapi.yaml) | Снимок OpenAPI 3.0 SingularityApp REST API |
| [api/wiki-rest-api.md](./api/wiki-rest-api.md) | Текстовый снимок wiki REST API |
| [openapi-codegen.md](./openapi-codegen.md) | Как обновлять OpenAPI и генерировать Go-клиент |
| [spec-kit.md](./spec-kit.md) | Как уложены артефакты GitHub Spec Kit в репозитории |

## Spec Kit

Методология: [GitHub Spec Kit](https://github.com/github/spec-kit) (Spec → Plan → Tasks → Implement).

- Конституция: [`.specify/memory/constitution.md`](../.specify/memory/constitution.md)
- Фича MVP/продукт: [`specs/001-singctl-cli-tui/`](../specs/001-singctl-cli-tui/)

## Внешние источники API

- Wiki: https://singularity-app.ru/wiki/api/
- Swagger UI: https://api.singularity-app.com/v2/api
- OpenAPI JSON: https://api.singularity-app.com/v2/api-json
- OpenAPI YAML: https://api.singularity-app.com/v2/api-yaml
- Личный кабинет (токены): https://me.singularity-app.com
