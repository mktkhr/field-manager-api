# api-endpoint-impl

APIエンドポイント実装を効率化するSkill

## 発動条件

以下のようなリクエストで発動する:
- 「〜APIを実装して」
- 「〜エンドポイントを実装して」
- 「一覧取得を実装して」
- 「詳細取得を実装して」
- 「登録APIを実装して」
- 「更新APIを実装して」
- 「削除APIを実装して」

## アーキテクチャ概要

このプロジェクトは **Package by Feature + クリーンアーキテクチャ** を採用している。

```
依存関係: Presentation → Application → Domain ← Infrastructure
```

### 層の責務

| 層 | 責務 |
|---|---|
| Domain | エンティティ、リポジトリインターフェース、機能間DTO |
| Application | ユースケース、クエリインターフェース(CQRS) |
| Infrastructure | DB実装(リポジトリ=書き込み、クエリ=読み取り) |
| Presentation | HTTPハンドラー、リクエスト/レスポンス変換 |

## 実装順序

以下の順序で実装を進める:

### 1. OpenAPI仕様定義

**作業内容**:
- `api/paths/{feature}.yaml` にエンドポイント定義を追加
- `api/components/schemas/{feature}.yaml` にスキーマ定義を追加
- `api/components/parameters/common.yaml` に共通パラメータを追加(必要な場合)

**コマンド**:
```bash
make api-validate   # 仕様の検証
make api-generate   # コード生成
```

**参照ファイル**:
- `api/paths/fields.yaml` - パス定義の例
- `api/components/schemas/field.yaml` - スキーマ定義の例
- `api/components/parameters/common.yaml` - 共通パラメータの例

### 2. Application層

#### 2-1. Query インターフェース (読み取り系の場合)

**ファイル**: `internal/features/{feature}/application/query/{feature}.go`

```go
// {Feature}Query は{feature}の読み取りクエリインターフェース
type {Feature}Query interface {
    List(ctx context.Context, limit, offset int32) ([]*entity.{Feature}, error)
    Count(ctx context.Context) (int64, error)
    // 他のクエリメソッド...
}
```

**参照ファイル**: `internal/features/field/application/query/field.go`

#### 2-2. UseCase

**ファイル**: `internal/features/{feature}/application/usecase/{action}_{feature}.go`

```go
type {Action}{Feature}Input struct {
    // 入力パラメータ
}

type {Action}{Feature}Output struct {
    // 出力データ
}

type {Action}{Feature}UseCase struct {
    query  query.{Feature}Query  // または repository
    logger *slog.Logger
}

func New{Action}{Feature}UseCase(...) *{Action}{Feature}UseCase { ... }
func (u *{Action}{Feature}UseCase) Execute(ctx context.Context, input {Action}{Feature}Input) (*{Action}{Feature}Output, error) { ... }
```

**参照ファイル**: `internal/features/field/application/usecase/list_fields.go`

### 3. Infrastructure層

#### 3-1. Query実装 (読み取り系)

**ファイル**: `internal/features/{feature}/infrastructure/query/{feature}.go`

```go
type {feature}Query struct {
    db *pgxpool.Pool
}

func New{Feature}Query(db *pgxpool.Pool) query.{Feature}Query {
    return &{feature}Query{db: db}
}

func (q *{feature}Query) List(ctx context.Context, limit, offset int32) ([]*entity.{Feature}, error) {
    queries := sqlc.New(q.db)
    rows, err := queries.List{Feature}s(ctx, &sqlc.List{Feature}sParams{...})
    // 変換処理...
}
```

**参照ファイル**: `internal/features/field/infrastructure/query/field.go`

#### 3-2. Repository実装 (書き込み系)

**ファイル**: `internal/features/{feature}/infrastructure/repository/{feature}_postgres.go`

**参照ファイル**: `internal/features/cluster/infrastructure/repository/cluster_postgres.go`

### 4. Presentation層

**ファイル**: `internal/features/{feature}/presentation/handler.go`

