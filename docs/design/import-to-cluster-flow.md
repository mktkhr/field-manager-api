# インポートからクラスター生成までのフロー

wagriファイルのインポートからクラスター計算までの一連の処理フローを説明する。

## アーキテクチャ概要

```mermaid
graph TB
    subgraph "クライアント"
        Client[Webアプリ/CLI]
    end

    subgraph "API Server"
        API[REST API<br/>Gin Server]
    end

    subgraph "非同期ワーカー"
        ImportProcessor[import-processor<br/>インポート処理]
        ClusterWorker[cluster-worker<br/>クラスター計算]
    end

    subgraph "データストア"
        S3[(S3/MinIO<br/>ファイルストレージ)]
        PostgreSQL[(PostgreSQL<br/>メインDB)]
        Redis[(Redis<br/>キャッシュ)]
    end

    Client -->|1. ファイルアップロード| S3
    Client -->|2. インポートリクエスト| API
    API -->|3. ジョブ作成| PostgreSQL

    ImportProcessor -->|4. ジョブ取得| PostgreSQL
    ImportProcessor -->|5. ファイル取得| S3
    ImportProcessor -->|6. 圃場データ保存| PostgreSQL
    ImportProcessor -->|7. クラスタージョブ登録| PostgreSQL

    ClusterWorker -->|8. ジョブ取得| PostgreSQL
    ClusterWorker -->|9. クラスター計算・保存| PostgreSQL
    ClusterWorker -->|10. キャッシュ削除| Redis

    Client -->|11. クラスター取得| API
    API -->|12. キャッシュ確認| Redis
    API -->|13. DB問い合わせ| PostgreSQL
```

## 処理フロー詳細

### 1. インポートリクエスト〜ジョブ登録

```mermaid
sequenceDiagram
    autonumber
    participant Client as クライアント
    participant S3 as S3/MinIO
    participant API as API Server
    participant DB as PostgreSQL

    Client->>S3: wagriファイルアップロード
    S3-->>Client: アップロード完了(オブジェクトキー)

    Client->>API: POST /api/v1/imports<br/>{objectKey, fileName}
    API->>DB: import_jobsにレコード作成<br/>status: pending
    DB-->>API: ジョブID
    API-->>Client: 202 Accepted<br/>{jobId}
```

### 2. インポート処理(import-processor)

```mermaid
sequenceDiagram
    autonumber
    participant Worker as import-processor
    participant DB as PostgreSQL
    participant S3 as S3/MinIO

    loop ポーリング(RUN_ONCE=falseの場合)
        Worker->>DB: 未処理ジョブ取得<br/>status: pending

        alt ジョブあり
            DB-->>Worker: import_job
            Worker->>DB: status: processing に更新

            Worker->>S3: wagriファイル取得
            S3-->>Worker: ファイルデータ

            Worker->>Worker: wagriパース<br/>LinearPolygon→Polygon変換

            loop バッチ処理
                Worker->>DB: 既存圃場のH3インデックス取得<br/>(プリフェッチ)
                DB-->>Worker: 旧H3インデックス

                Worker->>Worker: 旧H3を影響セルに追加

                Worker->>Worker: H3インデックス計算<br/>(res3, res5, res7, res9)

                Worker->>DB: 圃場データUPSERT

                Worker->>Worker: 新H3を影響セルに追加
            end

            Worker->>DB: cluster_jobs作成<br/>affected_h3_cells付き
            Worker->>DB: status: completed に更新
        else ジョブなし
            DB-->>Worker: (empty)
        end
    end
```

### 3. クラスター計算(cluster-worker)

```mermaid
sequenceDiagram
    autonumber
    participant Worker as cluster-worker
    participant DB as PostgreSQL
    participant Redis as Redis

    loop ポーリング(RUN_ONCE=falseの場合)
        Worker->>DB: 未処理クラスタージョブ取得<br/>status: pending

        alt ジョブあり
            DB-->>Worker: cluster_job
            Worker->>DB: status: processing に更新

            alt affected_h3_cells が空
                Note over Worker: 全範囲再計算モード

                loop 各解像度(res3, res5, res7, res9)
                    Worker->>DB: 全圃場をH3で集計
                    DB-->>Worker: 集計結果
                    Worker->>DB: cluster_resultsに保存
                end
            else affected_h3_cells あり
                Note over Worker: 差分更新モード

                loop 各解像度(res3, res5, res7, res9)
                    Worker->>DB: 影響セルの既存結果を削除
                    Worker->>DB: 影響セルのみ再集計
                    DB-->>Worker: 集計結果
                    Worker->>DB: cluster_resultsにUPSERT
                end
            end

            Worker->>Redis: クラスターキャッシュ削除
            Worker->>DB: status: completed に更新
        else ジョブなし
            DB-->>Worker: (empty)
        end
    end
```

### 4. クラスター取得API

