-- 関数と拡張機能を依存関係の逆順で削除
DROP FUNCTION IF EXISTS refresh_updated_at();
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "postgis";
DROP EXTENSION IF EXISTS "uuid-ossp";
