package port

import (
	"context"
	"io"
)

// StorageClient はオブジェクトストレージ(S3/RustFS)操作のインターフェース
type StorageClient interface {
	// Upload はデータをストレージにアップロードする
	Upload(ctx context.Context, key string, data io.Reader, contentType string) error

	// Download はストレージからデータをダウンロードする
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// GetObjectStream はストレージからストリーミング読み取り用のリーダーを取得する
	GetObjectStream(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete はストレージからオブジェクトを削除する
	Delete(ctx context.Context, key string) error

	// Exists はオブジェクトが存在するかどうかを確認する
	Exists(ctx context.Context, key string) (bool, error)
}
