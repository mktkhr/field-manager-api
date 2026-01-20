package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// fieldExistsChecker はFieldExistsCheckerの実装
type fieldExistsChecker struct {
	db *pgxpool.Pool
}

// NewFieldExistsChecker は新しいFieldExistsCheckerを作成する
func NewFieldExistsChecker(db *pgxpool.Pool) *fieldExistsChecker {
	return &fieldExistsChecker{db: db}
}

// ExistsByID は圃場の存在を確認する
func (c *fieldExistsChecker) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM fields WHERE id = $1)`

	var exists bool
	err := c.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("圃場存在確認に失敗: %w", err)
	}

	return exists, nil
}