```mermaid
sequenceDiagram
    autonumber
    participant Client as クライアント
    participant API as API Server
    participant Redis as Redis
    participant DB as PostgreSQL

    Client->>API: GET /api/v1/clusters?zoom=X

    API->>API: zoom → resolution変換<br/>(zoom 1-6→res3, 7-10→res5, 11-14→res7, 15+→res9)

    API->>Redis: キャッシュ確認

    alt キャッシュヒット
        Redis-->>API: クラスターデータ
    else キャッシュミス
        API->>DB: cluster_results取得<br/>WHERE resolution = X
        DB-->>API: クラスターデータ
        API->>Redis: キャッシュ保存
    end

    API-->>Client: 200 OK<br/>[{h3Index, fieldCount, center}]
```

## H3インデックスと解像度

### 解像度とズームレベルの対応

| ズームレベル | H3解像度 | セル面積(概算) | 用途 |
|-------------|---------|---------------|------|
| 1-6 | res3 | ~12,000 km² | 国/地方レベル |
| 7-10 | res5 | ~250 km² | 都道府県レベル |
| 11-14 | res7 | ~5 km² | 市区町村レベル |
| 15+ | res9 | ~0.1 km² | 地区レベル |

### H3インデックス計算フロー

```mermaid
flowchart TD
    A[圃場ポリゴン] --> B[重心座標計算]
    B --> C[緯度/経度取得]
    C --> D[h3.LatLngToCell]

    D --> E1[res3インデックス]
    D --> E2[res5インデックス]
    D --> E3[res7インデックス]
    D --> E4[res9インデックス]

    E1 --> F[fieldsテーブルに保存]
    E2 --> F
    E3 --> F
    E4 --> F
```

## 差分更新の仕組み

### 影響セル収集

```mermaid
flowchart TD
    subgraph "インポート処理"
        A[バッチ処理開始] --> B{既存圃場?}
        B -->|Yes| C[旧H3インデックス取得]
        B -->|No| D[新規圃場]
        C --> E[旧H3を影響セットに追加]
        D --> F[H3インデックス計算]
        E --> F
        F --> G[圃場UPSERT]
        G --> H[新H3を影響セットに追加]
        H --> I{次のバッチ?}
        I -->|Yes| A
        I -->|No| J[影響セル確定]
    end

    J --> K[cluster_job作成<br/>affected_h3_cells付き]
```

### 差分計算 vs 全範囲計算

```mermaid
flowchart TD
    A[cluster_job取得] --> B{affected_h3_cells<br/>が空?}

    B -->|Yes| C[全範囲再計算]
    B -->|No| D[差分更新]

    subgraph "全範囲再計算"
        C --> C1[全解像度で集計]
        C1 --> C2[cluster_results全更新]
    end

    subgraph "差分更新"
        D --> D1[影響セルを解像度別に分類]
        D1 --> D2[影響セルの既存結果削除]
        D2 --> D3[影響セルのみ再集計]
        D3 --> D4[結果をUPSERT]
    end

    C2 --> E[キャッシュクリア]
    D4 --> E
```

## トランザクション管理

### トランザクション境界の概要

```mermaid
flowchart TB
    subgraph "インポート処理"
        direction TB
        I1[バッチ1] --> I2[バッチ2] --> I3[バッチN]
        I1 -.- T1[TX1]
        I2 -.- T2[TX2]
        I3 -.- T3[TXN]
    end

    subgraph "クラスター計算"
        direction TB
        C1[解像度3] --> C2[解像度5] --> C3[解像度7] --> C4[解像度9]
        C1 -.- CT1[TX1]
        C2 -.- CT2[TX2]
        C3 -.- CT3[TX3]
        C4 -.- CT4[TX4]
    end

    I3 --> CJ[cluster_job作成]
    CJ --> C1
```

### 処理別トランザクション詳細

| 処理 | トランザクション | 範囲 | 失敗時の挙動 |
|-----|----------------|-----|------------|
| 圃場バッチUPSERT | あり | 1バッチ(複数圃場) | バッチ全体ロールバック、他バッチは継続 |
| クラスター結果保存 | あり | 1解像度分の全クラスター | 解像度全体ロールバック |
| ジョブステータス更新 | なし(単一クエリ) | 1レコード | 自動コミット |
| H3集計クエリ | なし(単一クエリ) | - | 自動コミット |
| キャッシュ削除 | なし(Redis操作) | - | ログ出力のみ、処理継続 |

### 圃場バッチUPSERT(field.go:68-205)

```mermaid
sequenceDiagram
    participant UC as UseCase
    participant Repo as FieldRepository
    participant DB as PostgreSQL

    UC->>Repo: UpsertBatch(inputs)

    rect rgb(200, 230, 200)
        Note over Repo,DB: トランザクション開始
        Repo->>DB: BEGIN

        loop 各圃場
            Repo->>DB: 土壌タイプUPSERT
            Repo->>DB: マスタデータUPSERT
            Repo->>DB: 圃場UPSERT
            Repo->>DB: 農地台帳DELETE/INSERT
        end

        alt 全操作成功
            Repo->>DB: COMMIT
        else エラー発生
            Repo->>DB: ROLLBACK
            Repo-->>UC: error
        end
        Note over Repo,DB: トランザクション終了
    end

    Repo-->>UC: nil
```

