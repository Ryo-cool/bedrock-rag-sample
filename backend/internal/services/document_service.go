package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"bedrock-rag-sample/backend/pkg/aws"
)

// DocumentService はドキュメント（PDF、画像）処理を行うサービス
type DocumentService struct {
	textractClient   *aws.TextractClient
	summarizeService *SummarizeService
}

// NewDocumentService は新しいDocumentServiceを作成する
func NewDocumentService(textractClient *aws.TextractClient, summarizeService *SummarizeService) *DocumentService {
	return &DocumentService{
		textractClient:   textractClient,
		summarizeService: summarizeService,
	}
}

// DocumentProcessResult はドキュメント処理結果
type DocumentProcessResult struct {
	OriginalText string             `json:"original_text"`
	Summary      string             `json:"summary,omitempty"`
	DocumentInfo aws.TextractResult `json:"document_info"`
	FileType     string             `json:"file_type"`
}

// ProcessDocument はドキュメントを処理し、テキスト抽出と要約を行う
func (s *DocumentService) ProcessDocument(ctx context.Context, file *multipart.FileHeader) (*DocumentProcessResult, error) {
	// ファイル拡張子を確認
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// サポートされる形式を確認
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".tiff" {
		return nil, fmt.Errorf("サポートされていないファイル形式です: %s", ext)
	}

	// Textractを使用してテキスト抽出
	extractResult, err := s.textractClient.ExtractTextFromDocument(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("テキスト抽出に失敗しました: %w", err)
	}

	// 抽出されたテキストが空でないか確認
	if extractResult.Text == "" {
		return nil, fmt.Errorf("ドキュメントからテキストを抽出できませんでした")
	}

	// 結果オブジェクトを作成
	result := &DocumentProcessResult{
		OriginalText: extractResult.Text,
		DocumentInfo: *extractResult,
		FileType:     ext[1:], // 先頭の.を削除
	}

	// テキストが十分な長さの場合は要約も生成
	// テキストが短い場合は要約を省略
	if len(extractResult.Text) > 200 {
		// 要約サービスを使用してテキスト要約
		summaryResult, err := s.summarizeService.SummarizeText(ctx, extractResult.Text)
		if err == nil && summaryResult.Summary != "" {
			result.Summary = summaryResult.Summary
		}
	}

	return result, nil
}
