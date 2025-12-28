-- name: GetLandCategory :one
-- 土地種別をコードで取得
SELECT
    code,
    name,
    description
FROM land_categories
WHERE code = $1;

-- name: ListLandCategories :many
-- 土地種別一覧を取得
SELECT
    code,
    name,
    description
FROM land_categories
ORDER BY code;

-- name: UpsertLandCategory :one
-- 土地種別をUPSERT
INSERT INTO land_categories (
    code,
    name
) VALUES (
    $1, $2
)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name
RETURNING *;
