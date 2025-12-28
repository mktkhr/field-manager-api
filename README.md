# field-manager-api

Go + Gin + oapi-codegenを使用した圃場管理APIサーバー。

## データベース

### 技術スタック

- PostgreSQL 17 + PostGIS 3.5(RDS/Aurora対応)
- golang-migrate(マイグレーション管理)
- SQLC(型安全なクエリ生成)
- uber/h3-go(H3インデックス計算 - アプリ側)

### マイグレーションコマンド

| コマンド | 説明 |
|----------|------|
| `make migrate-install` | golang-migrateのインストール |
| `make migrate-create NAME=xxx` | 新規マイグレーション作成 |
| `make migrate-up` | 全マイグレーション適用 |
| `make migrate-up-one` | 1つ次のマイグレーション適用 |
| `make migrate-down` | 1つ前にロールバック |
| `make migrate-down-all` | 全ロールバック |
| `make migrate-version` | 現在のバージョン確認 |
| `make migrate-force VERSION=xxx` | バージョン強制設定(障害復旧用) |

### マイグレーション手順

#### 初回セットアップ

```bash
# 1. golang-migrateをインストール
make migrate-install

# 2. DBコンテナを起動
cd docker && docker compose up -d postgres

# 3. マイグレーションを適用
make migrate-up
```

#### 新規マイグレーション追加

```bash
# 1. マイグレーションファイルを生成
make migrate-create NAME=add_xxx_table

# 2. 生成された.up.sqlと.down.sqlを編集
#    db/migrations/NNNNNN_add_xxx_table.up.sql
#    db/migrations/NNNNNN_add_xxx_table.down.sql

# 3. マイグレーションを適用
make migrate-up

# 4. SQLCでクエリを再生成(必要に応じて)
make sqlc-generate
```

#### ロールバック

```bash
# 現在のバージョンを確認
make migrate-version

# 1つ前に戻る
make migrate-down

# 問題があればバージョンを強制設定
make migrate-force VERSION=N
```

### テーブル構成

| テーブル | 説明 |
|----------|------|
| soil_types | 土壌マスタ(大/中/小分類の階層構造) |
| fields | 圃場メイン(PostGIS geometry + H3インデックス) |
| field_land_registries | 農地台帳情報 |
| field_divisions | 分筆履歴 |
| field_mergers | 合筆履歴 |
| field_overlaps | オーバーラップ検知記録 |
