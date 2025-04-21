package services_test

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/textproto"
	"testing"

	"bedrock-rag-sample/backend/internal/services"
	servicemocks "bedrock-rag-sample/backend/internal/services/mocks" // Bedrock, Upload モック
	awsmock "bedrock-rag-sample/backend/pkg/aws/mock"                 // S3 モック

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarizeService_SummarizeText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBedrockClient := servicemocks.NewMockBedrockClientInterface(ctrl)
	mockUploadService := servicemocks.NewMockUploadServiceInterface(ctrl) // このテストでは使わないが初期化

	summarizeService := services.NewSummarizeService(mockBedrockClient, mockUploadService)

	ctx := context.Background()
	inputText := "これは要約対象の長いテキストです。"
	expectedSummary := "要約結果"

	t.Run("正常系", func(t *testing.T) {
		mockBedrockClient.EXPECT().
			GenerateSummary(ctx, inputText).
			Return(expectedSummary, nil).
			Times(1)

		result, err := summarizeService.SummarizeText(ctx, inputText)

		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedSummary, result.Summary)
		assert.Equal(t, inputText, result.SourceText) // 元テキストも含まれることを確認
	})

	t.Run("異常系_テキストが空", func(t *testing.T) {
		result, err := summarizeService.SummarizeText(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "テキストが空です")
	})

	t.Run("異常系_Bedrockエラー", func(t *testing.T) {
		bedrockError := errors.New("bedrock summary error")
		mockBedrockClient.EXPECT().
			GenerateSummary(ctx, inputText).
			Return("", bedrockError).
			Times(1)

		result, err := summarizeService.SummarizeText(ctx, inputText)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, bedrockError)
		assert.Contains(t, err.Error(), "要約の生成に失敗しました")
	})
}

func TestSummarizeService_SummarizeFileByS3Key(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBedrockClient := servicemocks.NewMockBedrockClientInterface(ctrl)
	mockUploadService := servicemocks.NewMockUploadServiceInterface(ctrl)
	mockS3Client := awsmock.NewMockS3ClientInterface(ctrl) // S3 モックも必要

	summarizeService := services.NewSummarizeService(mockBedrockClient, mockUploadService)

	ctx := context.Background()
	s3Key := "path/to/file.txt"
	fileContent := []byte("S3からダウンロードしたファイルの内容です。")
	expectedSummary := "S3ファイル要約結果"

	t.Run("正常系", func(t *testing.T) {
		// --- モックの期待動作設定 ---
		// 1. UploadService から S3 クライアントを取得
		mockUploadService.EXPECT().
			GetS3Client().
			Return(mockS3Client). // モック S3 クライアントを返す
			Times(1)
		// 2. S3 からファイルをダウンロード
		mockS3Client.EXPECT().
			DownloadFileContent(ctx, s3Key).
			Return(fileContent, nil).
			Times(1)
		// 3. Bedrock で要約を生成
		mockBedrockClient.EXPECT().
			GenerateSummary(ctx, string(fileContent)).
			Return(expectedSummary, nil).
			Times(1)

		// --- テスト実行 ---
		result, err := summarizeService.SummarizeFileByS3Key(ctx, s3Key)

		// --- アサーション ---
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedSummary, result.Summary)
	})

	t.Run("異常系_S3ダウンロードエラー", func(t *testing.T) {
		downloadError := errors.New("s3 download error")
		mockUploadService.EXPECT().GetS3Client().Return(mockS3Client).Times(1)
		mockS3Client.EXPECT().
			DownloadFileContent(ctx, s3Key).
			Return(nil, downloadError).
			Times(1)

		result, err := summarizeService.SummarizeFileByS3Key(ctx, s3Key)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, downloadError)
		assert.Contains(t, err.Error(), "S3からのファイルダウンロードに失敗しました")
	})

	t.Run("異常系_Bedrock要約エラー", func(t *testing.T) {
		summaryError := errors.New("bedrock summary error")
		mockUploadService.EXPECT().GetS3Client().Return(mockS3Client).Times(1)
		mockS3Client.EXPECT().
			DownloadFileContent(ctx, s3Key).
			Return(fileContent, nil).
			Times(1)
		mockBedrockClient.EXPECT().
			GenerateSummary(ctx, string(fileContent)).
			Return("", summaryError).
			Times(1)

		result, err := summarizeService.SummarizeFileByS3Key(ctx, s3Key)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, summaryError)
		assert.Contains(t, err.Error(), "Bedrockでのファイル要約に失敗しました")
	})
}

// --- ヘルパー関数: multipart.FileHeader のモックを作成 ---
// (SummarizeFile のテストで使用する可能性があるため残す)
type mockMultipartFileHeader struct {
	filename string
	content  []byte
	header   textproto.MIMEHeader
}

func (mfh *mockMultipartFileHeader) Open() (multipart.File, error) {
	return &mockMultipartFile{bytes.NewReader(mfh.content)}, nil
}

type mockMultipartFile struct {
	*bytes.Reader
}

func (mf *mockMultipartFile) Close() error {
	return nil // bytes.Reader doesn't need closing
}

func createMockFileHeader(filename string, content string) *multipart.FileHeader {
	// このヘルパーは SummarizeFile のテストで必要に応じて調整・使用します。
	// 現状のテストでは直接使用されていません。
	return &multipart.FileHeader{
		Filename: filename,
		Header:   textproto.MIMEHeader{},
		// Size: int64(len(content)),
	}
}

// --- ここまでヘルパー関数 ---

// TODO: SummarizeFile のテストを追加 (multipart.FileHeader の扱いと UploadFile モックが必要)
