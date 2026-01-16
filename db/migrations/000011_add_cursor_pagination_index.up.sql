-- カーソルベースページネーション用インデックス追加
-- 圃場一覧のパフォーマンス向上のため、created_at DESC, id DESC の複合インデックスを作成

CREATE INDEX idx_fields_cursor_pagination ON fields(created_at DESC, id DESC);

COMMENT ON INDEX idx_fields_cursor_pagination IS 'カーソルベースページネーション用インデックス';
