package services

import (
	"context"
	"errors"
	"fmt"
	"io"

	"bedrock-rag-sample/backend/pkg/aws"
)

// SummarizeService はテキスト要約処理を行うサービス
type SummarizeService struct {
	bedrockClient aws.BedrockClientInterface
	uploadService UploadServiceInterface
}

// NewSummarizeService は新しいSummarizeServiceを作成する
func NewSummarizeService(bedrockClient aws.BedrockClientInterface, uploadService UploadServiceInterface) *SummarizeService {
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

// SummarizeFile は io.Reader から内容を読み込み、要約する
// 引数: fileContent (ファイル内容), fileName (拡張子判定用)
func (s *SummarizeService) SummarizeFile(ctx context.Context, fileContent io.Reader, fileName string) (*SummarizeResult, error) {
	// // まずファイルをアップロード // Upload 処理を削除
	// uploadResult, err := s.uploadService.UploadFile(ctx, file)
	// if err != nil {
	// 	return nil, fmt.Errorf("ファイルのアップロードに失敗しました: %w", err)
	// }

	// // ファイルからテキストを抽出 (現在はシンプルな方法で) // fileHeader からの読み込みを削除
	// src, err := file.Open()
	// if err != nil {
	// 	return nil, fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	// }
	// defer src.Close()

	// ファイルの内容を読み込む (引数の Reader から読み込む)
	contentBytes, err := io.ReadAll(fileContent) // fileContent を直接読む
	if err != nil {
		return nil, fmt.Errorf("ファイル内容の読み込みに失敗しました: %w", err)
	}
	text := string(contentBytes)

	// テキストが空かチェック (追加)
	if text == "" {
		return nil, errors.New("ファイルの内容が空です")
	}

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
		// UploadInfo は削除 (このメソッドはアップロードしなくなったため)
		// UploadInfo: uploadResult,
	}, nil
}

// SummarizeFileByS3Key はS3キーで指定されたファイルを要約する (新規追加)
func (s *SummarizeService) SummarizeFileByS3Key(ctx context.Context, s3Key string) (*SummarizeResult, error) {
	// uploadService から S3 クライアントを取得
	s3Client := s.uploadService.GetS3Client()
	if s3Client == nil {
		// S3 クライアントが取得できない場合のエラーハンドリング (例)
		return nil, errors.New("S3 client is not available through upload service")
	}

	// S3からファイルをダウンロード
	fileContent, err := s3Client.DownloadFileContent(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("S3からのファイルダウンロードに失敗しました (key: %s): %w", s3Key, err)
	}

	// テキストを要約
	summary, err := s.bedrockClient.GenerateSummary(ctx, string(fileContent))
	if err != nil {
		return nil, fmt.Errorf("bedrockでのファイル要約に失敗しました (key: %s): %w", s3Key, err)
	}

	return &SummarizeResult{Summary: summary}, nil // OriginalTextは含めない（任意）
}
