package external

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appConfig "github.com/mktkhr/field-manager-api/internal/config"
	"github.com/mktkhr/field-manager-api/internal/features/import/application/port"
)

// s3API はS3操作のインターフェース
type s3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// s3Client はStorageClientの実装
type s3Client struct {
	api    s3API
	bucket string
}

// NewS3Client は新しいS3Clientを作成する
func NewS3Client(ctx context.Context, cfg *appConfig.AWSConfig) (port.StorageClient, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	if cfg.LocalStackEnabled {
		// LocalStack互換性のため deprecated API を使用
		//nolint:staticcheck
		opts = append(opts, config.WithEndpointResolverWithOptions(
			//nolint:staticcheck
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				//nolint:staticcheck
				return aws.Endpoint{
					URL:               cfg.LocalStackURL,
					HostnameImmutable: true,
				}, nil
			}),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.LocalStackEnabled {
			o.UsePathStyle = true
		}
	})

	return &s3Client{
		api:    client,
		bucket: cfg.S3Bucket,
	}, nil
}

// Upload はデータをS3にアップロードする
func (c *s3Client) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	_, err := c.api.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	return err
}

// Download はS3からデータをダウンロードする
func (c *s3Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := c.api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

// GetObjectStream はS3からストリーミング読み取り用のリーダーを取得する
func (c *s3Client) GetObjectStream(ctx context.Context, key string) (io.ReadCloser, error) {
	return c.Download(ctx, key)
}

// Delete はS3からオブジェクトを削除する
func (c *s3Client) Delete(ctx context.Context, key string) error {
	_, err := c.api.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// Exists はオブジェクトが存在するかどうかを確認する
func (c *s3Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.api.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}
