package aws

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"bedrock-rag-sample/backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

// TextractClient はTextract操作のためのクライアント
type TextractClient struct {
	client     *textract.Client
	s3Client   *S3Client
	region     string
	bucketName string
}

// NewTextractClient は新しいTextractClientを作成する
func NewTextractClient(cfg *config.Config, s3Client *S3Client) (*TextractClient, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	client := textract.NewFromConfig(awsCfg)

	return &TextractClient{
		client:     client,
		s3Client:   s3Client,
		region:     cfg.AWS.Region,
		bucketName: cfg.AWS.S3BucketName,
	}, nil
}

// TextractResult はテキスト抽出結果
type TextractResult struct {
	Text       string `json:"text"`
	DocumentID string `json:"document_id,omitempty"`
	S3Key      string `json:"s3_key,omitempty"`
	Pages      int    `json:"pages,omitempty"`
}

// ExtractTextFromDocument はファイルからテキストを抽出する
func (t *TextractClient) ExtractTextFromDocument(ctx context.Context, file *multipart.FileHeader) (*TextractResult, error) {
	// ファイル拡張子を確認
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".tiff" {
		return nil, fmt.Errorf("サポートされていないファイル形式です: %s", ext)
	}

	// まずS3にアップロード
	s3Key, err := t.s3Client.UploadFile(ctx, file, "textract")
	if err != nil {
		return nil, fmt.Errorf("ファイルのアップロードに失敗しました: %w", err)
	}

	// Textractによるテキスト抽出（S3経由）
	docText, pages, err := t.extractTextFromS3(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("テキスト抽出に失敗しました: %w", err)
	}

	return &TextractResult{
		Text:       docText,
		S3Key:      s3Key,
		DocumentID: filepath.Base(file.Filename),
		Pages:      pages,
	}, nil
}

// extractTextFromS3 はS3上のドキュメントからテキストを抽出する
func (t *TextractClient) extractTextFromS3(ctx context.Context, s3Key string) (string, int, error) {
	// Textractが認識するS3パスを作成 (このinput変数は非同期処理では不要)
	/*
		input := &textract.DetectDocumentTextInput{
			Document: &types.Document{
				S3Object: &types.S3Object{
					Bucket: aws.String(t.bucketName),
					Name:   aws.String(s3Key),
				},
			},
		}
	*/

	// Textractにテキスト検出ジョブを送信
	startResp, err := t.client.StartDocumentTextDetection(ctx, &textract.StartDocumentTextDetectionInput{
		DocumentLocation: &types.DocumentLocation{
			S3Object: &types.S3Object{
				Bucket: aws.String(t.bucketName),
				Name:   aws.String(s3Key),
			},
		},
	})
	if err != nil {
		return "", 0, fmt.Errorf("textract検出に失敗しました: %w", err)
	}

	jobId := startResp.JobId

	// テキスト検出を実行
	output, err := t.client.GetDocumentTextDetection(ctx, &textract.GetDocumentTextDetectionInput{
		JobId: jobId,
	})
	if err != nil {
		return "", 0, fmt.Errorf("textract検出に失敗しました: %w", err)
	}

	// 結果を解析してテキストを抽出
	var sb strings.Builder
	var currentPage int32 = 1
	var maxPage int32 = 1

	for _, block := range output.Blocks {
		// ページ番号を確認（複数ページのPDFの場合）
		if block.Page != nil && *block.Page > maxPage {
			maxPage = *block.Page
		}

		// LINEタイプのブロックからテキストを取得
		if block.BlockType == types.BlockTypeLine {
			if block.Page != nil && *block.Page != currentPage {
				// ページが変わったら改ページを追加
				sb.WriteString("\n\n--- Page " + fmt.Sprintf("%d", *block.Page) + " ---\n\n")
				currentPage = *block.Page
			}

			sb.WriteString(*block.Text)
			sb.WriteString("\n")
		}
	}

	return sb.String(), int(maxPage), nil
}

// ExtractTextFromS3Key はS3上のドキュメントからテキストを抽出する（キーから直接）
func (t *TextractClient) ExtractTextFromS3Key(ctx context.Context, s3Key string) (*TextractResult, error) {
	docText, pages, err := t.extractTextFromS3(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("テキスト抽出に失敗しました: %w", err)
	}

	return &TextractResult{
		Text:       docText,
		S3Key:      s3Key,
		DocumentID: filepath.Base(s3Key),
		Pages:      pages,
	}, nil
}

// ドキュメント解析のより高度なバージョン（テーブル検出など）も追加できますが、
// 簡略化のため、ここではテキスト抽出のみを実装しています。
