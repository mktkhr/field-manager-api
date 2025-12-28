-- 分筆履歴テーブル
-- 圃場が分筆された際の親子関係を記録
CREATE TABLE field_divisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- 親圃場(分筆前の元の圃場)
    parent_field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,

    -- 子圃場(分筆後の新しい圃場)
    child_field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,

    -- 分筆日時
    divided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 理由/備考(任意)
    reason TEXT,

    -- 監査
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID
);

-- 制約: 親と子は異なる必要がある
ALTER TABLE field_divisions ADD CONSTRAINT chk_division_different_fields
    CHECK (parent_field_id != child_field_id);

-- インデックス
CREATE INDEX idx_field_divisions_parent ON field_divisions(parent_field_id);
CREATE INDEX idx_field_divisions_child ON field_divisions(child_field_id);
CREATE INDEX idx_field_divisions_divided_at ON field_divisions(divided_at);

-- コメント
COMMENT ON TABLE field_divisions IS '分筆履歴(親子関係のみ)';
COMMENT ON COLUMN field_divisions.id IS '主キー';
COMMENT ON COLUMN field_divisions.parent_field_id IS '分筆前の元の圃場';
COMMENT ON COLUMN field_divisions.child_field_id IS '分筆により作成された新しい圃場';
COMMENT ON COLUMN field_divisions.divided_at IS '分筆日時';
COMMENT ON COLUMN field_divisions.reason IS '理由/備考';
COMMENT ON COLUMN field_divisions.created_at IS '作成日時';
COMMENT ON COLUMN field_divisions.created_by IS '作成者ID';
