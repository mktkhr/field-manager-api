---
name: openapi-editor
description: |
  OpenAPI仕様書の編集・追加を行うスキル。
  以下のリクエストで発動:
  - エンドポイントの追加・編集
  - スキーマの追加・編集
  - API仕様の更新
  - OpenAPIファイルの作成
allowed-tools: Read, Write, Edit, Glob, Grep, Bash
---

# OpenAPI Editor Skill

このプロジェクトのOpenAPI仕様書を編集するスキルです。

## プロジェクト構成

このプロジェクトは分割構成のOpenAPI仕様書を使用しています:

```
api/
├── openapi.yaml           # ルートファイル($ref参照のみ)
├── paths/                  # エンドポイント定義
│   ├── health.yaml
│   ├── fields.yaml
│   ├── imports.yaml
│   └── clusters.yaml
└── components/
    ├── schemas/            # スキーマ定義
    │   ├── common.yaml     # Error, HealthResponse
    │   ├── field.yaml
    │   ├── import.yaml
    │   └── cluster.yaml
    ├── parameters/
    │   └── common.yaml     # 共通パラメータ
    └── responses/
        └── errors.yaml     # エラーレスポンス
```

## 必須ルール

### 1. レスポンス形式
すべてのAPIレスポンス(/healthを除く)は data/errors ラッパー形式を使用:

```yaml
ExampleResponse:
  type: object
  required:
    - data
    - errors
  properties:
    data:
      allOf:
        - $ref: "./example.yaml#/ExampleData"
      nullable: true
    errors:
      type: array
      nullable: true
      items:
        $ref: "./common.yaml#/Error"
```

### 2. 日本語コメント
すべてのdescriptionは日本語で記載

### 3. 参照パス
- schemas間の参照: `./filename.yaml#/SchemaName`
- pathsからの参照: `../components/schemas/filename.yaml#/SchemaName`
- responsesへの参照: `../components/responses/errors.yaml#/ErrorType`

## 作業手順

### エンドポイント追加時
1. `api/paths/<機能名>.yaml` を作成または編集
2. 必要なスキーマを `api/components/schemas/<機能名>.yaml` に追加
3. `api/openapi.yaml` に$ref参照を追加
4. `make api-generate` でコード生成

### スキーマ追加時
1. 適切な `api/components/schemas/*.yaml` に定義を追加
2. Data型とResponse型の両方を作成
3. 既存パターンに従う

### パラメータ追加時
1. 共通パラメータは `api/components/parameters/common.yaml` に追加
2. pathsファイルから`$ref`で参照

## パターン例

### エンドポイント定義(paths/*.yaml)

```yaml
example:
  get:
    tags:
      - example
    summary: サンプル取得
    operationId: getExample
    parameters:
      - $ref: "../components/parameters/common.yaml#/ExampleParam"
    responses:
      "200":
        description: 成功
        content:
          application/json:
            schema:
              $ref: "../components/schemas/example.yaml#/ExampleResponse"
      "400":
        $ref: "../components/responses/errors.yaml#/BadRequest"
      "500":
        $ref: "../components/responses/errors.yaml#/InternalServerError"
```

### スキーマ定義(components/schemas/*.yaml)

```yaml
ExampleData:
  type: object
  required:
    - id
    - name
  properties:
    id:
      type: string
      format: uuid
      description: ID
    name:
      type: string
      description: 名前

ExampleResponse:
  type: object
  required:
    - data
    - errors
  properties:
    data:
      allOf:
        - $ref: "./example.yaml#/ExampleData"
      nullable: true
    errors:
      type: array
      nullable: true
      items:
        $ref: "./common.yaml#/Error"
```

### 非同期処理エンドポイント(POST)

202 Acceptedを返す場合はLocationヘッダーを含める:

```yaml
example:
  post:
    responses:
      "202":
        description: リクエスト受付
        headers:
          Location:
            description: ステータス確認用URL
            schema:
              type: string
              format: uri
        content:
          application/json:
            schema:
              $ref: "../components/schemas/example.yaml#/ExampleResponse"
```

## 参照ファイル

実装時は以下のファイルを参考にしてください:

- @api/paths/clusters.yaml (エンドポイントのパターン例)
- @api/components/schemas/cluster.yaml (スキーマのパターン例)
- @api/components/responses/errors.yaml (エラーレスポンス)
- @api/components/parameters/common.yaml (パラメータ定義)
- @CLAUDE.md (プロジェクト規約)

## コード生成

OpenAPI仕様書を編集した後は必ずコード生成を実行:

```bash
make api-generate
```

生成されたコードは `internal/generated/openapi/` に出力されます。このディレクトリのファイルは直接編集しないでください。