**特徴**:
- 1バッチ内の全圃場を1トランザクションで処理
- バッチ失敗時は該当バッチのみロールバック、他バッチの処理は継続
- 部分的成功を許容(partially_completed ステータス)

### クラスター結果保存(cluster_postgres.go:55-90)

```mermaid
sequenceDiagram
    participant UC as CalculateClustersUseCase
    participant Repo as ClusterRepository
    participant DB as PostgreSQL

    UC->>Repo: SaveClusters(clusters)

    rect rgb(200, 230, 200)
        Note over Repo,DB: トランザクション開始
        Repo->>DB: BEGIN

        loop 各クラスター
            Repo->>DB: UPSERT cluster_result
        end

        alt 全操作成功
            Repo->>DB: COMMIT
        else エラー発生
            Repo->>DB: ROLLBACK
            Repo-->>UC: error
        end
        Note over Repo,DB: トランザクション終了
    end

    UC->>UC: 次の解像度へ
```

**特徴**:
- 1解像度の全クラスター結果を1トランザクションで保存
- 解像度ごとに独立したトランザクション
- 途中の解像度で失敗した場合、それ以前の解像度の結果は保持される

### インポート処理全体のトランザクション設計

```mermaid
flowchart TD
    subgraph "ユースケースレベル(トランザクションなし)"
        A[インポート開始] --> B[バッチ1処理]
        B --> C[バッチ2処理]
        C --> D[...]
        D --> E[バッチN処理]
        E --> F[進捗更新]
        F --> G[ステータス更新]
        G --> H[クラスタージョブ作成]
    end

    subgraph "TX1"
        B1[圃場UPSERT]
    end
    subgraph "TX2"
        C1[圃場UPSERT]
    end
    subgraph "TXN"
        E1[圃場UPSERT]
    end

    B --> B1
    C --> C1
    E --> E1

    style A fill:#e0e0e0
    style F fill:#e0e0e0
    style G fill:#e0e0e0
    style H fill:#e0e0e0
```

**設計理由**:
- 大量データ処理時のメモリ効率とリカバリ性を優先
- バッチ単位での部分的成功を許容
- 失敗バッチのリトライが可能な設計

### エラー発生時のデータ整合性

| シナリオ | 影響範囲 | リカバリ方法 |
|---------|---------|------------|
| バッチ処理中にDB接続断 | 該当バッチのみ未反映 | ジョブ再実行(べき等性あり) |
| クラスター計算中に障害 | 一部解像度のみ更新 | 手動再計算API実行 |
| キャッシュ削除失敗 | キャッシュと実データの不整合 | TTL経過で自動解消 or 手動再計算 |

## データベーステーブル

### 関連テーブル構成

```mermaid
erDiagram
    import_jobs ||--o{ fields : "creates"
    import_jobs ||--o| cluster_jobs : "triggers"
    fields ||--o{ cluster_results : "aggregates to"

    import_jobs {
        uuid id PK
        varchar status
        varchar object_key
        varchar file_name
        timestamp created_at
        timestamp updated_at
    }

    fields {
        uuid id PK
        geometry polygon
        varchar h3_index_res3
        varchar h3_index_res5
        varchar h3_index_res7
        varchar h3_index_res9
        timestamp created_at
        timestamp updated_at
    }

    cluster_jobs {
        uuid id PK
        varchar status
        int priority
        text_array affected_h3_cells
        timestamp created_at
        timestamp updated_at
    }

    cluster_results {
        uuid id PK
        int resolution
        varchar h3_index
        int field_count
        geometry center
        timestamp created_at
        timestamp updated_at
    }
```

## ワーカーの動作モード

### import-processor / cluster-worker 共通

| 環境変数 | デフォルト | 説明 |
|---------|-----------|------|
| `RUN_ONCE` | `false` | `true`: 1回実行して終了(Lambda/K8s Job向け) |
| `BATCH_SIZE` | `10` | 1回のポーリングで処理するジョブ数 |
| `POLL_INTERVAL` | `60s` | ポーリング間隔(デーモンモード時) |

```mermaid
flowchart TD
    A[ワーカー起動] --> B{RUN_ONCE?}

    B -->|true| C[ジョブ処理1回実行]
    C --> D[終了]

    B -->|false| E[デーモンモード]
    E --> F[POLL_INTERVAL待機]
    F --> G[ジョブ処理実行]
    G --> H{シグナル受信?}
    H -->|No| F
    H -->|Yes| I[グレースフル終了]
```

## 手動再計算API

インポートとは別に、管理者が手動で全範囲再計算を実行できる。

```mermaid
sequenceDiagram
    autonumber
    participant Admin as 管理者
    participant API as API Server
    participant DB as PostgreSQL

    Admin->>API: POST /api/v1/clusters/recalculate
    API->>DB: cluster_jobs作成<br/>affected_h3_cells: NULL<br/>priority: 10(高優先度)
    DB-->>API: ジョブID
    API-->>Admin: 202 Accepted

    Note over DB: cluster-workerが<br/>全範囲再計算を実行
```
