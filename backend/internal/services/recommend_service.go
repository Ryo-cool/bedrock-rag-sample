package services

import (
	"context"
	"fmt"
	"strings"

	"bedrock-rag-sample/backend/internal/domain"
	"bedrock-rag-sample/backend/pkg/aws"
)

// RecommendService はテキスト類似度に基づく推薦を行うサービス
type RecommendService struct {
	bedrockClient aws.BedrockClientInterface
	dbHandler     domain.DBHandlerInterface
}

// NewRecommendService は新しいRecommendServiceを作成する
func NewRecommendService(bedrockClient aws.BedrockClientInterface, dbHandler domain.DBHandlerInterface) *RecommendService {
	return &RecommendService{
		bedrockClient: bedrockClient,
		dbHandler:     dbHandler,
	}
}

// ChunkSize は分割するチャンクのサイズ
const ChunkSize = 1000

// MaxChunks は1つのドキュメントから作成する最大チャンク数
const MaxChunks = 20

// RecommendResult は推薦結果を表す構造体
type RecommendResult struct {
	Query             string                     `json:"query"`
	RecommendedChunks []domain.DocumentChunk     `json:"recommended_chunks"`
	Documents         map[int64]*domain.Document `json:"documents,omitempty"`
}

// ProcessDocumentForEmbedding はドキュメントをチャンクに分割し、Embeddingを生成する
func (s *RecommendService) ProcessDocumentForEmbedding(ctx context.Context, doc *domain.Document) error {
	// ドキュメントをチャンクに分割
	chunks := s.splitIntoChunks(doc.Content)

	// 各チャンクのEmbeddingを生成し保存
	for i, chunk := range chunks {
		// Embeddingを生成
		embedding, err := s.bedrockClient.GenerateEmbedding(ctx, chunk)
		if err != nil {
			return fmt.Errorf("embedding生成に失敗しました: %w", err)
		}

		// 生成したEmbeddingを保存
		_, err = s.dbHandler.SaveDocumentEmbedding(ctx, doc.ID, chunk, i, embedding)
		if err != nil {
			return fmt.Errorf("embeddingの保存に失敗しました: %w", err)
		}
	}

	return nil
}

// FindSimilarDocuments はクエリに類似したドキュメントを検索する
func (s *RecommendService) FindSimilarDocuments(ctx context.Context, query string, limit int) (*RecommendResult, error) {
	if limit <= 0 {
		limit = 5 // デフォルト値
	}

	// クエリのEmbeddingを生成
	queryEmbedding, err := s.bedrockClient.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("クエリのEmbedding生成に失敗しました: %w", err)
	}

	// 類似したチャンクを検索
	chunks, err := s.dbHandler.FindSimilarChunks(ctx, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("類似チャンクの検索に失敗しました: %w", err)
	}

	// 結果を構築
	result := &RecommendResult{
		Query:             query,
		RecommendedChunks: chunks,
		Documents:         make(map[int64]*domain.Document),
	}

	// ドキュメント情報を取得 (重複を除く)
	docIDs := make(map[int64]bool)
	for _, chunk := range chunks {
		if !docIDs[chunk.DocumentID] {
			docIDs[chunk.DocumentID] = true
			doc, err := s.dbHandler.GetDocumentByID(ctx, chunk.DocumentID)
			if err == nil {
				result.Documents[chunk.DocumentID] = doc
			}
		}
	}

	return result, nil
}

// splitIntoChunks はテキストをチャンクに分割する
func (s *RecommendService) splitIntoChunks(text string) []string {
	var chunks []string

	// 単純な行分割（より洗練された分割方法も実装可能）
	paragraphs := strings.Split(text, "\n\n")

	var currentChunk strings.Builder

	for _, para := range paragraphs {
		// 段落が空でない場合のみ処理
		trimmedPara := strings.TrimSpace(para)
		if trimmedPara == "" {
			continue
		}

		// 現在のチャンクに追加すると長すぎる場合は、新しいチャンクを開始
		if currentChunk.Len()+len(trimmedPara) > ChunkSize {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()

				// 最大チャンク数に達した場合は終了
				if len(chunks) >= MaxChunks {
					break
				}
			}
		}

		// 段落を追加
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(trimmedPara)
	}

	// 最後のチャンクを追加
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
