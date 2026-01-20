-- 管理者タイプのENUM定義
-- 拡張時: ALTER TYPE manager_type ADD VALUE 'new_type';
CREATE TYPE manager_type AS ENUM ('organization', 'user', 'region');

-- 圃場管理者の中間テーブル
-- 1つの圃場に複数の管理者(組織、ユーザー、地域など)を紐付け可能
CREATE TABLE field_managers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    manager_type manager_type NOT NULL,
    manager_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    UNIQUE (field_id, manager_type, manager_id)
);

-- 管理者タイプ+管理者IDでの検索用インデックス
-- 例: organization + org_uuid で管理配下の圃場を取得
CREATE INDEX idx_field_managers_lookup
ON field_managers(manager_type, manager_id);

-- field_id逆引き用インデックス
-- 例: 特定の圃場に紐づく管理者一覧を取得
CREATE INDEX idx_field_managers_field_id
ON field_managers(field_id);

-- H3検索との複合用インデックス(将来拡張)
-- manager_type + manager_id + field_id でカバリングクエリ可能
CREATE INDEX idx_field_managers_type_id_field
ON field_managers(manager_type, manager_id, field_id);

-- テーブルコメント
COMMENT ON TABLE field_managers IS '圃場と管理者(組織・ユーザー・地域)の中間テーブル';
COMMENT ON COLUMN field_managers.id IS '管理関係ID';
COMMENT ON COLUMN field_managers.field_id IS '圃場ID';
COMMENT ON COLUMN field_managers.manager_type IS '管理者タイプ(organization/user/region)';
COMMENT ON COLUMN field_managers.manager_id IS '管理者のUUID(組織ID/ユーザーID/地域IDなど)';
COMMENT ON COLUMN field_managers.created_at IS '作成日時';
COMMENT ON COLUMN field_managers.created_by IS '作成者UUID';
