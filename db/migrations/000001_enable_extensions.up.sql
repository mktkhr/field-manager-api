-- PostgreSQL拡張機能の有効化
-- uuid-ossp: UUID生成
-- postgis: 空間データ型と関数
-- pg_trgm: あいまい検索用トライグラム

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- updated_atの自動更新用function
CREATE OR REPLACE FUNCTION refresh_updated_at() RETURNS trigger AS $$
BEGIN
  -- 更新前後でupdated_atが変わらない場合はNULLに設定
  IF NEW.updated_at = OLD.updated_at THEN
    NEW.updated_at := NULL;
  END IF;

  -- updated_atがNULLの場合、更新前の値を再設定
  IF NEW.updated_at IS NULL THEN
    NEW.updated_at := OLD.updated_at;
  END IF;

  -- updated_atがNULLの場合、現在のタイムスタンプを設定
  IF NEW.updated_at IS NULL THEN
    NEW.updated_at := CURRENT_TIMESTAMP;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
