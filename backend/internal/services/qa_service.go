package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"bedrock-rag-sample/backend/config"
	"bedrock-rag-sample/backend/pkg/aws"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
)

// QAService はQ&A処理を行うサービス
type QAService struct {
	bedrockClient aws.BedrockClientInterface
	kbID          string
	agentClient   *bedrockagent.Client
}

// NewQAService は新しいQAServiceを作成する
func NewQAService(bedrockClient aws.BedrockClientInterface, cfg *config.Config) (*QAService, error) {
	// AWSクライアントの初期化
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	// Knowledge Base IDの確認
	if cfg.AWS.KnowledgeBaseID == "" {
		return nil, errors.New("Knowledge Base IDが設定されていません")
	}

	return &QAService{
		bedrockClient: bedrockClient,
		kbID:          cfg.AWS.KnowledgeBaseID,
		agentClient:   bedrockagent.NewFromConfig(awsCfg),
	}, nil
}

// RetrievedDocument は検索結果として取得されたドキュメント
type RetrievedDocument struct {
	Content    string  `json:"content"`
	Location   string  `json:"location,omitempty"`
	DocumentID string  `json:"document_id,omitempty"`
	Score      float64 `json:"score,omitempty"`
}

// QAResult はQ&A処理の結果
type QAResult struct {
	Query              string              `json:"query"`
	Answer             string              `json:"answer"`
	RetrievedDocuments []RetrievedDocument `json:"retrieved_documents,omitempty"`
}

// SimpleRAG はシンプルなRAG（Retrieval Augmented Generation）を実行する
// 直接BedrockのLLMを利用する簡易実装
func (s *QAService) SimpleRAG(ctx context.Context, query string) (*QAResult, error) {
	// クエリとシステムが空ではないことを確認
	if query == "" {
		return nil, errors.New("クエリが空です")
	}

	// 関連ドキュメントの検索
	docs, err := s.retrieveDocuments(ctx, query)
	if err != nil {
		// 検索エラーは記録するが処理は続行（ドキュメントなしで回答を生成）
		docs = []RetrievedDocument{}
	}

	// RAGプロンプトの構築
	ragPrompt := buildRAGPrompt(query, docs)

	// LLMで回答を生成
	answer, err := s.bedrockClient.GenerateText(ctx, ragPrompt)
	if err != nil {
		return nil, fmt.Errorf("回答の生成に失敗しました: %w", err)
	}

	return &QAResult{
		Query:              query,
		Answer:             answer,
		RetrievedDocuments: docs,
	}, nil
}

// retrieveDocuments はKnowledge Baseから関連ドキュメントを検索する
// 注: エラーが発生した場合は空の配列を返す
func (s *QAService) retrieveDocuments(ctx context.Context, query string) ([]RetrievedDocument, error) {
	// KBが設定されていない場合は空配列を返す
	if s.kbID == "" {
		return []RetrievedDocument{}, nil
	}

	// Knowledge Base API呼び出し (具体的なAPI呼び出しはインフラ構築後に実装)
	// 現在はモックデータを返す
	mockDoc := RetrievedDocument{
		Content:    "これはテスト用のモックドキュメントです。実際のKnowledge Baseからの検索結果が表示される予定です。",
		DocumentID: "mock-doc-001",
		Score:      0.95,
	}

	return []RetrievedDocument{mockDoc}, nil
}

// buildRAGPrompt はRAG用のプロンプトを構築する
func buildRAGPrompt(query string, docs []RetrievedDocument) string {
	var sb strings.Builder

	sb.WriteString("<human>")

	// コンテキスト情報が存在する場合は追加
	if len(docs) > 0 {
		sb.WriteString("以下は質問に関連する情報です:\n\n")

		for i, doc := range docs {
			sb.WriteString(fmt.Sprintf("文書[%d]:\n%s\n\n", i+1, doc.Content))
		}
	}

	// 質問を追加
	sb.WriteString(fmt.Sprintf("質問: %s", query))
	sb.WriteString("</human>\n\n<assistant>")

	return sb.String()
}
