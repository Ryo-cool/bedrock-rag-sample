package services

import (
	"context"
	"mime/multipart"

	"bedrock-rag-sample/backend/pkg/aws"
)

// UploadServiceInterface はファイルアップロードサービス操作のためのインターフェース
type UploadServiceInterface interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader) (*UploadFileResult, error)
	GetS3Client() aws.S3ClientInterface // S3クライアントへのアクセスを提供
}

// TODO: UploadService が UploadServiceInterface を実装していることを静的にチェックする
// 例: var _ UploadServiceInterface = (*UploadService)(nil)
// (UploadService に GetS3Client を実装してから追加)
