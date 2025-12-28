-- 圃場(農地区画)メインテーブル
CREATE TABLE fields (
    -- 主キーと識別子
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- ジオメトリ(PostGIS)
    geometry GEOMETRY(POLYGON, 4326) NOT NULL,
    centroid GEOMETRY(POINT, 4326) NOT NULL,
    area_sqm DOUBLE PRECISION GENERATED ALWAYS AS (ST_Area(geometry::geography)) STORED,  -- 面積(平方メートル、自動計算)

    -- H3インデックス(クラスタリング用4解像度)
    h3_index_res3 VARCHAR(15),  -- 約100km - 地方レベル
    h3_index_res5 VARCHAR(15),  -- 約10km - 都道府県レベル
    h3_index_res7 VARCHAR(15),  -- 約1km - 市区町村レベル
    h3_index_res9 VARCHAR(15),  -- 約100m - 詳細レベル

    -- wagri API由来の属性
    city_code VARCHAR(10) NOT NULL,

    -- 圃場名(field_land_registries.addressから自動生成)
    name TEXT NOT NULL DEFAULT '名称不明',

    -- 土壌タイプ(FK)
    soil_type_id UUID REFERENCES soil_types(id),

    -- 監査
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- 空間クエリ用GISTインデックス
CREATE INDEX idx_fields_geometry_gist ON fields USING GIST(geometry);
CREATE INDEX idx_fields_centroid_gist ON fields USING GIST(centroid);

-- クラスタリング用H3インデックス
CREATE INDEX idx_fields_h3_res3 ON fields(h3_index_res3) WHERE h3_index_res3 IS NOT NULL;
CREATE INDEX idx_fields_h3_res5 ON fields(h3_index_res5) WHERE h3_index_res5 IS NOT NULL;
CREATE INDEX idx_fields_h3_res7 ON fields(h3_index_res7) WHERE h3_index_res7 IS NOT NULL;
CREATE INDEX idx_fields_h3_res9 ON fields(h3_index_res9) WHERE h3_index_res9 IS NOT NULL;

-- 一般的なクエリ用インデックス
CREATE INDEX idx_fields_city_code ON fields(city_code);
CREATE INDEX idx_fields_soil_type ON fields(soil_type_id);

-- あいまい検索用GINインデックス
CREATE INDEX idx_fields_name_trgm ON fields USING GIN(name gin_trgm_ops);

-- コメント
COMMENT ON TABLE fields IS '圃場マスタテーブル';
COMMENT ON COLUMN fields.id IS '主キー';
COMMENT ON COLUMN fields.geometry IS 'ポリゴン形状(SRID: 4326 = WGS84)';
COMMENT ON COLUMN fields.centroid IS '重心座標(SRID: 4326 = WGS84)';
COMMENT ON COLUMN fields.area_sqm IS '面積(平方メートル、自動計算)';
COMMENT ON COLUMN fields.h3_index_res3 IS 'H3インデックス解像度3(約100km - 地方クラスタリング)';
COMMENT ON COLUMN fields.h3_index_res5 IS 'H3インデックス解像度5(約10km - 都道府県クラスタリング)';
COMMENT ON COLUMN fields.h3_index_res7 IS 'H3インデックス解像度7(約1km - 市区町村クラスタリング)';
COMMENT ON COLUMN fields.h3_index_res9 IS 'H3インデックス解像度9(約100m - 詳細クラスタリング)';
COMMENT ON COLUMN fields.city_code IS '市区町村コード';
COMMENT ON COLUMN fields.name IS '圃場名(field_land_registries.addressから自動生成)';
COMMENT ON COLUMN fields.soil_type_id IS '土壌タイプID(FK)';
COMMENT ON COLUMN fields.created_at IS '作成日時';
COMMENT ON COLUMN fields.updated_at IS '更新日時';
COMMENT ON COLUMN fields.created_by IS '作成者ID';
COMMENT ON COLUMN fields.updated_by IS '更新者ID';

-- updated_at自動更新トリガー
CREATE TRIGGER trg_fields_updated_at
    BEFORE UPDATE ON fields
    FOR EACH ROW
    EXECUTE FUNCTION refresh_updated_at();
