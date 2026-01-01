-- インポートジョブ管理テーブル
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    city_code VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending, processing, completed, failed, partially_completed

    total_records INTEGER,
    processed_records INTEGER NOT NULL DEFAULT 0,
    failed_records INTEGER NOT NULL DEFAULT 0,
    last_processed_batch INTEGER NOT NULL DEFAULT 0,

    s3_key TEXT,
    execution_arn TEXT,
    error_message TEXT,
    failed_record_ids JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

-- インデックス
CREATE INDEX idx_import_jobs_city_code ON import_jobs(city_code);
CREATE INDEX idx_import_jobs_status ON import_jobs(status);
CREATE INDEX idx_import_jobs_created_at ON import_jobs(created_at DESC);

-- コメント
COMMENT ON TABLE import_jobs IS 'wagriインポートジョブ管理テーブル';
COMMENT ON COLUMN import_jobs.id IS '主キー';
COMMENT ON COLUMN import_jobs.city_code IS '市区町村コード';
COMMENT ON COLUMN import_jobs.status IS 'ステータス(pending/processing/completed/failed/partially_completed)';
COMMENT ON COLUMN import_jobs.total_records IS '総レコード数';
COMMENT ON COLUMN import_jobs.processed_records IS '処理済みレコード数';
COMMENT ON COLUMN import_jobs.failed_records IS '失敗レコード数';
COMMENT ON COLUMN import_jobs.last_processed_batch IS '最後に処理したバッチ番号';
COMMENT ON COLUMN import_jobs.s3_key IS 'S3に保存したレスポンスのキー';
COMMENT ON COLUMN import_jobs.execution_arn IS 'Step Functions実行ARN';
COMMENT ON COLUMN import_jobs.error_message IS 'エラーメッセージ';
COMMENT ON COLUMN import_jobs.failed_record_ids IS '失敗したレコードIDのJSON配列';
COMMENT ON COLUMN import_jobs.created_at IS '作成日時';
COMMENT ON COLUMN import_jobs.started_at IS '処理開始日時';
COMMENT ON COLUMN import_jobs.completed_at IS '処理完了日時';
