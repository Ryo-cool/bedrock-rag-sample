package services_test

import (
	"context"
	"errors"
	"testing"

	"bedrock-rag-sample/backend/internal/domain"
	domainmocks "bedrock-rag-sample/backend/internal/domain/mocks" // DB モック
	"bedrock-rag-sample/backend/internal/services"
	servicemocks "bedrock-rag-sample/backend/internal/services/mocks" // Bedrock モック

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendService_FindSimilarDocuments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBedrockClient := servicemocks.NewMockBedrockClientInterface(ctrl)
	mockDBHandler := domainmocks.NewMockDBHandlerInterface(ctrl)

	// テスト対象サービス生成
	recommendService := services.NewRecommendService(mockBedrockClient, mockDBHandler)

	ctx := context.Background()
	query := "類似文書を探すクエリ"
	limit := 3
	queryEmbedding := []float32{0.1, 0.2, 0.3}
	similarChunks := []domain.DocumentChunk{
		{ID: 1, DocumentID: 10, Content: "類似チャンク1", ChunkIndex: 0, Similarity: 0.9},
		{ID: 2, DocumentID: 20, Content: "類似チャンク2", ChunkIndex: 1, Similarity: 0.8},
		{ID: 3, DocumentID: 10, Content: "類似チャンク3", ChunkIndex: 1, Similarity: 0.7}, // Doc 10 again
	}
	doc10 := &domain.Document{ID: 10, Filename: "文書10", Content: "..."}
	doc20 := &domain.Document{ID: 20, Filename: "文書20", Content: "..."}

	t.Run("正常系", func(t *testing.T) {
		// --- モックの期待動作設定 ---
		// 1. クエリの Embedding 生成
		mockBedrockClient.EXPECT().
			GenerateEmbedding(ctx, query).
			Return(queryEmbedding, nil).
			Times(1)

		// 2. 類似チャンク検索
		mockDBHandler.EXPECT().
			FindSimilarChunks(ctx, queryEmbedding, limit).
			Return(similarChunks, nil).
			Times(1)

		// 3. ドキュメント情報の取得 (重複排除されているか確認)
		mockDBHandler.EXPECT().
			GetDocumentByID(ctx, int64(10)). // Doc ID 10
			Return(doc10, nil).
			Times(1) // 一度だけ呼ばれるはず
		mockDBHandler.EXPECT().
			GetDocumentByID(ctx, int64(20)). // Doc ID 20
			Return(doc20, nil).
			Times(1)

		// --- テスト実行 ---
		result, err := recommendService.FindSimilarDocuments(ctx, query, limit)

		// --- アサーション ---
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, query, result.Query)
		assert.Equal(t, similarChunks, result.RecommendedChunks)
		require.Len(t, result.Documents, 2) // ドキュメントは2つのはず
		assert.Equal(t, doc10, result.Documents[10])
		assert.Equal(t, doc20, result.Documents[20])
	})

	t.Run("異常系_Embedding生成エラー", func(t *testing.T) {
		embeddingError := errors.New("embedding error")
		mockBedrockClient.EXPECT().
			GenerateEmbedding(ctx, query).
			Return(nil, embeddingError).
			Times(1)

		result, err := recommendService.FindSimilarDocuments(ctx, query, limit)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, embeddingError)
		assert.Contains(t, err.Error(), "クエリのEmbedding生成に失敗しました")
	})

	t.Run("異常系_チャンク検索エラー", func(t *testing.T) {
		findError := errors.New("find error")
		mockBedrockClient.EXPECT().
			GenerateEmbedding(ctx, query).
			Return(queryEmbedding, nil).
			Times(1)
		mockDBHandler.EXPECT().
			FindSimilarChunks(ctx, queryEmbedding, limit).
			Return(nil, findError).
			Times(1)

		result, err := recommendService.FindSimilarDocuments(ctx, query, limit)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, findError)
		assert.Contains(t, err.Error(), "類似チャンクの検索に失敗しました")
	})

	// GetDocumentByID エラーケースは、エラーが発生しても結果には含まれないだけなので、
	// ハンドラーレベルで重要なエラーでなければ、このレベルでのテストは省略可能。
	// 必要であれば追加。
}

// TODO: ProcessDocumentForEmbedding のテストを追加
// TODO: splitIntoChunks のテストを追加 (これは独立してテスト可能)
