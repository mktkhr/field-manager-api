BINARY_NAME=server
MIGRATE_VERSION=v4.18.1

# Load .env file if exists
-include .env
export

# Database URL for migrations
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

help:	## ヘルプ
	@awk 'BEGIN {FS = ":.*##"} /^([a-zA-Z_-]+):.*##/ { printf "\033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## ビルド
	@echo "Building the Go application..."
	@go build -o ./bin/$(BINARY_NAME) ./cmd/server/main.go

run: build ## ビルド&実行
	@echo "Running the application..."
	@./bin/$(BINARY_NAME)

clean: ## バイナリ削除
	@echo "Cleaning up..."
	@rm -f ./bin/$(BINARY_NAME) coverage.out coverage.html

lint: ## Lint
	@echo "🚨 go fmt を実行中..."
	@go fmt ./...
	@echo "🚨 golangci-lint を実行中..."
	@golangci-lint run
	@echo "✅ 全てのlintの実行に成功しました"

test: ## 全テスト実行（単体テスト + 統合テスト）
	@echo "Running all tests..."
	@go test ./...

test-unit: ## 単体テストのみ実行（統合テストをスキップ）
	@echo "Running unit tests only..."
	@go test -short ./...

test-integration: ## 統合テストのみ実行（TestContainers使用）
	@echo "Running integration tests..."
	@go test -v -run ".*IntegrationTest.*" ./...

cover: ## テスト&カバレッジ出力(自動生成コード以外)
	go test -coverprofile=coverage.out $$(go list ./... | grep -v "/internal/generated")
	go tool cover -html=coverage.out -o coverage.html
	@echo "\n📊 カバレッジレポート:"
	@go tool cover -func=coverage.out

deps: ## Go モジュールの依存関係インストール
	@echo "Installing dependencies..."
	@go mod tidy

# API仕様書関連
api-install: ## oapi-codegenのインストール
	@echo "Installing oapi-codegen..."
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

api-validate: ## OpenAPI仕様書の検証
	@echo "Validating OpenAPI specification..."
	@redocly lint api/openapi.yaml

api-generate: api-install api-bundle ## OpenAPIからGoコード生成
	@echo "Generating Go code from OpenAPI specification..."
	@mkdir -p api/_build internal/generated/openapi
	@redocly bundle api/openapi.yaml -o api/_build/openapi-bundled.yaml
	@oapi-codegen -config api/oapi-codegen-types.yaml api/_build/openapi-bundled.yaml
	@oapi-codegen -config api/oapi-codegen-server.yaml api/_build/openapi-bundled.yaml
	@oapi-codegen -config api/oapi-codegen-spec.yaml api/_build/openapi-bundled.yaml
	@echo "Code generation completed."

api-bundle: ## OpenAPI仕様書を単一ファイルにバンドル
	@echo "Bundling OpenAPI specification..."
	@mkdir -p api/_build
	@redocly bundle api/openapi.yaml -o api/_build/openapi-bundled.yaml

api-gendoc: api-bundle ## API仕様書を生成
	@echo "Generation redocly API document..."
	@mkdir -p api/_doc
	@redocly build-docs api/_build/openapi-bundled.yaml -o api/_doc/index.html

air-install: ## airのインストール
	@echo "Installing air..."
	@go install github.com/air-verse/air@latest

dev: air-install ## 開発サーバー起動(ホットリロード)
	@air -c .air.toml

gosec-install: ## Gosecのインストール
	@echo "Installing gosec..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest

gosec-scan: gosec-install ## Gosecセキュリティスキャン実行
	@echo "🔍 Gosec セキュリティスキャンを実行中..."
	@rm -f gosec-report.json
	@gosec -fmt json -out gosec-report.json \
		-exclude-dir=.git \
		-exclude-dir=.go \
		-exclude-dir=vendor \
		-exclude-dir=internal/generated \
		-exclude-generated \
		-tests=false \
		-concurrency=4 \
		-severity=high \
		--quiet \
		./...; \
	GOSEC_EXIT_CODE=$$?; \
	if [ -f gosec-report.json ]; then \
		if command -v jq >/dev/null 2>&1; then \
			ISSUE_COUNT=$$(jq '.Stats.found // 0' gosec-report.json); \
		else \
			ISSUE_COUNT=$$(grep -o '"found": [0-9]*' gosec-report.json | grep -o '[0-9]*' || echo "0"); \
		fi; \
		if [ "$$ISSUE_COUNT" -gt 0 ]; then \
			echo ""; \
			echo "❌ セキュリティ上の問題が $$ISSUE_COUNT 件検出されました"; \
			echo ""; \
			echo "📋 検出された問題:"; \
			if command -v jq >/dev/null 2>&1; then \
				jq -r '.Issues[] | "  [\(.severity)] \(.file):\(.line) - \(.details)"' gosec-report.json; \
			else \
				cat gosec-report.json; \
			fi; \
			echo ""; \
			echo "📄 詳細レポート: gosec-report.json"; \
			exit 1; \
		else \
			echo "✅ セキュリティ上の問題は検出されませんでした"; \
		fi \
	else \
		echo "✅ セキュリティ上の問題は検出されませんでした"; \
		exit $$GOSEC_EXIT_CODE; \
	fi

