-- cluster_jobsテーブルにaffected_h3_cellsカラムを追加
-- NULLの場合は全範囲再計算、値がある場合は指定セルのみ再集計
ALTER TABLE cluster_jobs ADD COLUMN affected_h3_cells TEXT[];

COMMENT ON COLUMN cluster_jobs.affected_h3_cells IS '影響を受けたH3セルのリスト(NULLの場合は全範囲再計算)';
