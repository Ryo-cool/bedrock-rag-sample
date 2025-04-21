package aws

import (
	"context"
	"mime/multipart"
)

// S3ClientInterface はS3クライアントのインターフェース
type S3ClientInterface interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, customPath string) (string, error)
	GetFileURL(ctx context.Context, key string) (string, error)
	DownloadFileContent(ctx context.Context, key string) ([]byte, error)
}

// インターフェースを実装していることを静的にチェック
var _ S3ClientInterface = (*S3Client)(nil)
