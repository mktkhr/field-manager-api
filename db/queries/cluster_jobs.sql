-- name: CreateClusterJob :one
-- クラスタージョブを作成
INSERT INTO cluster_jobs (
    id,
    status,
    priority,
    created_at
)
VALUES ($1, 'pending', $2, NOW())
RETURNING *;

-- name: GetClusterJob :one
-- クラスタージョブをIDで取得
SELECT
    id,
    status,
    priority,
    created_at,
    started_at,
    completed_at,
    error_message
FROM cluster_jobs
WHERE id = $1;

-- name: GetPendingClusterJobs :many
-- 保留中のジョブを優先度順に取得(排他ロック)
SELECT
    id,
    status,
    priority,
    created_at,
    started_at,
    completed_at,
    error_message
FROM cluster_jobs
WHERE status = 'pending'
ORDER BY priority DESC, created_at
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: UpdateClusterJobToProcessing :exec
-- ジョブを処理中に更新
UPDATE cluster_jobs
SET
    status = 'processing',
    started_at = NOW()
WHERE id = $1;

-- name: UpdateClusterJobToCompleted :exec
-- ジョブを完了に更新
UPDATE cluster_jobs
SET
    status = 'completed',
    completed_at = NOW()
WHERE id = $1;

-- name: UpdateClusterJobToFailed :exec
-- ジョブを失敗に更新
UPDATE cluster_jobs
SET
    status = 'failed',
    completed_at = NOW(),
    error_message = $2
WHERE id = $1;

-- name: HasPendingOrProcessingJob :one
-- 保留中または処理中のジョブがあるか確認
SELECT EXISTS(
    SELECT 1 FROM cluster_jobs
    WHERE status IN ('pending', 'processing')
) AS has_job;

-- name: DeleteOldCompletedJobs :exec
-- 7日以上前に完了したジョブを削除
DELETE FROM cluster_jobs
WHERE status = 'completed' AND completed_at < NOW() - INTERVAL '7 days';

-- name: DeleteOldFailedJobs :exec
-- 30日以上前に失敗したジョブを削除
DELETE FROM cluster_jobs
WHERE status = 'failed' AND completed_at < NOW() - INTERVAL '30 days';

-- name: CreateClusterJobWithAffectedCells :one
-- 影響セル情報付きでクラスタージョブを作成
INSERT INTO cluster_jobs (
    id,
    status,
    priority,
    affected_h3_cells,
    created_at
)
VALUES ($1, 'pending', $2, $3, NOW())
RETURNING *;

-- name: GetPendingClusterJobsWithAffectedCells :many
-- 保留中のジョブを影響セル情報付きで優先度順に取得(排他ロック)
SELECT
    id,
    status,
    priority,
    affected_h3_cells,
    created_at,
    started_at,
    completed_at,
    error_message
FROM cluster_jobs
WHERE status = 'pending'
ORDER BY priority DESC, created_at
LIMIT $1
FOR UPDATE SKIP LOCKED;
