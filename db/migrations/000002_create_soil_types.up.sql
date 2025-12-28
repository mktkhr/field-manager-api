-- 土壌タイプマスタテーブル
-- 階層構造: 大分類 -> 中分類 -> 小分類
CREATE TABLE soil_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    large_code VARCHAR(10) NOT NULL,
    middle_code VARCHAR(10) NOT NULL,
    small_code VARCHAR(20) NOT NULL UNIQUE,
    small_name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 階層検索用インデックス
CREATE INDEX idx_soil_types_large_code ON soil_types(large_code);
CREATE INDEX idx_soil_types_middle_code ON soil_types(middle_code);
CREATE INDEX idx_soil_types_small_code ON soil_types(small_code);

-- コメント
COMMENT ON TABLE soil_types IS '土壌分類マスタテーブル(階層: 大/中/小分類)';
COMMENT ON COLUMN soil_types.id IS '主キー';
COMMENT ON COLUMN soil_types.large_code IS '大分類コード(例: F3)';
COMMENT ON COLUMN soil_types.middle_code IS '中分類コード(例: F3a7)';
COMMENT ON COLUMN soil_types.small_code IS '小分類コード(例: F3a7t4) - ユニーク';
COMMENT ON COLUMN soil_types.small_name IS '小分類名(例: 粗粒グライ灰色低地土)';
COMMENT ON COLUMN soil_types.description IS '説明';
COMMENT ON COLUMN soil_types.created_at IS '作成日時';
COMMENT ON COLUMN soil_types.updated_at IS '更新日時';

-- updated_at自動更新トリガー
CREATE TRIGGER trg_soil_types_updated_at
    BEFORE UPDATE ON soil_types
    FOR EACH ROW
    EXECUTE FUNCTION refresh_updated_at();

-- 土地種別マスタ
CREATE TABLE land_categories (
    code VARCHAR(10) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT
);

COMMENT ON TABLE land_categories IS '土地種別マスタ';
COMMENT ON COLUMN land_categories.code IS '土地種別コード';
COMMENT ON COLUMN land_categories.name IS '土地種別名(例: 田、畑)';
COMMENT ON COLUMN land_categories.description IS '説明';

-- 遊休農地状況マスタ
CREATE TABLE idle_land_statuses (
    code VARCHAR(10) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT
);

COMMENT ON TABLE idle_land_statuses IS '遊休農地状況マスタ';
COMMENT ON COLUMN idle_land_statuses.code IS '遊休農地状況コード';
COMMENT ON COLUMN idle_land_statuses.name IS '遊休農地状況名';
COMMENT ON COLUMN idle_land_statuses.description IS '説明';
