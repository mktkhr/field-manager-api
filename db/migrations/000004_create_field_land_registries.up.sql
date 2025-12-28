-- 農地台帳テーブル(PinInfo正規化)
-- 圃場との1対多関係
CREATE TABLE field_land_registries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    field_id UUID NOT NULL REFERENCES fields(id),

    -- 基本情報
    farmer_number VARCHAR(64),
    address TEXT,
    area_sqm INTEGER,

    -- 土地種別(FK)
    land_category_code VARCHAR(10) REFERENCES land_categories(code),

    -- 遊休農地状況(FK)
    idle_land_status_code VARCHAR(10) REFERENCES idle_land_statuses(code),

    -- 各種日付
    descriptive_study_data DATE,

    -- 監査
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_field_land_registries_field_id ON field_land_registries(field_id);
CREATE INDEX idx_field_land_registries_farmer_number ON field_land_registries(farmer_number);
CREATE INDEX idx_field_land_registries_land_category ON field_land_registries(land_category_code);
CREATE INDEX idx_field_land_registries_idle ON field_land_registries(idle_land_status_code);

-- コメント
COMMENT ON TABLE field_land_registries IS '農地台帳';
COMMENT ON COLUMN field_land_registries.id IS '主キー';
COMMENT ON COLUMN field_land_registries.field_id IS '圃場ID(FK)';
COMMENT ON COLUMN field_land_registries.farmer_number IS 'ハッシュ化された耕作者識別番号';
COMMENT ON COLUMN field_land_registries.address IS '所在地';
COMMENT ON COLUMN field_land_registries.area_sqm IS '面積(平方メートル)';
COMMENT ON COLUMN field_land_registries.land_category_code IS '土地種別コード(FK)';
COMMENT ON COLUMN field_land_registries.idle_land_status_code IS '遊休農地状況コード(FK)';
COMMENT ON COLUMN field_land_registries.descriptive_study_data IS '実態調査日';
COMMENT ON COLUMN field_land_registries.created_at IS '作成日時';
COMMENT ON COLUMN field_land_registries.updated_at IS '更新日時';

-- updated_at自動更新トリガー
CREATE TRIGGER trg_field_land_registries_updated_at
    BEFORE UPDATE ON field_land_registries
    FOR EACH ROW
    EXECUTE FUNCTION refresh_updated_at();

-- fields.name自動更新関数
CREATE FUNCTION update_field_name() RETURNS TRIGGER AS $$
BEGIN
    UPDATE fields SET name = (
        SELECT COALESCE(
            NULLIF(string_agg(address, ', '), ''),
            '名称不明'
        )
        FROM field_land_registries
        WHERE field_id = COALESCE(NEW.field_id, OLD.field_id)
    )
    WHERE id = COALESCE(NEW.field_id, OLD.field_id);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- fields.name自動更新トリガー
CREATE TRIGGER trg_update_field_name
    AFTER INSERT OR UPDATE OR DELETE ON field_land_registries
    FOR EACH ROW EXECUTE FUNCTION update_field_name();
