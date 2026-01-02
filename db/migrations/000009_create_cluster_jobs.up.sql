-- クラスタリングジョブ管理テーブル
CREATE TABLE cluster_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT
);

-- インデックス
CREATE INDEX idx_cluster_jobs_status_priority ON cluster_jobs(status, priority DESC, created_at);

-- コメント
COMMENT ON TABLE cluster_jobs IS 'クラスタリングジョブ管理';
COMMENT ON COLUMN cluster_jobs.status IS 'ジョブステータス(pending, processing, completed, failed)';
COMMENT ON COLUMN cluster_jobs.priority IS '優先度(高いほど先に処理)';
COMMENT ON COLUMN cluster_jobs.created_at IS 'ジョブ作成日時';
COMMENT ON COLUMN cluster_jobs.started_at IS 'ジョブ開始日時';
COMMENT ON COLUMN cluster_jobs.completed_at IS 'ジョブ完了日時';
COMMENT ON COLUMN cluster_jobs.error_message IS 'エラーメッセージ(失敗時のみ)';
