-- name: GetClusterResults :many
-- 指定解像度のクラスター結果を取得
SELECT
    id,
    resolution,
    h3_index,
    field_count,
    center_lat,
    center_lng,
    calculated_at
FROM cluster_results
WHERE resolution = $1
ORDER BY h3_index;

-- name: UpsertClusterResult :exec
-- クラスター結果をUPSERT
INSERT INTO cluster_results (
    id,
    resolution,
    h3_index,
    field_count,
    center_lat,
    center_lng,
    calculated_at
)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (resolution, h3_index)
DO UPDATE SET
    field_count = EXCLUDED.field_count,
    center_lat = EXCLUDED.center_lat,
    center_lng = EXCLUDED.center_lng,
    calculated_at = NOW();

-- name: DeleteClusterResultsByResolution :exec
-- 指定解像度のクラスター結果を全削除
DELETE FROM cluster_results
WHERE resolution = $1;

-- name: DeleteAllClusterResults :exec
-- 全クラスター結果を削除
DELETE FROM cluster_results;

-- name: AggregateClustersByRes3 :many
-- H3解像度3でfieldsを集計
SELECT
    h3_index_res3 AS h3_index,
    COUNT(*)::INT AS field_count
FROM fields
WHERE h3_index_res3 IS NOT NULL
GROUP BY h3_index_res3;

-- name: AggregateClustersByRes5 :many
-- H3解像度5でfieldsを集計
SELECT
    h3_index_res5 AS h3_index,
    COUNT(*)::INT AS field_count
FROM fields
WHERE h3_index_res5 IS NOT NULL
GROUP BY h3_index_res5;

-- name: AggregateClustersByRes7 :many
-- H3解像度7でfieldsを集計
SELECT
    h3_index_res7 AS h3_index,
    COUNT(*)::INT AS field_count
FROM fields
WHERE h3_index_res7 IS NOT NULL
GROUP BY h3_index_res7;

-- name: AggregateClustersByRes9 :many
-- H3解像度9でfieldsを集計
SELECT
    h3_index_res9 AS h3_index,
    COUNT(*)::INT AS field_count
FROM fields
WHERE h3_index_res9 IS NOT NULL
GROUP BY h3_index_res9;
