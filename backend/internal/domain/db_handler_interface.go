package domain

import (
	"context"
)

// DBHandlerInterface はデータベース操作のためのインターフェース
type DBHandlerInterface interface {
	SaveDocumentEmbedding(ctx context.Context, documentID int64, chunk string, chunkIndex int, embedding []float32) (int64, error)
	FindSimilarChunks(ctx context.Context, embedding []float32, limit int) ([]DocumentChunk, error)
	GetDocumentByID(ctx context.Context, documentID int64) (*Document, error)
	// 他の DBHandler メソッドが必要であればここに追加
}

// TODO: DBHandler が DBHandlerInterface を実装していることを静的にチェックする
// 例: var _ DBHandlerInterface = (*DBHandler)(nil)
// (DBHandler の実装を確認してから追加)
