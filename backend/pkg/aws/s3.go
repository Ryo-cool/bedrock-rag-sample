package aws

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	appconfig "bedrock-rag-sample/backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client はS3操作のためのクライアント
type S3Client struct {
	client   *s3.Client
	bucket   string
	basePath string
}

// NewS3Client は新しいS3クライアントを作成する
func NewS3Client(cfg *appconfig.Config) (*S3Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Client{
		client:   client,
		bucket:   cfg.AWS.S3BucketName,
		basePath: cfg.AWS.S3DocumentsPath,
	}, nil
}

// UploadFile はファイルをS3にアップロードする
func (s *S3Client) UploadFile(ctx context.Context, file *multipart.FileHeader, customPath string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	}
	defer src.Close()

	// ファイル名を取得
	filename := filepath.Base(file.Filename)

	// S3内のパスを構築
	s3Path := s.basePath
	if customPath != "" {
		s3Path = filepath.Join(s3Path, customPath)
	}
	s3Key := filepath.Join(s3Path, filename)

	// S3にアップロード
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
		Body:   src,
	})

	if err != nil {
		return "", fmt.Errorf("S3へのアップロードに失敗しました: %w", err)
	}

	// S3内のファイルへのパスを返す
	return s3Key, nil
}

// GetFileURL はS3内のファイルへのアクセスURLを生成する
func (s *S3Client) GetFileURL(ctx context.Context, key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", fmt.Errorf("署名付きURLの生成に失敗しました: %w", err)
	}

	return presignedReq.URL, nil
}
