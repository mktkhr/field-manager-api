# クラスター計算機能 動作確認手順

本ドキュメントでは、H3クラスタリング計算機能のローカル環境での動作確認手順を説明します。

## 概要

クラスター計算機能は以下のコンポーネントで構成されています:

```
[ジョブエンキュー] → [cluster_jobs] → [cluster-worker] → [cluster_results] → [Redis Cache]
       ↑                                                           ↓
   - 手動API                                              [GET /api/v1/clusters]
   - インポート完了時
```

| コンポーネント   | 役割                                           |
| ---------------- | ---------------------------------------------- |
| 手動API          | POST /api/v1/clusters/recalculate でジョブ登録 |
| import-processor | インポート完了時に自動でジョブ登録             |
| cluster_jobs     | ジョブキューテーブル(PostgreSQL)               |
| cluster-worker   | ジョブを取得し、クラスター計算を実行           |
| cluster_results  | 計算結果テーブル(PostgreSQL)                   |
| Redis Cache      | 計算結果のキャッシュ                           |

### 処理フロー

1. **エンキュー**: 手動API or インポート完了時に`cluster_jobs`テーブルにジョブ登録
2. **ジョブ取得**: `cluster-worker`が`pending`状態のジョブを取得
3. **計算実行**: `fields`テーブルをH3インデックスで集計(解像度3, 5, 7, 9)
4. **結果保存**: `cluster_results`テーブルにUPSERT
5. **キャッシュクリア**: Redisキャッシュを削除(次回API呼び出し時に再キャッシュ)

---

## 前提条件

### 必要なサービス

```bash
# Docker Composeで起動
docker compose -f docker/compose.yaml up -d postgres valkey
```

### 環境変数

`.env`ファイルに以下が設定されていること:

```bash
# Database
DB_HOST=localhost
DB_PORT=5435
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=field_manager_db
DB_SSL_MODE=disable

# Cache(Valkey/Redis)
CACHE_HOST=localhost
CACHE_PORT=6379
CACHE_PASSWORD=
CACHE_DB=0
```

### マイグレーション実行

```bash
make migrate-up
```

---

## 1. ジョブエンキュー

### 1.1 手動API経由

APIサーバーを起動:
```bash
make run
```

エンドポイント呼び出し:
```bash
curl -X POST http://localhost:8080/api/v1/clusters/recalculate
```

期待されるレスポンス:
```json
{
  "message": "クラスター再計算ジョブをエンキューしました",
  "enqueued": true
}
```

既にジョブが存在する場合(409):
```json
{
  "code": "already_running",
  "message": "既にクラスター再計算ジョブが実行中です"
}
```

### 1.2 インポート完了時の自動エンキュー

`import-processor`がインポート処理を完了すると、自動的にクラスタージョブがエンキューされます。

詳細は[wagri-import.md](./wagri-import.md)を参照してください。

### 1.3 直接DBへ登録(テスト用)

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
INSERT INTO cluster_jobs (id, status, priority, created_at)
VALUES ('00000000-0000-0000-0000-000000000001', 'pending', 10, NOW());
"
```

---

## 2. cluster-workerの動作確認

### 2.1 Dockerイメージのビルド

```bash
make cluster-worker-build
```

### 2.2 1回実行モード(Lambda/K8s Job向け)

Docker経由:
```bash
make cluster-worker-run
```

または直接実行:
```bash
source .env && RUN_ONCE=true go run ./cmd/cluster-worker
```

期待される出力:
```json
{"level":"INFO","msg":"クラスターワーカーを起動しています..."}
{"level":"INFO","msg":"1回実行モードで起動します"}
{"level":"INFO","msg":"ジョブ処理を開始します","batch_size":10}
{"level":"INFO","msg":"処理対象のジョブを取得しました","job_count":1}
{"level":"INFO","msg":"ジョブの処理を開始します","job_id":"..."}
{"level":"INFO","msg":"クラスター計算を開始します"}
{"level":"INFO","msg":"解像度別のクラスター計算を開始します","resolution":"res3"}
{"level":"INFO","msg":"解像度別のクラスター計算が完了しました","resolution":"res3","cluster_count":5}
...
{"level":"INFO","msg":"クラスター計算が完了しました"}
{"level":"INFO","msg":"ジョブの処理が完了しました","job_id":"..."}
{"level":"INFO","msg":"1回実行モードが完了しました"}
```

### 2.3 デーモンモード(ポーリング)

Docker経由:
```bash
make cluster-worker-daemon
```

または直接実行:
```bash
source .env && POLL_INTERVAL=30s go run ./cmd/cluster-worker
```

| 環境変数        | 説明                       | デフォルト |
| --------------- | -------------------------- | ---------- |
| `RUN_ONCE`      | 1回実行で終了              | false      |
| `BATCH_SIZE`    | 1回に処理するジョブ数      | 10         |
| `POLL_INTERVAL` | ポーリング間隔(例: 30s, 1m) | 60s        |

停止はCtrl+C(SIGINT/SIGTERM)

---

## 3. 結果確認

### 3.1 cluster_jobsテーブル

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
SELECT id, status, priority, created_at, started_at, completed_at, error_message
FROM cluster_jobs
ORDER BY created_at DESC
LIMIT 10;
"
```

