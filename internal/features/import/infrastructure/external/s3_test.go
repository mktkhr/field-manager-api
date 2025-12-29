package external

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
)

// mockS3API はS3 APIのモック実装
type mockS3API struct {
	putObjectFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	getObjectFunc    func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	deleteObjectFunc func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	headObjectFunc   func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

func (m *mockS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3API) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, params, optFns...)
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader([]byte("test content"))),
	}, nil
}

func (m *mockS3API) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(ctx, params, optFns...)
	}
	return &s3.DeleteObjectOutput{}, nil
}

func (m *mockS3API) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if m.headObjectFunc != nil {
		return m.headObjectFunc(ctx, params, optFns...)
	}
	return &s3.HeadObjectOutput{
		ResultMetadata: middleware.Metadata{},
	}, nil
}

// TestS3Client_Upload はUploadメソッドが正常系とS3エラーを正しく処理することをテストする
func TestS3Client_Upload(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		data        string
		contentType string
		mockFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
		wantErr     bool
	}{
		{
			name:        "success",
			key:         "test/file.json",
			data:        `{"test": "data"}`,
			contentType: "application/json",
			mockFunc:    nil,
			wantErr:     false,
		},
		{
			name:        "S3 error",
			key:         "test/file.json",
			data:        `{"test": "data"}`,
			contentType: "application/json",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return nil, errors.New("S3 error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3API{
				putObjectFunc: tt.mockFunc,
			}
			client := &s3Client{
				api:    mock,
				bucket: "test-bucket",
			}

			err := client.Upload(context.Background(), tt.key, bytes.NewReader([]byte(tt.data)), tt.contentType)

			if tt.wantErr {
				if err == nil {
					t.Error("Upload() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Upload() error = %v", err)
			}
		})
	}
}

