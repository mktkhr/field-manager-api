BINARY_NAME=server

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

.PHONY: build run clean lint test test-unit test-integration deps api-install api-validate api-bundle api-generate api-clean gosec-install gosec-scan sqlc-install sqlc-generate generate
