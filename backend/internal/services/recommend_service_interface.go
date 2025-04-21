package services

import (
	"context"

	"bedrock-rag-sample/backend/internal/domain"
)

// RecommendServiceInterface は推薦サービスのインターフェース
type RecommendServiceInterface interface {
	ProcessDocumentForEmbedding(ctx context.Context, doc *domain.Document) error
	FindSimilarDocuments(ctx context.Context, query string, limit int) (*RecommendResult, error)
	// 他の RecommendService メソッドが必要であればここに追加
}

// TODO: RecommendService が RecommendServiceInterface を実装していることを静的にチェックする
// 例: var _ RecommendServiceInterface = (*RecommendService)(nil)
