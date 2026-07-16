# sa-cli / singctl — регулярные задачи разработки
# Параметры: файл .env (см. .env.example). Не коммитьте .env.

.DEFAULT_GOAL := help

ifneq (,$(wildcard .env))
  include .env
  export
endif

API_BASE_URL ?= https://api.singularity-app.com
OPENAPI_JSON_URL ?= $(API_BASE_URL)/v2/api-json
OPENAPI_YAML_URL ?= $(API_BASE_URL)/v2/api-yaml
EXPECTED_API_OPS ?= 51

.PHONY: help openapi-fetch api-coverage-check generate build test pre-commit smoke

help: ## Показать доступные таргеты
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n" } /^[a-zA-Z0-9_-]+:.*?##/ { printf "  %-22s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

openapi-fetch: ## Скачать снимок OpenAPI в docs/api/
	@mkdir -p docs/api
	curl -fsSL "$(OPENAPI_JSON_URL)" -o docs/api/openapi.json
	curl -fsSL "$(OPENAPI_YAML_URL)" -o docs/api/openapi.yaml
	@echo "Updated docs/api/openapi.{json,yaml} from $(API_BASE_URL)"

api-coverage-check: ## Сверить число HTTP operations в docs/api/openapi.json
	@python3 -c 'import json,sys; p=json.load(open("docs/api/openapi.json")); n=sum(1 for methods in p["paths"].values() for m in methods if not m.startswith("x-")); exp=int("$(EXPECTED_API_OPS)"); \
print(f"operations={n} expected={exp}"); \
sys.exit(0 if n==exp else 1)'
	@test -f docs/api/coverage.md
	@echo "OK: coverage matrix present at docs/api/coverage.md"

generate: ## Сгенерировать Go-клиент (нужен api/oapi-codegen.yaml)
	@test -f api/oapi-codegen.yaml || (echo "Missing api/oapi-codegen.yaml — создайте на этапе реализации codegen (см. docs/openapi-codegen.md)" >&2; exit 1)
	@command -v oapi-codegen >/dev/null || (echo "Install: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest" >&2; exit 1)
	oapi-codegen -config api/oapi-codegen.yaml docs/api/openapi.yaml

build: ## Собрать бинарник singctl
	@test -f go.mod || (echo "go.mod ещё не создан — сначала инициализация модуля / Spec Kit implement" >&2; exit 1)
	go build -o bin/singctl ./cmd/singctl

test: ## Запустить unit-тесты
	@test -f go.mod || (echo "go.mod ещё не создан" >&2; exit 1)
	go test ./... -v

pre-commit: ## Прогнать все pre-commit hooks по всем файлам
	@command -v pre-commit >/dev/null || (echo "Install: pipx install pre-commit  (или brew install pre-commit)" >&2; exit 1)
	pre-commit run --all-files

smoke: ## Smoke GET /v2/project (требует SINGCTL_TOKEN в .env)
	@test -n "$(SINGCTL_TOKEN)" || (echo "Set SINGCTL_TOKEN in .env" >&2; exit 1)
	curl -fsSL -H "Authorization: Bearer $(SINGCTL_TOKEN)" \
	  "$(API_BASE_URL)/v2/project?maxCount=1" >/dev/null
	@echo "smoke OK"
