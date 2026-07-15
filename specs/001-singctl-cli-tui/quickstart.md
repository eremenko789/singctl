# Quickstart: проверка singctl

Сценарии валидации после реализации (и частично уже сейчас — для API/docs).

## Prefetch: контракт доступен

```bash
curl -fsSL https://api.singularity-app.com/v2/api-json | python3 -c "import sys,json; print(json.load(sys.stdin)['info'])"
test -f docs/api/openapi.yaml
```

## A. Конфиг и токен (US1)

```bash
export SINGCTL_TOKEN=...   # из https://me.singularity-app.com
singctl config set-token "$SINGCTL_TOKEN"
singctl config show          # токен замаскирован
singctl config validate      # exit 0
```

Негатив: неверный токен → сообщение 401 + exit 1.

## B. Задачи MVP (US2–US3)

```bash
singctl task list --limit 5
singctl task list -o json | jq '.[0].id // .tasks[0].id'
ID=$(singctl task create --title "singctl smoke" -o json | jq -r '.id // .task.id')
singctl task get "$ID"
singctl task update "$ID" --note "updated"
singctl task archive "$ID"
singctl task delete "$ID"
```

Pipe: без цвета в `| cat`.

## C. Остальные ресурсы (US4)

```bash
singctl project list --limit 5
singctl habit list
singctl tag list
singctl time list --limit 5
```

## D. TUI (US5)

```bash
singctl tui
# 1 → Задачи, n → создать, Esc, q → выход
```

## E. Codegen regression

После обновления OpenAPI:

```bash
# см. docs/openapi-codegen.md
oapi-codegen -config api/oapi-codegen.yaml docs/api/openapi.yaml
go test ./...
```
