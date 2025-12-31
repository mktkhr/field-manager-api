# CLAUDE.md

Go + Gin + oapi-codegen を使用した Package by Feature + クリーンアーキテクチャ構成のAPIサーバー。

## 前提
ユーザーはGoのスペシャリストですが、作業の効率化を図るためにあなたの作業を依頼しています。
ClaudeもGoのスペシャリストです。修正はGoのベストプラクティスに沿って実行してください。

## 絶対に守るルール
### 対話ルール
Claudeがユーザーと対話する場合は必ず日本語で行ってください。思考プロセスにおいて英語の方が都合が良ければ英語で思考して構いません。

### 禁止事項
- 全角（）の使用
- Testをパスさせるために、Skipを使用すること
- Test以外でのerrorの `_` を使ったハンドリングの回避
- 機能(Feature)パッケージ間の直接的な `import` での参照
  - ある機能(Consumer)が別の機能(Provider)のロジックを必要とする場合、**Consumer側で必要なインターフェースを定義** し、Providerの実装に依存しないようにして対応してください
- **英語でのコメント・ログ出力の使用**
  - すべてのコード内コメント、ログメッセージ、エラーメッセージは日本語で記載すること
  - テストコード内のコメント、エラーメッセージも日本語で記載すること

### テスト実装ガイドライン

**必須方針**: 実装と同時にテスト作成(TDD推奨)

**フレームワーク**: testify/suite + testify/mock + httptest / 統合テスト=`*_integration_test.go`

**ファイル構成**: `*_test.go`(単体) / `*_integration_test.go`(統合) / `mock_*_test.go`(モック)

**カバレッジ要件**: Presentation層=90% / Usecase層=95% / 重要BL=100% / Domain層=100%

**必須テストケース**:
- 正常系: 有効データで成功
- 異常系: 必須欠如/文字数違反/範囲外/フォーマット違反/不正JSON
- 境界値: 最小値/最大値/最小値-1/最大値+1

**命名規則**:
- `Test{メソッド}_{種類}_{詳細}` (Success / ValidationError_{原因} / BoundaryValue_{詳細} / BusinessLogicError_{原因})
  - **すべてのテストに日本語で `何をテストしているのか` のコメントを記載すること**

**バリデーション責務**:
- Presentation層: フィールドバリデーション (形式/文字数/必須)
- Usecase層: ビジネスルール (重複/存在/権限)

**テストコードの品質ルール**:
- **エラーハンドリング**: `require.NoError(t, err)` パターンを使用し、`_` でのエラー無視は禁止
- **TestMain内のログ出力**: `fmt.Printf` ではなく `log.Fatalf` を使用
- **エラーメッセージ**: 日本語で記載し、何が失敗したかを明確にすること

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

### 依存関係の詳細ルール
- **Infrastructure → Application の参照禁止**: Infrastructure層からApplication層のインターフェースを直接importしない
- **型変換はApplication層で**: 異なる機能間のデータ変換はApplication層(Usecase)で行う
- **機能間の依存方向**: Consumer機能がProvider機能のデータを必要とする場合、Consumer側で入力型を定義し、Application層で変換を行う

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

## 開発ワークフロー（新機能追加）

**手順**:
1. **API仕様**: `api/openapi.yaml`定義 → `make api-validate` → `make api-generate`
2. **パッケージ作成**: `mkdir -p features/<name>/{application/{query,usecase},domain/{entity,repository},infrastructure/{query,repository},presentation}`
3. **実装**: Domain(Entity+RepoIF) → Application(QueryIF+Usecase) → Infrastructure(Query/Repo実装) → Presentation(ServerIF)
4. **テスト**: 各レイヤーで`*_test.go`（単体）/`*_integration_test.go`(統合)作成
5. **DI登録**: `internal/server/router.go`に追加
6. **検証**: `make test` → `make cover` → `make lint` → `make gesec-scan` → `make build`

## データベース開発ルール

### マイグレーション

- **CLI運用**: `make migrate-*` コマンドで管理(サーバー起動時実行ではない)
- **命名規則**: `NNNNNN_動詞_対象.sql`(例: `000001_create_fields.sql`)
- **必須**: up.sql と down.sql は必ずペアで作成
- **down.sql**: 完全にロールバック可能であること(データ損失注意)

### SQLC

- **クエリファイル**: `db/queries/<テーブル名>.sql`
- **生成先**: `internal/generated/sqlc/`(編集禁止)
- **PostGIS型**: geometry型は `github.com/twpayne/go-geom` を使用

### 圃場(Field)テーブル設計

- **H3インデックス**: 4解像度(res3, res5, res7, res9)をGo側で計算、DBはVARCHAR保存
- **監査カラム**: `created_at`, `updated_at`, `created_by`, `updated_by`

### H3インデックス計算

- **計算場所**: Go側(`uber/h3-go`ライブラリ使用)
- **理由**: RDS/Auroraでh3-pg拡張が使用不可のため
- **保存形式**: VARCHAR(15)

### wagriデータ変換

- `LinearPolygon` → `Polygon` に変換
- 座標が閉じていない場合は最初の点を末尾に追加
