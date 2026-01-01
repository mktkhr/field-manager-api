-- クラスタリング結果テーブル
-- 全圃場データを対象にH3インデックスでグループ化した結果を保存
CREATE TABLE cluster_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resolution INT NOT NULL CHECK (resolution IN (3, 5, 7, 9)),
    h3_index VARCHAR(15) NOT NULL,
    field_count INT NOT NULL,
    center_lat DOUBLE PRECISION NOT NULL,
    center_lng DOUBLE PRECISION NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 解像度+H3インデックスで一意
    UNIQUE (resolution, h3_index)
);

-- インデックス
CREATE INDEX idx_cluster_results_resolution ON cluster_results(resolution);
CREATE INDEX idx_cluster_results_h3_index ON cluster_results(h3_index);
CREATE INDEX idx_cluster_results_calculated_at ON cluster_results(calculated_at);

-- コメント
COMMENT ON TABLE cluster_results IS 'H3クラスタリング結果(全圃場対象)';
COMMENT ON COLUMN cluster_results.resolution IS 'H3解像度(3, 5, 7, 9)';
COMMENT ON COLUMN cluster_results.h3_index IS 'H3インデックス(16進数文字列)';
COMMENT ON COLUMN cluster_results.field_count IS 'クラスターに含まれる圃場数';
COMMENT ON COLUMN cluster_results.center_lat IS 'クラスター中心の緯度';
COMMENT ON COLUMN cluster_results.center_lng IS 'クラスター中心の経度';
COMMENT ON COLUMN cluster_results.calculated_at IS '計算日時';
