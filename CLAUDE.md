# CLAUDE.md

Go + Gin + oapi-codegen を使用した Package by Feature + クリーンアーキテクチャ構成のAPIサーバー。

## ディレクトリ構成

```
project-root/
├── api/                    # OpenAPI仕様書、生成設定
├── bin/                    # ビルド成果物
├── cmd/server/             # エントリーポイント(main.go)
├── config/                 # 設定ファイル、秘密鍵
├── db/queries/             # SQLCクエリファイル(.sql)
├── docker/                 # Docker設定
├── docs/                   # ドキュメント
├── internal/               # アプリケーション本体
└── scripts/                # ユーティリティスクリプト
```

## internal/ 構成

| ディレクトリ         | 役割                                                 |
| -------------------- | ---------------------------------------------------- |
| `features/`          | 機能別パッケージ(auth, token, session, user, client) |
| `generated/openapi/` | oapi-codegen生成コード(編集禁止)                     |
| `generated/sqlc/`    | SQLC生成コード(編集禁止)                             |
| `server/`            | Ginルーター、DI、サーバー起動                        |
| `middleware/`        | 認証、CORS、ログ等のミドルウェア                     |
| `infrastructure/`    | DB接続(postgres, redis, valkey)、マイグレーション    |
| `config/`            | 設定構造体、ローダー                                 |
| `apperror/`          | カスタムエラー型                                     |
| `logger/`            | slogベースのロギング                                 |
| `utils/`             | 共通ユーティリティ                                   |

## Feature パッケージ構成(クリーンアーキテクチャ)

```
internal/features/<feature>/
├── domain/
│   ├── entity/         # エンティティ、Value Object
│   └── repository/     # リポジトリインターフェース
├── application/
│   ├── usecase/        # ユースケース実装
│   ├── port/           # 外部サービスインターフェース
│   └── query/          # CQRS読み取り系インターフェース
├── infrastructure/
│   ├── repository/     # リポジトリ実装(DB書き込み)
│   ├── query/          # クエリ実装(DB読み取り)
│   └── external/       # 外部サービス実装
└── presentation/
    └── handler.go      # HTTPハンドラー(ServerInterface実装)
```

## 依存関係ルール

```
Presentation → Application → Domain ← Infrastructure
```

- 内側の層は外側に依存しない
- Domain層でインターフェース定義、Infrastructure層で実装

## 主要コマンド

| コマンド             | 説明                         |
| -------------------- | ---------------------------- |
| `make build`         | ビルド                       |
| `make run`           | ビルド&実行                  |
| `make test`          | 全テスト実行                 |
| `make lint`          | Lint実行                     |
| `make api-generate`  | OpenAPIからコード生成        |
| `make sqlc-generate` | SQLCでコード生成             |
| `make generate`      | 全コード生成                 |
| `make dev`           | 開発サーバー(ホットリロード) |

## Gitコミット

**フォーマット**: `{type}:{emoji}{対象の説明}(#チケット番号)`
**Type**: add(新機能) / fix(バグ) / update(改善) / refactor / docs / test / style / chore / remove
**Emoji**: 📝(ドキュメント) / 🐛(バグ) / ⚡(改善) / ♻️(リファクタ) / 📚(ドキュメント) / 🧪(テスト) / 🎨(UI) / 🔧(設定) / 🗑️(削除) / ✨(新機能) / 🔒(セキュリティ)

**重要ルール**:
- "Generated with Claude Code" / "Co-Authored-By: Claude" 含めない
- 変更内容のみ記述(理由・影響明記)
- チケット番号=ブランチ名`feature/xxxxx`の`xxxxx`部分

**粒度**: 1機能=1コミット / ファイル種別別 / 影響範囲別 / WIP禁止

## 禁止事項
- 全角（）の使用
- Testをパスさせるために、Skipを使用すること
- Test以外でのerrorの `_` を使ったハンドリングの回避
- 機能(Feature)パッケージ間の直接的な `import` での参照
  - ある機能(Consumer)が別の機能(Provider)のロジックを必要とする場合、**Consumer側で必要なインターフェースを定義** し、Providerの実装に依存しないようにして対応してください