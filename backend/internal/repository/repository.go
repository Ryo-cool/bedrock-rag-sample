package repository

import (
	"context"

	"bedrock-rag-sample/backend/internal/domain"
)

// DocumentRepository はドキュメント情報の永続化を担当するリポジトリのインターフェース
type DocumentRepository interface {
	// ドキュメント関連の操作
	GetDocumentByID(ctx context.Context, documentID int64) (*domain.Document, error)
	SaveDocument(ctx context.Context, doc *domain.Document) (int64, error)

	// Embedding関連の操作
	SaveDocumentEmbedding(ctx context.Context, documentID int64, chunkContent string, chunkIndex int, embedding []float32) (int64, error)
	FindSimilarChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]domain.DocumentChunk, error)

	// その他の必要なメソッド...

	// リソース管理
	Close() error
}
