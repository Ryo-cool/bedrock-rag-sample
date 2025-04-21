package services_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"bedrock-rag-sample/backend/config"
	"bedrock-rag-sample/backend/internal/services"
	"bedrock-rag-sample/backend/internal/services/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to build the expected RAG prompt based on current logic
func buildExpectedRAGPrompt(query string, docs []services.RetrievedDocument) string {
	var sb strings.Builder
	sb.WriteString("<human>")
	if len(docs) > 0 {
		sb.WriteString("以下は質問に関連する情報です:\n\n")
		for i, doc := range docs {
			sb.WriteString(fmt.Sprintf("文書[%d]:\n%s\n\n", i+1, doc.Content))
		}
	}
	sb.WriteString(fmt.Sprintf("質問: %s", query))
	sb.WriteString("</human>\n\n<assistant>")
	return sb.String()
}

func TestQAService_NewQAService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBedrockClient := mocks.NewMockBedrockClientInterface(ctrl)

	t.Run("正常系", func(t *testing.T) {
		cfg := &config.Config{
			AWS: config.AWSConfig{
				Region:          "ap-northeast-1",
				KnowledgeBaseID: "test-kb-id",
			},
		}
		qas, err := services.NewQAService(mockBedrockClient, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, qas)
	})

	t.Run("異常系_KB_IDなし", func(t *testing.T) {
		cfg := &config.Config{
			AWS: config.AWSConfig{
				Region: "ap-northeast-1",
				// KnowledgeBaseID: "", // KB ID is empty
			},
		}
		qas, err := services.NewQAService(mockBedrockClient, cfg)
		assert.Error(t, err)
		assert.Nil(t, qas)
		assert.Contains(t, err.Error(), "Knowledge Base IDが設定されていません")
	})

	// AWS設定読み込みエラーのテストは、awsconfig.LoadDefaultConfig をモック化する必要があり、やや複雑になるため省略
	// (または、実際のAWS認証情報がない環境で実行してエラーになることを確認する)
}

func TestQAService_SimpleRAG(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBedrockClient := mocks.NewMockBedrockClientInterface(ctrl)

	cfg := &config.Config{
		AWS: config.AWSConfig{
			Region:          "us-east-1",
			KnowledgeBaseID: "test-kb-id",
		},
	}

	// NewQAService を使ってインスタンスを生成 (bedrockClient はモック)
	qas, err := services.NewQAService(mockBedrockClient, cfg)
	require.NoError(t, err) // テストの前提条件としてエラーがないことを確認
	require.NotNil(t, qas)

	ctx := context.Background()

	t.Run("正常系_ドキュメントあり", func(t *testing.T) {
		query := "テストクエリです"
		expectedAnswer := "これはテスト回答です。"
		// retrieveDocuments は現在モックデータを返すので、そのデータを期待値として使う
		mockRetrievedDocs := []services.RetrievedDocument{
			{Content: "これはテスト用のモックドキュメントです。実際のKnowledge Baseからの検索結果が表示される予定です。", DocumentID: "mock-doc-001", Score: 0.95},
		}
		expectedPrompt := buildExpectedRAGPrompt(query, mockRetrievedDocs)

		// モックの設定: GenerateText が期待通り呼ばれるか
		mockBedrockClient.EXPECT().
			GenerateText(gomock.Any(), expectedPrompt).
			Return(expectedAnswer, nil).
			Times(1)

		result, err := qas.SimpleRAG(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, query, result.Query)
		assert.Equal(t, expectedAnswer, result.Answer)
		assert.Equal(t, mockRetrievedDocs, result.RetrievedDocuments)
	})

	t.Run("異常系_クエリが空", func(t *testing.T) {
		result, err := qas.SimpleRAG(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "クエリが空です")
	})

	t.Run("異常系_Bedrockエラー", func(t *testing.T) {
		query := "エラーを起こすクエリ"
		bedrockError := errors.New("Bedrock API error")
		mockRetrievedDocs := []services.RetrievedDocument{
			{Content: "これはテスト用のモックドキュメントです。実際のKnowledge Baseからの検索結果が表示される予定です。", DocumentID: "mock-doc-001", Score: 0.95},
		}
		expectedPrompt := buildExpectedRAGPrompt(query, mockRetrievedDocs)

		// モックの設定: GenerateText がエラーを返す
		mockBedrockClient.EXPECT().
			GenerateText(gomock.Any(), expectedPrompt).
			Return("", bedrockError).
			Times(1)

		result, err := qas.SimpleRAG(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, bedrockError) // エラー内容を具体的に確認する場合
		assert.Contains(t, err.Error(), "回答の生成に失敗しました")
	})

	// retrieveDocuments がエラーを返すケースのテストも追加可能だが、
	// 現在の実装ではエラーを無視して空のドキュメントで続行するため、
	// SimpleRAG レベルでは正常系と同じような動作になる。
	// retrieveDocuments 自体のテストでエラーハンドリングを確認するのが適切。
}

// TODO: retrieveDocuments のテストを追加
// (KB IDが空の場合、将来的に実際のAPI呼び出しをモック化した場合など)