期待される結果:
```
                  id                  |  status   | priority |          created_at          |          started_at          |         completed_at
--------------------------------------+-----------+----------+------------------------------+------------------------------+------------------------------
 00000000-0000-0000-0000-000000000001 | completed |       10 | 2024-01-01 00:00:00+00       | 2024-01-01 00:00:01+00       | 2024-01-01 00:00:05+00
```

| status     | 説明                     |
| ---------- | ------------------------ |
| pending    | 待機中(未処理)           |
| processing | 処理中                   |
| completed  | 完了                     |
| failed     | 失敗(error_messageあり)  |

### 3.2 cluster_resultsテーブル

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
SELECT resolution, h3_index, field_count, center_lat, center_lng, calculated_at
FROM cluster_results
ORDER BY resolution, field_count DESC
LIMIT 20;
"
```

期待される結果:
```
 resolution |    h3_index     | field_count | center_lat  | center_lng  |       calculated_at
------------+-----------------+-------------+-------------+-------------+----------------------------
          3 | 831f8ffffffffff |        1234 |  35.6812405 | 139.7671248 | 2024-01-01 00:00:05+00
          5 | 852e638bfffffff |         567 |  35.6762101 | 139.7689752 | 2024-01-01 00:00:05+00
          7 | 872e638b3ffffff |         123 |  35.6795032 | 139.7695123 | 2024-01-01 00:00:05+00
          9 | 892e638b307ffff |          45 |  35.6801234 | 139.7698456 | 2024-01-01 00:00:05+00
```

### 3.3 解像度別の集計確認

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
SELECT resolution, COUNT(*) as cluster_count, SUM(field_count) as total_fields
FROM cluster_results
GROUP BY resolution
ORDER BY resolution;
"
```

---

## 4. クラスター取得API

### 4.1 APIサーバー起動

```bash
make run
```

### 4.2 クラスター取得

```bash
curl "http://localhost:8080/api/v1/clusters?zoom=5&sw_lat=30&sw_lng=125&ne_lat=45&ne_lng=150"
```

期待されるレスポンス:
```json
{
  "clusters": [
    {
      "h3Index": "852e638bfffffff",
      "lat": 35.6762101,
      "lng": 139.7689752,
      "count": 567
    }
  ],
  "isStale": false
}
```

| パラメータ | 説明                 | 範囲        |
| ---------- | -------------------- | ----------- |
| zoom       | ズームレベル         | 1.0 - 22.0  |
| sw_lat     | 南西端の緯度         | -90 - 90    |
| sw_lng     | 南西端の経度         | -180 - 180  |
| ne_lat     | 北東端の緯度         | -90 - 90    |
| ne_lng     | 北東端の経度         | -180 - 180  |

### 4.3 ズームレベルと解像度の対応

| ズームレベル | H3解像度 |
| ------------ | -------- |
| < 6.0        | res3     |
| 6.0 - < 10.0 | res5     |
| 10.0 - < 14.0| res7     |
| 14.0 - 22.0  | res9     |

---

## 5. クリーンアップ

### テストデータ削除

```bash
# cluster_jobsテーブル
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
DELETE FROM cluster_jobs WHERE id = '00000000-0000-0000-0000-000000000001';
"

# cluster_resultsテーブル(全削除)
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
TRUNCATE cluster_results;
"
```

### Redisキャッシュクリア

```bash
docker compose -f docker/compose.yaml exec valkey redis-cli FLUSHDB
```

---

## トラブルシューティング

### ジョブがpendingのまま処理されない

**原因**: cluster-workerが起動していない
**対処**: `RUN_ONCE=true go run ./cmd/cluster-worker` でワーカーを起動

### ジョブがfailedになる

```bash
docker compose -f docker/compose.yaml exec postgres psql -U postgres -d field_manager_db -c "
SELECT id, error_message FROM cluster_jobs WHERE status = 'failed';
"
```

**原因例**:
- DB接続エラー: 環境変数を確認
- fieldsテーブルにデータがない: インポートを先に実行

### cluster_resultsが0件

**原因**: fieldsテーブルにデータがない
**対処**: [wagri-import.md](./wagri-import.md)に従ってインポートを実行

### APIで常にisStale=true

**原因**: cluster_resultsの`calculated_at`が古い(30分以上前)
**対処**: クラスター再計算を実行

```bash
curl -X POST http://localhost:8080/api/v1/clusters/recalculate
RUN_ONCE=true go run ./cmd/cluster-worker
```

### Redis接続エラー

```
failed to connect to Redis: dial tcp [::1]:6379: connect: connection refused
```

**原因**: Valkey/Redisが起動していない
**対処**: `docker compose -f docker/compose.yaml up -d valkey`

---

## 関連コマンド一覧

| コマンド                                             | 説明                           |
| ---------------------------------------------------- | ------------------------------ |
| `make cluster-worker-build`                          | Dockerイメージビルド           |
| `make cluster-worker-run`                            | 1回実行モード(Docker)          |
| `make cluster-worker-daemon`                         | デーモンモード(Docker)         |
| `source .env && go run ./cmd/cluster-worker`         | 直接実行                       |
| `curl -X POST .../api/v1/clusters/recalculate`       | 手動ジョブエンキュー           |
| `curl ".../api/v1/clusters?zoom=5&sw_lat=..."`       | クラスター取得                 |
