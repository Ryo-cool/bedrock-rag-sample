package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"bedrock-rag-sample/backend/pkg/aws"
)

// UploadService はファイルアップロード処理を行うサービス
type UploadService struct {
	s3Client *aws.S3Client
}

// NewUploadService は新しいUploadServiceを作成する
func NewUploadService(s3Client *aws.S3Client) *UploadService {
	return &UploadService{
		s3Client: s3Client,
	}
}

// UploadFileResult はアップロード結果を表す構造体
type UploadFileResult struct {
	Key      string `json:"key"`
	Filename string `json:"filename"`
	URL      string `json:"url,omitempty"`
}

// UploadFile はファイルをS3にアップロードする
func (s *UploadService) UploadFile(ctx context.Context, file *multipart.FileHeader) (*UploadFileResult, error) {
	// ファイルの拡張子を取得
	ext := filepath.Ext(file.Filename)
	var folderPath string

	// ファイル種類に応じてフォルダを分ける
	switch ext {
	case ".pdf":
		folderPath = "pdf"
	case ".png", ".jpg", ".jpeg":
		folderPath = "images"
	default:
		folderPath = "others"
	}

	// S3にアップロード
	key, err := s.s3Client.UploadFile(ctx, file, folderPath)
	if err != nil {
		return nil, fmt.Errorf("ファイルのアップロードに失敗しました: %w", err)
	}

	// 署名付きURLを生成
	url, err := s.s3Client.GetFileURL(ctx, key)
	if err != nil {
		// URLの生成に失敗してもアップロード自体は成功しているので、エラーはログに記録するだけにする
		url = ""
	}

	return &UploadFileResult{
		Key:      key,
		Filename: file.Filename,
		URL:      url,
	}, nil
}
