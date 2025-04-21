package aws

import (
	"context"
	"mime/multipart"
)

// TextractClientInterface はTextractクライアントのインターフェース
type TextractClientInterface interface {
	ExtractTextFromDocument(ctx context.Context, file *multipart.FileHeader) (*TextractResult, error)
	ExtractTextFromS3Key(ctx context.Context, s3Key string) (*TextractResult, error)
}

// インターフェースを実装していることを静的にチェック
var _ TextractClientInterface = (*TextractClient)(nil)
