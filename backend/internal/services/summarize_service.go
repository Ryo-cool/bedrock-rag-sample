package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"bedrock-rag-sample/backend/pkg/aws"
)

// SummarizeService はテキスト要約処理を行うサービス
type SummarizeService struct {
	bedrockClient aws.BedrockClientInterface
	uploadService *UploadService
}

// NewSummarizeService は新しいSummarizeServiceを作成する
func NewSummarizeService(bedrockClient *aws.BedrockClient, uploadService *UploadService) *SummarizeService {
	return &SummarizeService{
		bedrockClient: bedrockClient,
		uploadService: uploadService,
	}
}

// SummarizeResult は要約結果を表す構造体
type SummarizeResult struct {
	OriginalText string            `json:"original_text,omitempty"`
	Summary      string            `json:"summary"`
	SourceText   string            `json:"source_text,omitempty"`
	UploadInfo   *UploadFileResult `json:"upload_info,omitempty"`
}

// SummarizeText はテキストを要約する
func (s *SummarizeService) SummarizeText(ctx context.Context, text string) (*SummarizeResult, error) {
	if text == "" {
		return nil, errors.New("テキストが空です")
	}

	// テキストの長さを制限（例: 最大10000文字）
	if len(text) > 10000 {
		text = text[:10000]
	}

	// Bedrockを使って要約を生成
	summary, err := s.bedrockClient.GenerateSummary(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("要約の生成に失敗しました: %w", err)
	}

	return &SummarizeResult{
		Summary:    summary,
		SourceText: text,
	}, nil
}

// SummarizeFile はアップロードされたファイルから情報を抽出し、要約する
// 注: 現在はシンプルなテキストファイルのみをサポート (PDF/画像はTextractが必要)
func (s *SummarizeService) SummarizeFile(ctx context.Context, file *multipart.FileHeader) (*SummarizeResult, error) {
	// まずファイルをアップロード
	uploadResult, err := s.uploadService.UploadFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("ファイルのアップロードに失敗しました: %w", err)
	}

	// ファイルからテキストを抽出 (現在はシンプルな方法で)
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	}
	defer src.Close()

	// ファイルの内容を読み込む (テキストファイルと仮定)
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	text := string(content)

	// テキストの長さを制限
	if len(text) > 10000 {
		text = text[:10000]
	}

	// Bedrockを使って要約を生成
	summary, err := s.bedrockClient.GenerateSummary(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("要約の生成に失敗しました: %w", err)
	}

	return &SummarizeResult{
		Summary:    summary,
		SourceText: text,
		UploadInfo: uploadResult,
	}, nil
}

// SummarizeFileByS3Key はS3キーで指定されたファイルを要約する (新規追加)
func (s *SummarizeService) SummarizeFileByS3Key(ctx context.Context, s3Key string) (*SummarizeResult, error) {
	// S3からファイルをダウンロード
	fileContent, err := s.uploadService.s3Client.DownloadFileContent(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("S3からのファイルダウンロードに失敗しました (key: %s): %w", s3Key, err)
	}

	// テキストを要約
	summary, err := s.bedrockClient.GenerateSummary(ctx, string(fileContent))
	if err != nil {
		return nil, fmt.Errorf("Bedrockでのファイル要約に失敗しました (key: %s): %w", s3Key, err)
	}

	return &SummarizeResult{Summary: summary}, nil // OriginalTextは含めない（任意）
}
