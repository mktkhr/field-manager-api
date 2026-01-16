-- name: GetField :one
-- 圃場をIDで取得
SELECT
    id,
    geometry,
    centroid,
    area_sqm,
    h3_index_res3,
    h3_index_res5,
    h3_index_res7,
    h3_index_res9,
    city_code,
    name,
    soil_type_id,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM fields
WHERE id = $1;

-- name: ListFieldsByCursor :many
-- カーソルベースで圃場一覧を取得
-- cursor_created_atとcursor_idがNULLの場合は先頭から取得
SELECT
    id,
    geometry,
    centroid,
    area_sqm,
    h3_index_res3,
    h3_index_res5,
    h3_index_res7,
    h3_index_res9,
    city_code,
    name,
    soil_type_id,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM fields
WHERE
    CASE
        WHEN @cursor_created_at::timestamptz IS NULL THEN TRUE
        ELSE (created_at < @cursor_created_at)
             OR (created_at = @cursor_created_at AND id < @cursor_id)
    END
ORDER BY created_at DESC, id DESC
LIMIT @page_limit;

-- name: ListFieldsByCityCode :many
-- 市区町村コードで圃場一覧を取得
SELECT
    id,
    geometry,
    centroid,
    area_sqm,
    h3_index_res3,
    h3_index_res5,
    h3_index_res7,
    h3_index_res9,
    city_code,
    name,
    soil_type_id,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM fields
WHERE city_code = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: CreateField :one
-- 圃場を作成
INSERT INTO fields (
    geometry,
    centroid,
    h3_index_res3,
    h3_index_res5,
    h3_index_res7,
    h3_index_res9,
    city_code,
    name,
    soil_type_id,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateField :one
-- 圃場を更新
UPDATE fields
SET
    geometry = COALESCE($2, geometry),
    centroid = COALESCE($3, centroid),
    h3_index_res3 = COALESCE($4, h3_index_res3),
    h3_index_res5 = COALESCE($5, h3_index_res5),
    h3_index_res7 = COALESCE($6, h3_index_res7),
    h3_index_res9 = COALESCE($7, h3_index_res9),
    city_code = COALESCE($8, city_code),
    name = COALESCE($9, name),
    soil_type_id = $10,
    updated_by = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteField :exec
-- 圃場を削除
DELETE FROM fields WHERE id = $1;

-- name: UpsertField :one
-- 圃場をUPSERT(wagriインポート用)
-- geometry, centroidはWKB形式のbytea型で受け取り、ST_GeomFromWKBで変換
INSERT INTO fields (
    id,
    geometry,
    centroid,
    h3_index_res3,
    h3_index_res5,
    h3_index_res7,
    h3_index_res9,
    city_code,
    soil_type_id
) VALUES (
    @id,
    ST_GeomFromWKB(@geometry_wkb::bytea, 4326),
    ST_GeomFromWKB(@centroid_wkb::bytea, 4326),
    @h3_index_res3, @h3_index_res5, @h3_index_res7, @h3_index_res9, @city_code, @soil_type_id
)
ON CONFLICT (id) DO UPDATE SET
    geometry = EXCLUDED.geometry,
    centroid = EXCLUDED.centroid,
    h3_index_res3 = EXCLUDED.h3_index_res3,
    h3_index_res5 = EXCLUDED.h3_index_res5,
    h3_index_res7 = EXCLUDED.h3_index_res7,
    h3_index_res9 = EXCLUDED.h3_index_res9,
    city_code = EXCLUDED.city_code,
    soil_type_id = EXCLUDED.soil_type_id,
    updated_at = NOW()
RETURNING *;

-- name: GetH3IndexesByFieldIDs :many
-- 指定IDのフィールドのH3インデックスを取得(差分更新のプリフェッチ用)
SELECT
    id,
    h3_index_res3,
    h3_index_res5,
    h3_index_res7,
    h3_index_res9
FROM fields
WHERE id = ANY(@ids::UUID[]);
