# インポート機能 動作確認手順

本ドキュメントでは、Wagriデータインポート機能のローカル環境での動作確認手順を説明します。

## 概要

インポート機能は以下のコンポーネントで構成されています:

```
[API] → [Step Functions] → [Lambda: wagri-fetcher] → [RustFS(S3)] → [import-processor] → [PostgreSQL]
```

| コンポーネント   | 役割                             |
| ---------------- | -------------------------------- |
| Step Functions   | ワークフロー管理(LocalStack)     |
| wagri-fetcher    | Wagri APIからデータ取得、S3保存  |
| RustFS           | S3互換ストレージ(ローカル開発用) |
| import-processor | S3からデータ読み取り、DB登録     |

## 前提条件

### 必要なサービス

```bash
# Docker Composeで起動
docker compose -f docker/compose.yaml up -d postgres localstack rustfs
```

### 環境変数

`.env`ファイルに以下が設定されていること:

```bash
# Storage(RustFS)
STORAGE_S3_ENABLED=false
STORAGE_ENDPOINT=http://localhost:9000
STORAGE_BUCKET=field-manager
STORAGE_ACCESS_KEY_ID=rustfsadmin
STORAGE_SECRET_ACCESS_KEY=rustfsadmin
STORAGE_REGION=ap-northeast-1

# Wagri API(OAuth2)
WAGRI_BASE_URL=https://api.wagri2.net
WAGRI_CLIENT_ID=your-client-id
WAGRI_CLIENT_SECRET=your-client-secret

# Database
DB_HOST=localhost
DB_PORT=5435
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=field_manager_db
```

---

## 1. RustFSバケット作成

RustFSにバケットを作成(初回のみ):

```bash
AWS_ACCESS_KEY_ID=rustfsadmin AWS_SECRET_ACCESS_KEY=rustfsadmin \
aws --endpoint-url http://localhost:9000 s3 mb s3://field-manager
```

確認:
```bash
AWS_ACCESS_KEY_ID=rustfsadmin AWS_SECRET_ACCESS_KEY=rustfsadmin \
aws --endpoint-url http://localhost:9000 s3 ls
```

RustFS UIでも確認可能: http://localhost:9003

---

## 2. Lambda(wagri-fetcher)の動作確認

### 2.1 Lambdaのビルド&デプロイ

```bash
make localstack-deploy-lambda
```

このコマンドで以下が実行されます:
- Lambdaバイナリのビルド(Linux/amd64)
- LocalStackへのデプロイ
- 環境変数の設定(RustFS接続情報、Wagri認証情報)

### 2.2 Lambdaのテスト実行

```bash
make localstack-invoke-lambda
```

期待される結果:
- Wagri APIへのOAuth2認証
- 圃場データの取得
- RustFSへのJSON保存

**注意**: Wagri APIの認証情報が正しくない場合、401エラーが発生します。

### 2.3 Step Functionsワークフロー実行

```bash
make localstack-start-workflow
```

実行履歴の確認:
```bash
make localstack-list-executions
```

---

## 3. import-processorの動作確認

Wagri APIが利用できない場合でも、テストJSONを使用して動作確認できます。

### 3.1 テストJSONの作成

```bash
cat > /tmp/test-import.json << 'EOF'
{
  "targetFeatures": [
    {
      "type": "Feature",
      "geometry": {
        "type": "LinearPolygon",
        "coordinates": [[[136.123, 35.456], [136.124, 35.456], [136.124, 35.457], [136.123, 35.457], [136.123, 35.456]]]
      },
      "properties": {
        "ID": "11111111-1111-1111-1111-111111111111",
        "CityCode": "163210",
        "IssueYear": "2024",
        "EditYear": "2024",
        "PointLat": 35.4565,
        "PointLng": 136.1235,
        "FieldType": "田",
        "Number": 1,
        "SoilLargeCode": "01",
        "SoilMiddleCode": "01",
        "SoilSmallCode": "001",
        "SoilSmallName": "灰色低地土",
        "History": "{}",
        "LastPolygonUuid": "00000000-0000-0000-0000-000000000001",
        "PrevLastPolygonUuid": null,
        "PinInfo": []
      }
    }
  ]
}
EOF
```

**注意**: `ID`フィールドはUUID形式である必要があります。

### 3.2 RustFSへアップロード

```bash
AWS_ACCESS_KEY_ID=rustfsadmin AWS_SECRET_ACCESS_KEY=rustfsadmin \
aws --endpoint-url http://localhost:9000 s3 cp /tmp/test-import.json s3://field-manager/imports/163210/test-import.json
```