```go
type {Feature}Handler struct {
    {action}UC *usecase.{Action}{Feature}UseCase
    logger     *slog.Logger
}

func New{Feature}Handler(...) *{Feature}Handler { ... }

func (h *{Feature}Handler) {Action}(ctx context.Context, request openapi.{Action}RequestObject) (openapi.{Action}ResponseObject, error) {
    // 1. パラメータバリデーション
    // 2. ユースケース実行
    // 3. レスポンス変換
}
```

**参照ファイル**: `internal/features/field/presentation/handler.go`

### 5. DI設定

**ファイル**: `internal/server/router.go`

```go
// {feature}機能のDI
{feature}QueryImpl := {feature}Query.New{Feature}Query(pool)
{action}UC := {feature}Usecase.New{Action}{Feature}UseCase({feature}QueryImpl, logger)
{feature}Hdlr := {feature}Handler.New{Feature}Handler({action}UC, logger)
```

**StrictServerHandler構造体にハンドラーを追加**:
```go
type StrictServerHandler struct {
    // ...
    {feature}Handler *{feature}Handler.{Feature}Handler
}
```

**エンドポイントメソッドを実装**:
```go
func (h *StrictServerHandler) {Action}(ctx context.Context, request openapi.{Action}RequestObject) (openapi.{Action}ResponseObject, error) {
    return h.{feature}Handler.{Action}(ctx, request)
}
```

**参照ファイル**: `internal/server/router.go`

## テスト実装ガイドライン

### ファイル命名規則

| 種別 | ファイル名 |
|---|---|
| 単体テスト | `*_test.go` |
| 統合テスト | `*_integration_test.go` |
| モック | `mock_*_test.go` |

### テストカバレッジ要件

| 層 | カバレッジ |
|---|---|
| Presentation | 90%以上 |
| UseCase | 95%以上 |
| Domain | 100% |

### 必須テストケース

**正常系**:
- 有効なデータで成功

**異常系**:
- 必須パラメータ欠如
- 文字数違反
- 範囲外値
- フォーマット違反
- 不正JSON

**境界値**:
- 最小値 / 最大値
- 最小値-1 / 最大値+1

### テスト命名規則

```
Test{メソッド}_{種類}_{詳細}
```

- `Success` - 正常系
- `ValidationError_{原因}` - バリデーションエラー
- `BoundaryValue_{詳細}` - 境界値テスト
- `BusinessLogicError_{原因}` - ビジネスロジックエラー

### 参照テストファイル

| 層 | ファイル |
|---|---|
| UseCase単体 | `internal/features/field/application/usecase/list_fields_test.go` |
| Query単体 | `internal/features/field/infrastructure/query/field_test.go` |
| Query統合 | `internal/features/field/infrastructure/query/field_integration_test.go` |
| Handler単体 | `internal/features/field/presentation/handler_test.go` |

## レスポンス形式

### 成功レスポンス (一覧)

```json
{
  "data": {
    "items": [...]
  },
  "meta": {
    "pagination": {
      "total": 100,
      "page": 1,
      "pageSize": 20,
      "totalPages": 5
    }
  },
  "errors": null
}
```

### 成功レスポンス (単一)

```json
{
  "data": { ... },
  "meta": null,
  "errors": null
}
```

### エラーレスポンス

```json
{
  "data": null,
  "meta": null,
  "errors": [
    {
      "code": "error_code",
      "message": "エラーメッセージ"
    }
  ]
}
```

## 検証コマンド

実装完了後、以下のコマンドで検証する:

```bash
make test          # 全テスト実行
make lint          # Lint実行
make build         # ビルド確認
```

## 注意事項

1. **日本語でのコメント・ログ**: すべてのコメント、ログメッセージは日本語で記載
2. **エラーハンドリング**: `_` でのエラー無視は禁止、必ず `require.NoError()` でチェック
3. **機能間参照禁止**: 異なるfeatureパッケージ間で直接importしない
4. **generated編集禁止**: `internal/generated/` 配下のファイルは編集しない
