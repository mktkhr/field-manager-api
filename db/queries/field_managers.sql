-- name: GetFieldManager :one
-- 管理関係をIDで取得
SELECT
    id,
    field_id,
    manager_type,
    manager_id,
    created_at,
    created_by
FROM field_managers
WHERE id = $1;

-- name: GetFieldManagerByFieldAndManager :one
-- 圃場IDと管理者情報で管理関係を取得
SELECT
    id,
    field_id,
    manager_type,
    manager_id,
    created_at,
    created_by
FROM field_managers
WHERE field_id = $1 AND manager_type = $2 AND manager_id = $3;

-- name: ListFieldsByManager :many
-- 管理者タイプと管理者IDで管理配下の圃場ID一覧を取得(カーソルページネーション)
SELECT
    fm.id,
    fm.field_id,
    fm.manager_type,
    fm.manager_id,
    fm.created_at,
    fm.created_by
FROM field_managers fm
WHERE fm.manager_type = $1 AND fm.manager_id = $2
    AND CASE
        WHEN @cursor_created_at::timestamptz IS NULL THEN TRUE
        ELSE (fm.created_at < @cursor_created_at)
             OR (fm.created_at = @cursor_created_at AND fm.id < @cursor_id)
    END
ORDER BY fm.created_at DESC, fm.id DESC
LIMIT @page_limit;

-- name: ListManagersByField :many
-- 圃場IDで管理者一覧を取得
SELECT
    id,
    field_id,
    manager_type,
    manager_id,
    created_at,
    created_by
FROM field_managers
WHERE field_id = $1
ORDER BY manager_type, created_at DESC;

-- name: CreateFieldManager :one
-- 管理関係を作成
INSERT INTO field_managers (
    field_id,
    manager_type,
    manager_id,
    created_by
) VALUES (
    $1, $2, $3, $4
) RETURNING id, field_id, manager_type, manager_id, created_at, created_by;

-- name: DeleteFieldManager :exec
-- 管理関係をIDで削除
DELETE FROM field_managers WHERE id = $1;

-- name: DeleteFieldManagerByFieldAndManager :exec
-- 圃場IDと管理者情報で管理関係を削除
DELETE FROM field_managers
WHERE field_id = $1 AND manager_type = $2 AND manager_id = $3;

-- name: CountFieldsByManager :one
-- 管理者タイプと管理者IDで管理配下の圃場数を取得
SELECT COUNT(*) FROM field_managers
WHERE manager_type = $1 AND manager_id = $2;

-- name: CountManagersByField :one
-- 圃場IDで管理者数を取得
SELECT COUNT(*) FROM field_managers WHERE field_id = $1;

-- name: ExistsFieldManager :one
-- 管理関係の存在確認
SELECT EXISTS(
    SELECT 1 FROM field_managers
    WHERE field_id = $1 AND manager_type = $2 AND manager_id = $3
) AS exists;