sqlc-install: ## SQLCのインストール
	@echo "Installing sqlc..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc-generate: sqlc-install ## SQLCでGoコード生成
	@echo "Generating Go code from SQL..."
	@mkdir -p internal/generated/sqlc
	@cd db && sqlc generate
	@echo "SQLC generation completed."

generate: api-generate sqlc-generate ## 全コード生成(OpenAPI + SQLC)

# =============================================================================
# Migration (golang-migrate)
# =============================================================================
migrate-install: ## golang-migrateのインストール
	@echo "Installing golang-migrate..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

migrate-create: ## 新規マイグレーション作成 (NAME=xxx)
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required. Usage: make migrate-create NAME=xxx"; exit 1; fi
	@echo "Creating migration: $(NAME)..."
	@migrate create -ext sql -dir db/migrations -seq $(NAME)

migrate-up: ## 全マイグレーション適用
	@echo "Running migrations..."
	@migrate -path db/migrations -database "$(DB_URL)" up

migrate-up-one: ## 1つ次のマイグレーション適用
	@echo "Running next migration..."
	@migrate -path db/migrations -database "$(DB_URL)" up 1

migrate-down: ## 1つ前にロールバック
	@echo "Rolling back last migration..."
	@migrate -path db/migrations -database "$(DB_URL)" down 1

migrate-down-all: ## 全ロールバック(注意: データ損失)
	@echo "Rolling back all migrations..."
	@migrate -path db/migrations -database "$(DB_URL)" down -all

migrate-force: ## バージョン強制設定 (VERSION=xxx) ※障害復旧用
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION is required. Usage: make migrate-force VERSION=xxx"; exit 1; fi
	@echo "Forcing version: $(VERSION)..."
	@migrate -path db/migrations -database "$(DB_URL)" force $(VERSION)

migrate-version: ## 現在のバージョン確認
	@migrate -path db/migrations -database "$(DB_URL)" version

migrate-status: ## マイグレーション状態確認
	@echo "Migration status:"
	@migrate -path db/migrations -database "$(DB_URL)" version 2>&1 || true

# =============================================================================
# LocalStack
# =============================================================================
localstack-up: ## LocalStackを起動
	@echo "LocalStackを起動しています..."
	@docker compose -f docker/compose.yaml up -d localstack
	@echo "LocalStackの起動を待機中..."
	@sleep 10
	@echo "LocalStack起動完了"

localstack-logs: ## LocalStackのログを表示
	@docker compose -f docker/compose.yaml logs -f localstack

localstack-status: ## LocalStackのステータス確認
	@docker compose -f docker/compose.yaml exec localstack awslocal stepfunctions list-state-machines
	@docker compose -f docker/compose.yaml exec localstack awslocal s3 ls

localstack-build-lambda: ## wagri-fetcher Lambdaをビルド
	@echo "wagri-fetcher Lambdaをビルドしています..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cmd/wagri-fetcher/bootstrap ./cmd/wagri-fetcher
	@cd cmd/wagri-fetcher && zip -j wagri-fetcher.zip bootstrap
	@echo "ビルド完了: cmd/wagri-fetcher/wagri-fetcher.zip"

