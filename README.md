# sa-cli (`singctl`)

CLI и TUI-клиент для [SingularityApp REST API](https://singularity-app.ru/wiki/api/).

> Статус: отправные docs + constitution. Артефакты Spec Kit (`specs/…`) ещё не генерировались — их создают команды `/speckit.*`.

## Документация

| Путь | Содержание |
|---|---|
| [`docs/`](./docs/README.md) | Индекс |
| [`docs/tz/singularityapp-cli-tui-tz.md`](./docs/tz/singularityapp-cli-tui-tz.md) | Исходное ТЗ |
| [`docs/api/coverage.md`](./docs/api/coverage.md) | Все 51 REST operations → CLI |
| [`docs/makefile.md`](./docs/makefile.md) | Make + `.env` |
| [`.specify/memory/constitution.md`](./.specify/memory/constitution.md) | Принципы |
| [`docs/spec-kit.md`](./docs/spec-kit.md) | Что bootstrap, что Spec Kit |
| [`docs/feature-backlog.md`](./docs/feature-backlog.md) | Фичи F01–F39 для `/speckit-specify` |

## Стек (constitution)

- **Go** — бинарник `singctl`
- **OpenAPI codegen** — клиент API
- **Makefile + `.env`** — регулярные задачи разработки

```bash
cp .env.example .env
make help
make openapi-fetch
make api-coverage-check
```

## API

- Swagger: https://api.singularity-app.com/v2/api
- OpenAPI: https://api.singularity-app.com/v2/api-json
- Снимок: [`docs/api/openapi.yaml`](./docs/api/openapi.yaml)

Токен: [личный кабинет](https://me.singularity-app.com).

## Следующий шаг (Spec Kit)

Брать **одну** фичу из [`docs/feature-backlog.md`](./docs/feature-backlog.md) (старт с F01):

```text
/speckit-specify …  # карточка Fxx + docs/tz + coverage + constitution
/speckit-clarify …
/speckit-plan …
/speckit-tasks …
/speckit-implement …
```

## Лицензия

См. [LICENSE](./LICENSE).
