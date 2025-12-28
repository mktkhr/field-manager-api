-- name: GetIdleLandStatus :one
-- 遊休農地状況をコードで取得
SELECT
    code,
    name,
    description
FROM idle_land_statuses
WHERE code = $1;

-- name: ListIdleLandStatuses :many
-- 遊休農地状況一覧を取得
SELECT
    code,
    name,
    description
FROM idle_land_statuses
ORDER BY code;

-- name: UpsertIdleLandStatus :one
-- 遊休農地状況をUPSERT
INSERT INTO idle_land_statuses (
    code,
    name
) VALUES (
    $1, $2
)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name
RETURNING *;
