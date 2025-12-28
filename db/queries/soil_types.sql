-- name: GetSoilType :one
-- 土壌タイプをIDで取得
SELECT
    id,
    large_code,
    middle_code,
    small_code,
    small_name,
    description,
    created_at,
    updated_at
FROM soil_types
WHERE id = $1;

-- name: GetSoilTypeBySmallCode :one
-- 土壌タイプを小分類コードで取得
SELECT
    id,
    large_code,
    middle_code,
    small_code,
    small_name,
    description,
    created_at,
    updated_at
FROM soil_types
WHERE small_code = $1;

-- name: ListSoilTypes :many
-- 土壌タイプ一覧を取得
SELECT
    id,
    large_code,
    middle_code,
    small_code,
    small_name,
    description,
    created_at,
    updated_at
FROM soil_types
ORDER BY large_code, middle_code, small_code;

-- name: UpsertSoilType :one
-- 土壌タイプをUPSERT
INSERT INTO soil_types (
    large_code,
    middle_code,
    small_code,
    small_name
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (small_code) DO UPDATE SET
    large_code = EXCLUDED.large_code,
    middle_code = EXCLUDED.middle_code,
    small_name = EXCLUDED.small_name,
    updated_at = NOW()
RETURNING *;
