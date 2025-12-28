-- 合筆履歴テーブル
-- 圃場が合筆された際の関係を記録
CREATE TABLE field_mergers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- 合筆先圃場(他の圃場を吸収した圃場)
    merged_field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,

    -- ソース圃場(合筆先に統合された圃場の1つ)
    source_field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,

    -- 合筆日時
    merged_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 理由/備考(任意)
    reason TEXT,

    -- 監査
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID
);

-- 制約: 合筆先とソースは異なる必要がある
ALTER TABLE field_mergers ADD CONSTRAINT chk_merger_different_fields
    CHECK (merged_field_id != source_field_id);

-- インデックス
CREATE INDEX idx_field_mergers_merged ON field_mergers(merged_field_id);
CREATE INDEX idx_field_mergers_source ON field_mergers(source_field_id);
CREATE INDEX idx_field_mergers_merged_at ON field_mergers(merged_at);

-- コメント
COMMENT ON TABLE field_mergers IS '合筆履歴';
COMMENT ON COLUMN field_mergers.id IS '主キー';
COMMENT ON COLUMN field_mergers.merged_field_id IS 'ソース圃場を吸収した合筆先圃場';
COMMENT ON COLUMN field_mergers.source_field_id IS '合筆先に統合されたソース圃場';
COMMENT ON COLUMN field_mergers.merged_at IS '合筆日時';
COMMENT ON COLUMN field_mergers.reason IS '理由/備考';
COMMENT ON COLUMN field_mergers.created_at IS '作成日時';
COMMENT ON COLUMN field_mergers.created_by IS '作成者ID';
