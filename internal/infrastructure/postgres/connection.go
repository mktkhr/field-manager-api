// Package postgres はPostgreSQLへの接続機能を提供する
package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mktkhr/field-manager-api/internal/config"
)

// CreateConnectionPool はデータベース接続プールを作成する
func CreateConnectionPool(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// パスワードに特殊文字が含まれる場合に備えてurl.UserPasswordを使用
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}
	q := u.Query()
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()

	return pgxpool.New(ctx, u.String())
}
