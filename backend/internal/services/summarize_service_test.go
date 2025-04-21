package services_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
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
		assert.Contains(t, err.Error(), "bedrockでのファイル要約に失敗しました")
	})
}

func TestSummarizeService_SummarizeFile(t *testing.T) {
	// スキップを解除
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBedrockClient := servicemocks.NewMockBedrockClientInterface(ctrl)
	// mockUploadService は不要になった
	// mockUploadService := servicemocks.NewMockUploadServiceInterface(ctrl)

	// SummarizeService の生成 (uploadService は nil で OK)
	summarizeService := services.NewSummarizeService(mockBedrockClient, nil)

	ctx := context.Background()
	fileName := "test_summarize.txt"
	fileContent := "これは io.Reader から読み込まれるファイルの内容です。"
	expectedSummary := "リーダー内容の要約"

	t.Run("正常系", func(t *testing.T) {
		// --- モックの期待動作設定 ---
		// GenerateSummary がファイル内容で呼ばれる
		mockBedrockClient.EXPECT().
			GenerateSummary(ctx, fileContent).
			Return(expectedSummary, nil).
			Times(1)

		// --- テスト実行 ---
		// bytes.NewReader を使って io.Reader を作成
		reader := bytes.NewReader([]byte(fileContent))
		result, err := summarizeService.SummarizeFile(ctx, reader, fileName)

		// --- アサーション ---
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, expectedSummary, result.Summary)
		assert.Equal(t, fileContent, result.SourceText)
		assert.Nil(t, result.UploadInfo) // UploadInfo は nil になっているはず
	})

	t.Run("異常系_ファイル内容が空", func(t *testing.T) {
		reader := bytes.NewReader([]byte("")) // 空の内容
		result, err := summarizeService.SummarizeFile(ctx, reader, fileName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ファイルの内容が空です")
	})

	t.Run("異常系_GenerateSummaryエラー", func(t *testing.T) {
		summaryError := errors.New("summary failed")
		// GenerateSummary でエラーが発生
		mockBedrockClient.EXPECT().
			GenerateSummary(ctx, fileContent).
			Return("", summaryError).
			Times(1)

		reader := bytes.NewReader([]byte(fileContent))
		result, err := summarizeService.SummarizeFile(ctx, reader, fileName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, summaryError)
		assert.Contains(t, err.Error(), "要約の生成に失敗しました")
	})

	t.Run("エッジケース_テキスト長制限", func(t *testing.T) {
		longContent := strings.Repeat("a", 10001)
		truncatedContent := strings.Repeat("a", 10000)

		// 制限されたテキストで GenerateSummary が呼ばれることを確認
		mockBedrockClient.EXPECT().GenerateSummary(ctx, truncatedContent).Return("summary", nil).Times(1)

		reader := bytes.NewReader([]byte(longContent))
		result, err := summarizeService.SummarizeFile(ctx, reader, fileName)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, truncatedContent, result.SourceText) // SourceText も制限されていることを確認
		assert.Equal(t, "summary", result.Summary)
	})
}