localstack-deploy-lambda: localstack-build-lambda ## wagri-fetcher LambdaをLocalStackにデプロイ
	@echo "wagri-fetcher Lambdaをビルド＆デプロイしています..."
	@docker compose -f docker/compose.yaml cp cmd/wagri-fetcher/wagri-fetcher.zip localstack:/tmp/wagri-fetcher.zip
	@docker compose -f docker/compose.yaml exec localstack awslocal lambda create-function \
		--function-name wagri-fetcher \
		--runtime provided.al2023 \
		--handler bootstrap \
		--zip-file fileb:///tmp/wagri-fetcher.zip \
		--role arn:aws:iam::000000000000:role/lambda-role \
		2>/dev/null || \
	docker compose -f docker/compose.yaml exec localstack awslocal lambda update-function-code \
		--function-name wagri-fetcher \
		--zip-file fileb:///tmp/wagri-fetcher.zip
	@echo "環境変数を設定しています..."
	@docker compose -f docker/compose.yaml exec localstack awslocal lambda update-function-configuration \
		--function-name wagri-fetcher \
		--environment "Variables={STORAGE_S3_ENABLED=false,STORAGE_ENDPOINT=http://rustfs:9000,STORAGE_BUCKET=$(STORAGE_BUCKET),STORAGE_ACCESS_KEY_ID=$(STORAGE_ACCESS_KEY_ID),STORAGE_SECRET_ACCESS_KEY=$(STORAGE_SECRET_ACCESS_KEY),STORAGE_REGION=$(STORAGE_REGION),WAGRI_BASE_URL=$(WAGRI_BASE_URL),WAGRI_CLIENT_ID=$(WAGRI_CLIENT_ID),WAGRI_CLIENT_SECRET=$(WAGRI_CLIENT_SECRET)}"
	@echo "デプロイ完了"

localstack-invoke-lambda: ## wagri-fetcher Lambdaをテスト実行
	@echo "wagri-fetcher Lambdaをテスト実行しています..."
	@docker compose -f docker/compose.yaml exec localstack awslocal lambda invoke \
		--function-name wagri-fetcher \
		--payload '{"city_code":"163210","import_job_id":"00000000-0000-0000-0000-000000000001"}' \
		/tmp/response.json
	@docker compose -f docker/compose.yaml exec localstack cat /tmp/response.json

localstack-start-workflow: ## Step Functionsワークフローをテスト実行
	@echo "Step Functionsワークフローをテスト実行しています..."
	@docker compose -f docker/compose.yaml exec localstack awslocal stepfunctions start-execution \
		--state-machine-arn arn:aws:states:us-east-1:000000000000:stateMachine:wagri-import-workflow \
		--input '{"city_code":"163210","import_job_id":"00000000-0000-0000-0000-000000000001"}'

localstack-list-executions: ## Step Functions実行履歴を表示
	@docker compose -f docker/compose.yaml exec localstack awslocal stepfunctions list-executions \
		--state-machine-arn arn:aws:states:us-east-1:000000000000:stateMachine:wagri-import-workflow

# =============================================================================
# Import Processor (EKS Job)
# =============================================================================
import-processor-build: ## import-processor Dockerイメージをビルド
	@echo "import-processor Dockerイメージをビルドしています..."
	@docker build -f docker/import-processor/Dockerfile -t import-processor:local .
	@echo "ビルド完了: import-processor:local"

import-processor-run: ## import-processorをローカル実行 (S3_KEY=xxx IMPORT_JOB_ID=xxx)
	@if [ -z "$(S3_KEY)" ] || [ -z "$(IMPORT_JOB_ID)" ]; then \
		echo "Error: S3_KEY and IMPORT_JOB_ID are required."; \
		echo "Usage: make import-processor-run S3_KEY=imports/163210/xxx.json IMPORT_JOB_ID=xxx"; \
		exit 1; \
	fi
	@echo "import-processorをローカル実行しています..."
	@docker run --rm \
		--network field_manager_network \
		-e STORAGE_S3_ENABLED=false \
		-e STORAGE_ENDPOINT=http://rustfs:9000 \
		-e STORAGE_BUCKET=$(STORAGE_BUCKET) \
		-e STORAGE_ACCESS_KEY_ID=$(STORAGE_ACCESS_KEY_ID) \
		-e STORAGE_SECRET_ACCESS_KEY=$(STORAGE_SECRET_ACCESS_KEY) \
		-e STORAGE_REGION=$(STORAGE_REGION) \
		-e DB_HOST=postgres \
		-e DB_PORT=5432 \
		-e DB_USER=$(DB_USER) \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e DB_NAME=$(DB_NAME) \
		-e DB_SSL_MODE=disable \
		import-processor:local \
		--s3-key $(S3_KEY) \
		--import-job-id $(IMPORT_JOB_ID)

.PHONY: build run clean lint test test-unit test-integration deps api-install api-validate api-bundle api-generate api-clean gosec-install gosec-scan sqlc-install sqlc-generate generate migrate-install migrate-create migrate-up migrate-up-one migrate-down migrate-down-all migrate-force migrate-version migrate-status localstack-up localstack-logs localstack-status localstack-build-lambda localstack-deploy-lambda localstack-invoke-lambda localstack-start-workflow localstack-list-executions import-processor-build import-processor-run
