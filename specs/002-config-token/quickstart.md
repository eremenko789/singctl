# Quickstart: Config & Token Storage (F02)

**Feature**: `002-config-token`

Этот гайд предназначен для разработчиков и тестировщиков: быстро проверить end-to-end UX и контракты F02 без CRUD сущностей и без обязательных сетевых вызовов.

---
## Preconditions

1. Сборка бинарника выполнена: `make build` (или другой эквивалент).
2. Установлен (или подготовлен) `TOKEN` для ручной проверки. Для безопасного теста можно использовать фиктивный токен, если `validate` работает как локальная заглушка до появления полного HTTP/API слоя.

Рекомендуется тестировать в изолированной директории, чтобы не трогать реальный пользовательский конфиг.

---
## Step 0: Изолировать `XDG_CONFIG_HOME`

```bash
export XDG_CONFIG_HOME="$(mktemp -d)"
```

---
## Step 1: Убедиться, что `config show` обрабатывает отсутствие конфига

```bash
singctl config show
```

Ожидаемо:
- ненулевой exit code;
- предсказуемое сообщение об отсутствии/пустоте конфига;
- токен не печатается.

---
## Step 2: Записать токен

```bash
singctl config set-token "$TOKEN"
```

Ожидаемо:
- создание `"$XDG_CONFIG_HOME/singctl/config.yaml"` (при необходимости);
- exit code 0.

---
## Step 3: Проверить вывод `config show`

1. По умолчанию (человекочитаемо):
```bash
singctl config show
```

2. Машинные форматы:
```bash
singctl config show -o json
singctl config show -o yaml
```

Ожидаемо:
- везде присутствуют эффективные значения (base_url/timeout/output и т.п.);
- `api.token` отображается **только** в замаскированном виде:
  `первые 4 + **** + последние 4` (или `****` для коротких).

---
## Step 4: Проверить `config set`

Пример (меняем base_url):
```bash
singctl config set api.base_url https://example.invalid
```

Ожидаемо:
- exit code 0;
- обновлённое значение видно в `config show`.

Пример ошибки (недопустимый ключ):
```bash
singctl config set api.no_such_key value
```

Ожидаемо:
- ненулевой exit code;
- существующий конфиг не повреждён.

---
## Step 5: Проверить `config validate`

```bash
singctl config validate
```

Ожидаемо:
- если token отсутствует — ошибка с подсказкой `config set-token`;
- если token есть — локальная проверка пройдена, удалённая проверка отложена (без ложного OK).