// TestS3Client_Download はDownloadメソッドが正常系とS3エラーを正しく処理することをテストする
func TestS3Client_Download(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		mockFunc    func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
		wantContent string
		wantErr     bool
	}{
		{
			name: "success",
			key:  "test/file.json",
			mockFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"test": "data"}`))),
				}, nil
			},
			wantContent: `{"test": "data"}`,
			wantErr:     false,
		},
		{
			name: "S3 error",
			key:  "test/file.json",
			mockFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("S3 error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3API{
				getObjectFunc: tt.mockFunc,
			}
			client := &s3Client{
				api:    mock,
				bucket: "test-bucket",
			}

			reader, err := client.Download(context.Background(), tt.key)

			if tt.wantErr {
				if err == nil {
					t.Error("Download() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Download() error = %v", err)
				return
			}

			defer func() {
				_ = reader.Close()
			}()

			content, err := io.ReadAll(reader)
			if err != nil {
				t.Errorf("ReadAll() error = %v", err)
				return
			}

			if string(content) != tt.wantContent {
				t.Errorf("Download() content = %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}

// TestS3Client_GetObjectStream はGetObjectStreamメソッドがオブジェクトをストリームとして取得できることをテストする
func TestS3Client_GetObjectStream(t *testing.T) {
	mock := &mockS3API{
		getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte("streaming content"))),
			}, nil
		},
	}
	client := &s3Client{
		api:    mock,
		bucket: "test-bucket",
	}

	reader, err := client.GetObjectStream(context.Background(), "test/file.json")
	if err != nil {
		t.Errorf("GetObjectStream() error = %v", err)
		return
	}

	defer func() {
		_ = reader.Close()
	}()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("ReadAll() error = %v", err)
		return
	}

	if string(content) != "streaming content" {
		t.Errorf("GetObjectStream() content = %q, want %q", string(content), "streaming content")
	}
}

// TestS3Client_Delete はDeleteメソッドが正常系とS3エラーを正しく処理することをテストする
func TestS3Client_Delete(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		mockFunc func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
		wantErr  bool
	}{
		{
			name:     "success",
			key:      "test/file.json",
			mockFunc: nil,
			wantErr:  false,
		},
		{
			name: "S3 error",
			key:  "test/file.json",
			mockFunc: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
				return nil, errors.New("S3 error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3API{
				deleteObjectFunc: tt.mockFunc,
			}
			client := &s3Client{
				api:    mock,
				bucket: "test-bucket",
			}

			err := client.Delete(context.Background(), tt.key)

			if tt.wantErr {
				if err == nil {
					t.Error("Delete() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() error = %v", err)
			}
		})
	}
}

// TestS3Client_Exists はExistsメソッドがオブジェクトの存在確認を正しく行うことをテストする
func TestS3Client_Exists(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		mockFunc func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
		want     bool
	}{
		{
			name:     "exists",
			key:      "test/file.json",
			mockFunc: nil,
			want:     true,
		},
		{
			name: "not exists",
			key:  "test/nonexistent.json",
			mockFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
				return nil, errors.New("NoSuchKey")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3API{
				headObjectFunc: tt.mockFunc,
			}
			client := &s3Client{
				api:    mock,
				bucket: "test-bucket",
			}

			exists, err := client.Exists(context.Background(), tt.key)
			if err != nil {
				t.Errorf("Exists() error = %v", err)
				return
			}

			if exists != tt.want {
				t.Errorf("Exists() = %v, want %v", exists, tt.want)
			}
		})
	}
}

// TestS3Client_UploadValidatesParams はUploadメソッドがS3 APIに正しいパラメータを渡すことをテストする
func TestS3Client_UploadValidatesParams(t *testing.T) {
	var capturedParams *s3.PutObjectInput

	mock := &mockS3API{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			capturedParams = params
			return &s3.PutObjectOutput{}, nil
		},
	}
	client := &s3Client{
		api:    mock,
		bucket: "my-bucket",
	}

	data := []byte("test content")
	err := client.Upload(context.Background(), "path/to/file.txt", bytes.NewReader(data), "text/plain")
	if err != nil {
		t.Errorf("Upload() error = %v", err)
		return
	}

	if capturedParams == nil {
		t.Error("PutObject was not called")
		return
	}

	if *capturedParams.Bucket != "my-bucket" {
		t.Errorf("Bucket = %q, want %q", *capturedParams.Bucket, "my-bucket")
	}
	if *capturedParams.Key != "path/to/file.txt" {
		t.Errorf("Key = %q, want %q", *capturedParams.Key, "path/to/file.txt")
	}
	if *capturedParams.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want %q", *capturedParams.ContentType, "text/plain")
	}
}

// TestS3Client_DownloadValidatesParams はDownloadメソッドがS3 APIに正しいパラメータを渡すことをテストする
func TestS3Client_DownloadValidatesParams(t *testing.T) {
	var capturedParams *s3.GetObjectInput

	mock := &mockS3API{
		getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			capturedParams = params
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte("content"))),
			}, nil
		},
	}
	client := &s3Client{
		api:    mock,
		bucket: "my-bucket",
	}

	reader, err := client.Download(context.Background(), "path/to/file.txt")
	if err != nil {
		t.Errorf("Download() error = %v", err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()

	if capturedParams == nil {
		t.Error("GetObject was not called")
		return
	}

	if *capturedParams.Bucket != "my-bucket" {
		t.Errorf("Bucket = %q, want %q", *capturedParams.Bucket, "my-bucket")
	}
	if *capturedParams.Key != "path/to/file.txt" {
		t.Errorf("Key = %q, want %q", *capturedParams.Key, "path/to/file.txt")
	}
}

// TestS3Client_DeleteValidatesParams はDeleteメソッドがS3 APIに正しいパラメータを渡すことをテストする
func TestS3Client_DeleteValidatesParams(t *testing.T) {
	var capturedParams *s3.DeleteObjectInput

	mock := &mockS3API{
		deleteObjectFunc: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
			capturedParams = params
			return &s3.DeleteObjectOutput{}, nil
		},
	}
	client := &s3Client{
		api:    mock,
		bucket: "my-bucket",
	}

	err := client.Delete(context.Background(), "path/to/file.txt")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
		return
	}

	if capturedParams == nil {
		t.Error("DeleteObject was not called")
		return
	}

	if *capturedParams.Bucket != "my-bucket" {
		t.Errorf("Bucket = %q, want %q", *capturedParams.Bucket, "my-bucket")
	}
	if *capturedParams.Key != "path/to/file.txt" {
		t.Errorf("Key = %q, want %q", *capturedParams.Key, "path/to/file.txt")
	}
}

// TestS3Client_ExistsValidatesParams はExistsメソッドがS3 APIに正しいパラメータを渡すことをテストする
func TestS3Client_ExistsValidatesParams(t *testing.T) {
	var capturedParams *s3.HeadObjectInput

	mock := &mockS3API{
		headObjectFunc: func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			capturedParams = params
			return &s3.HeadObjectOutput{}, nil
		},
	}
	client := &s3Client{
		api:    mock,
		bucket: "my-bucket",
	}

	_, err := client.Exists(context.Background(), "path/to/file.txt")
	if err != nil {
		t.Errorf("Exists() error = %v", err)
		return
	}

	if capturedParams == nil {
		t.Error("HeadObject was not called")
		return
	}

	if *capturedParams.Bucket != "my-bucket" {
		t.Errorf("Bucket = %q, want %q", *capturedParams.Bucket, "my-bucket")
	}
	if *capturedParams.Key != "path/to/file.txt" {
		t.Errorf("Key = %q, want %q", *capturedParams.Key, "path/to/file.txt")
	}
}
