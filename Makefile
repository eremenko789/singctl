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

help: ## Показать доступные таргеты (рекомендуемый порядок OpenAPI: openapi-fetch → api-coverage-check → generate)
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n" } /^[a-zA-Z0-9_-]+:.*?##/ { printf "  %-22s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

openapi-fetch: ## Скачать снимок OpenAPI в docs/api/ (атомарно JSON+YAML)
	@mkdir -p docs/api
	@chmod +x scripts/openapi_fetch.sh
	./scripts/openapi_fetch.sh "$(OPENAPI_JSON_URL)" "$(OPENAPI_YAML_URL)" docs/api/openapi.json docs/api/openapi.yaml
	@echo "Updated docs/api/openapi.{json,yaml} from $(API_BASE_URL)"

api-coverage-check: ## Сверить число HTTP operations в docs/api/openapi.json
	@python3 scripts/count_openapi_ops.py docs/api/openapi.json \
		--expected "$(EXPECTED_API_OPS)" \
		--require-coverage-md docs/api/coverage.md

generate: ## Сгенерировать Go-клиент (нужен api/oapi-codegen.yaml)
	@test -f api/oapi-codegen.yaml || (echo "Missing api/oapi-codegen.yaml — создайте конфиг (см. docs/openapi-codegen.md)" >&2; exit 1)
	@command -v oapi-codegen >/dev/null || (echo "Install: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest" >&2; exit 1)
	@mkdir -p internal/apiclient
	oapi-codegen -config api/oapi-codegen.yaml docs/api/openapi.yaml

build: ## Собрать бинарник singctl
	@test -f go.mod || (echo "go.mod ещё не создан — сначала инициализация модуля / Spec Kit implement" >&2; exit 1)
	go build -o bin/singctl ./cmd/singctl

# Coverage profile excludes generated *.gen.go (constitution IX MAY).
test: ## Запустить unit-тесты с coverage (+ скриптовые контракты OpenAPI)
	@test -f go.mod || (echo "go.mod ещё не создан" >&2; exit 1)
	go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic
	@awk '!/\.gen\.go:/' coverage.out > coverage.nogenerated.out
	@go tool cover -func=coverage.nogenerated.out | tail -1
	@python3 -m unittest scripts/count_openapi_ops_test.py -v
	@chmod +x scripts/openapi_fetch_atomic_test.sh
	@./scripts/openapi_fetch_atomic_test.sh

pre-commit: ## Прогнать все pre-commit hooks по всем файлам
	@command -v pre-commit >/dev/null || (echo "Install: pipx install pre-commit  (или brew install pre-commit)" >&2; exit 1)
	pre-commit run --all-files

smoke: ## Smoke GET /v2/project (требует SINGCTL_TOKEN в .env)
	@test -n "$(SINGCTL_TOKEN)" || (echo "Set SINGCTL_TOKEN in .env" >&2; exit 1)
	curl -fsSL -H "Authorization: Bearer $(SINGCTL_TOKEN)" \
	  "$(API_BASE_URL)/v2/project?maxCount=1" >/dev/null
	@echo "smoke OK"
