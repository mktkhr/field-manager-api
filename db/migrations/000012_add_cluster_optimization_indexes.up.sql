-- Cluster Res9集計の最適化: カバリングインデックス追加
-- H3セルベースの集計クエリを高速化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fields_h3_res9_covering
ON fields(h3_index_res9)
INCLUDE (id)
WHERE h3_index_res9 IS NOT NULL;
