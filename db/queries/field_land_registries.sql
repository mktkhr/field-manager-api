-- name: GetFieldLandRegistry :one
-- 農地台帳をIDで取得
SELECT
    id,
    field_id,
    farmer_number,
    address,
    area_sqm,
    land_category_code,
    idle_land_status_code,
    descriptive_study_data,
    created_at,
    updated_at
FROM field_land_registries
WHERE id = $1;

-- name: ListFieldLandRegistriesByFieldID :many
-- 圃場IDで農地台帳一覧を取得
SELECT
    id,
    field_id,
    farmer_number,
    address,
    area_sqm,
    land_category_code,
    idle_land_status_code,
    descriptive_study_data,
    created_at,
    updated_at
FROM field_land_registries
WHERE field_id = $1
ORDER BY created_at;

-- name: CreateFieldLandRegistry :one
-- 農地台帳を作成
INSERT INTO field_land_registries (
    field_id,
    farmer_number,
    address,
    area_sqm,
    land_category_code,
    idle_land_status_code,
    descriptive_study_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: DeleteFieldLandRegistriesByFieldID :exec
-- 圃場IDで農地台帳を削除(REPLACE方式用)
DELETE FROM field_land_registries WHERE field_id = $1;

-- name: DeleteFieldLandRegistriesByFieldIDs :exec
-- 複数の圃場IDで農地台帳を一括削除(バッチREPLACE方式用)
DELETE FROM field_land_registries WHERE field_id = ANY($1::uuid[]);

-- name: CountFieldLandRegistriesByFieldID :one
-- 圃場IDで農地台帳の件数を取得
SELECT COUNT(*) FROM field_land_registries WHERE field_id = $1;