確認:
```bash
AWS_ACCESS_KEY_ID=rustfsadmin AWS_SECRET_ACCESS_KEY=rustfsadmin \
aws --endpoint-url http://localhost:9000 s3 ls s3://field-manager/imports/163210/
```

### 3.3 import_jobレコード作成

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
INSERT INTO import_jobs (id, city_code, status, s3_key, created_at, started_at)
VALUES ('00000000-0000-0000-0000-000000000001', '163210', 'processing', 'imports/163210/test-import.json', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET status = 'processing', processed_records = 0, failed_records = 0, started_at = NOW();
"
```

### 3.4 Dockerイメージのビルド

```bash
make import-processor-build
```

### 3.5 import-processor実行

```bash
make import-processor-run \
  S3_KEY=imports/163210/test-import.json \
  IMPORT_JOB_ID=00000000-0000-0000-0000-000000000001
```

期待される出力:
```json
{"level":"INFO","msg":"import-processor開始",...}
{"level":"INFO","msg":"インポート処理を開始",...}
{"level":"INFO","msg":"インポート処理が完了","processed":1,"failed":0,"status":"completed"}
```

---

## 4. 結果確認

### 4.1 fieldsテーブル確認

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
SELECT id, city_code, name, h3_index_res5, ST_AsText(centroid) as centroid
FROM fields
ORDER BY created_at DESC
LIMIT 10;
"
```

期待される結果:
```
                  id                  | city_code |   name   |  h3_index_res5  |       centroid
--------------------------------------+-----------+----------+-----------------+---------------------
 11111111-1111-1111-1111-111111111111 | 163210    | 名称不明 | 852e638bfffffff | POINT(136.12 35.46)
```

### 4.2 import_jobsテーブル確認

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
SELECT id, city_code, status, total_records, processed_records, failed_records
FROM import_jobs
WHERE id = '00000000-0000-0000-0000-000000000001';
"
```

期待される結果:
```
                  id                  | city_code |  status   | total_records | processed_records | failed_records
--------------------------------------+-----------+-----------+---------------+-------------------+----------------
 00000000-0000-0000-0000-000000000001 | 163210    | completed |             1 |                 1 |              0
```

---

## 5. クリーンアップ

### テストデータ削除

```bash
# fieldsテーブル
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
DELETE FROM fields WHERE id IN ('11111111-1111-1111-1111-111111111111');
"

# import_jobsテーブル
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
DELETE FROM import_jobs WHERE id = '00000000-0000-0000-0000-000000000001';
"

# RustFS
AWS_ACCESS_KEY_ID=rustfsadmin AWS_SECRET_ACCESS_KEY=rustfsadmin \
aws --endpoint-url http://localhost:9000 s3 rm s3://field-manager/imports/163210/test-import.json
```

---

## トラブルシューティング

### Lambda実行時に401エラー

```
{"errorMessage":"wagri API呼び出しに失敗: APIエラー: ステータスコード 401"}
```

**原因**: Wagri APIの認証情報が正しくない
**対処**: `.env`の`WAGRI_CLIENT_ID`と`WAGRI_CLIENT_SECRET`を確認

### import-processor実行時にUUIDエラー

```
{"error":"圃場ID変換失敗: invalid UUID length: 14"}
```

**原因**: JSONの`ID`フィールドがUUID形式でない
**対処**: `ID`をUUID形式(例: `11111111-1111-1111-1111-111111111111`)に修正

### RustFSへの接続エラー

```
AccessDenied: Access Denied
```

**原因**: 認証情報が正しくない
**対処**: `AWS_ACCESS_KEY_ID`と`AWS_SECRET_ACCESS_KEY`を確認

### NoSuchBucketエラー

```
NoSuchBucket: The specified bucket does not exist
```

**原因**: バケットが作成されていない
**対処**: `aws s3 mb s3://field-manager`でバケット作成

---

## 関連コマンド一覧

| コマンド                          | 説明                             |
| --------------------------------- | -------------------------------- |
| `make localstack-up`              | LocalStack起動                   |
| `make localstack-status`          | LocalStackステータス確認         |
| `make localstack-build-lambda`    | Lambdaビルド                     |
| `make localstack-deploy-lambda`   | Lambdaデプロイ(環境変数設定含む) |
| `make localstack-invoke-lambda`   | Lambdaテスト実行                 |
| `make localstack-start-workflow`  | Step Functionsワークフロー実行   |
| `make localstack-list-executions` | Step Functions実行履歴           |
| `make import-processor-build`     | import-processorビルド           |
| `make import-processor-run`       | import-processor実行             |
