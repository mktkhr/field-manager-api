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

// s3Client はStorageClientの実装
type s3Client struct {
	client *s3.Client
	bucket string
}

// NewS3Client は新しいS3Clientを作成する
func NewS3Client(ctx context.Context, cfg *appConfig.AWSConfig) (port.StorageClient, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	if cfg.LocalStackEnabled {
		opts = append(opts, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
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
		client: client,
		bucket: cfg.S3Bucket,
	}, nil
}

// Upload はデータをS3にアップロードする
func (c *s3Client) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	return err
}

// Download はS3からデータをダウンロードする
func (c *s3Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := c.client.GetObject(ctx, &s3.GetObjectInput{
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
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// Exists はオブジェクトが存在するかどうかを確認する
func (c *s3Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}
