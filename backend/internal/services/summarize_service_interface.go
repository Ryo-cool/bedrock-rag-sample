package services

import (
	"context"
	"io"
)

// SummarizeServiceInterface は要約サービスのインターフェース
type SummarizeServiceInterface interface {
	SummarizeText(ctx context.Context, text string) (*SummarizeResult, error)
	SummarizeFile(ctx context.Context, fileContent io.Reader, fileName string) (*SummarizeResult, error)
	SummarizeFileByS3Key(ctx context.Context, s3Key string) (*SummarizeResult, error)
	// 他の SummarizeService メソッドが必要であればここに追加
}

// TODO: SummarizeService が SummarizeServiceInterface を実装していることを静的にチェックする
// 例: var _ SummarizeServiceInterface = (*SummarizeService)(nil)
