-- name: GetImportJob :one
-- インポートジョブをIDで取得
SELECT
    id,
    city_code,
    status,
    total_records,
    processed_records,
    failed_records,
    last_processed_batch,
    s3_key,
    execution_arn,
    error_message,
    failed_record_ids,
    created_at,
    started_at,
    completed_at
FROM import_jobs
WHERE id = $1;

-- name: ListImportJobs :many
-- インポートジョブ一覧を取得
SELECT
    id,
    city_code,
    status,
    total_records,
    processed_records,
    failed_records,
    last_processed_batch,
    s3_key,
    execution_arn,
    error_message,
    failed_record_ids,
    created_at,
    started_at,
    completed_at
FROM import_jobs
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: ListImportJobsByCityCode :many
-- 市区町村コードでインポートジョブ一覧を取得
SELECT
    id,
    city_code,
    status,
    total_records,
    processed_records,
    failed_records,
    last_processed_batch,
    s3_key,
    execution_arn,
    error_message,
    failed_record_ids,
    created_at,
    started_at,
    completed_at
FROM import_jobs
WHERE city_code = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: CreateImportJob :one
-- インポートジョブを作成
INSERT INTO import_jobs (
    city_code,
    status
) VALUES (
    $1, 'pending'
) RETURNING *;

-- name: UpdateImportJobStatus :one
-- インポートジョブのステータスを更新
UPDATE import_jobs
SET
    status = sqlc.arg(status)::VARCHAR,
    started_at = CASE WHEN sqlc.arg(status)::VARCHAR = 'processing' AND started_at IS NULL THEN NOW() ELSE started_at END,
    completed_at = CASE WHEN sqlc.arg(status)::VARCHAR IN ('completed', 'failed', 'partially_completed') THEN NOW() ELSE completed_at END
WHERE id = $1
RETURNING *;

-- name: UpdateImportJobProgress :one
-- インポートジョブの進捗を更新
UPDATE import_jobs
SET
    processed_records = $2,
    failed_records = $3,
    last_processed_batch = $4
WHERE id = $1
RETURNING *;

-- name: UpdateImportJobS3Key :one
-- インポートジョブのS3キーを更新
UPDATE import_jobs
SET
    s3_key = $2
WHERE id = $1
RETURNING *;

-- name: UpdateImportJobExecutionArn :one
-- インポートジョブの実行ARNを更新
UPDATE import_jobs
SET
    execution_arn = $2
WHERE id = $1
RETURNING *;

-- name: UpdateImportJobTotalRecords :one
-- インポートジョブの総レコード数を更新
UPDATE import_jobs
SET
    total_records = $2
WHERE id = $1
RETURNING *;

-- name: UpdateImportJobError :one
-- インポートジョブのエラー情報を更新
UPDATE import_jobs
SET
    status = 'failed',
    error_message = $2,
    failed_record_ids = $3,
    completed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountImportJobs :one
-- インポートジョブの総数を取得
SELECT COUNT(*) FROM import_jobs;

-- name: CountImportJobsByStatus :one
-- ステータス別のインポートジョブ数を取得
SELECT COUNT(*) FROM import_jobs WHERE status = $1;
