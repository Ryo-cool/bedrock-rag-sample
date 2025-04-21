package services

import (
	"context"
	"mime/multipart"
)

// DocumentServiceInterface はドキュメント処理サービスのインターフェース
type DocumentServiceInterface interface {
	ProcessDocument(ctx context.Context, file *multipart.FileHeader) (*DocumentProcessResult, error)
	ProcessDocumentByS3Key(ctx context.Context, s3Key string) (*DocumentProcessResult, error)
	// 他の DocumentService メソッドが必要であればここに追加
}

// TODO: DocumentService が DocumentServiceInterface を実装していることを静的にチェックする
// 例: var _ DocumentServiceInterface = (*DocumentService)(nil)
