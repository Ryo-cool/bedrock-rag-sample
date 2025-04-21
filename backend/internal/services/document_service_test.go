package services

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"bedrock-rag-sample/backend/pkg/aws"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TextractClientのモック
type mockTextractClient struct {
	mock.Mock
}

func (m *mockTextractClient) ExtractTextFromDocument(ctx context.Context, file *multipart.FileHeader) (*aws.TextractResult, error) {
	args := m.Called(ctx, file)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aws.TextractResult), args.Error(1)
}

func (m *mockTextractClient) ExtractTextFromS3Key(ctx context.Context, s3Key string) (*aws.TextractResult, error) {
	args := m.Called(ctx, s3Key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aws.TextractResult), args.Error(1)
}

// インターフェースを実装していることを確認
var _ aws.TextractClientInterface = (*mockTextractClient)(nil)

// SummarizeServiceのモック
type mockSummarizeService struct {
	mock.Mock
}

func (m *mockSummarizeService) SummarizeText(ctx context.Context, text string) (*SummarizeResult, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SummarizeResult), args.Error(1)
}

func (m *mockSummarizeService) SummarizeFile(ctx context.Context, file *multipart.FileHeader) (*SummarizeResult, error) {
	args := m.Called(ctx, file)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SummarizeResult), args.Error(1)
}

// インターフェースを実装していることを確認
var _ SummarizeServiceInterface = (*mockSummarizeService)(nil)

func TestProcessDocument(t *testing.T) {
	longText := "これはテストドキュメントの内容です。十分な長さがあるので要約も生成されます。これはテストです。テストです。テストです。繰り返しです。長い文章です。"
	longErrorText := "これはテストドキュメントの内容です。十分な長さがあるので要約も生成されますが、要約に失敗します。これはテストです。テストです。テストです。繰り返しです。長い文章です。"

	testCases := []struct {
		name              string
		fileHeader        *multipart.FileHeader
		textractResult    *aws.TextractResult
		textractErr       error
		summarizeResult   *SummarizeResult
		summarizeErr      error
		expectedResult    *DocumentProcessResult
		expectError       bool
		shouldCallSummary bool
	}{
		{
			name: "PDFファイルの処理成功（要約あり）",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			textractResult: &aws.TextractResult{
				Text:       longText,
				DocumentID: "test.pdf",
				S3Key:      "documents/test.pdf",
				Pages:      1,
			},
			textractErr: nil,
			summarizeResult: &SummarizeResult{
				Summary:    "テストドキュメントの要約",
				SourceText: "元のテキスト",
			},
			summarizeErr:      nil,
			shouldCallSummary: true,
			expectedResult: &DocumentProcessResult{
				OriginalText: longText,
				Summary:      "テストドキュメントの要約",
				DocumentInfo: aws.TextractResult{
					Text:       longText,
					DocumentID: "test.pdf",
					S3Key:      "documents/test.pdf",
					Pages:      1,
				},
				FileType: "pdf",
			},
			expectError: false,
		},
		{
			name: "PDFファイルの処理成功（テキストが短いので要約なし）",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			textractResult: &aws.TextractResult{
				Text:       "短いテキスト",
				DocumentID: "test.pdf",
				S3Key:      "documents/test.pdf",
				Pages:      1,
			},
			textractErr:       nil,
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult: &DocumentProcessResult{
				OriginalText: "短いテキスト",
				Summary:      "",
				DocumentInfo: aws.TextractResult{
					Text:       "短いテキスト",
					DocumentID: "test.pdf",
					S3Key:      "documents/test.pdf",
					Pages:      1,
				},
				FileType: "pdf",
			},
			expectError: false,
		},
		{
			name: "サポートされていないファイル形式",
			fileHeader: &multipart.FileHeader{
				Filename: "test.docx",
				Size:     1024,
			},
			textractResult:    nil,
			textractErr:       nil,
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
		{
			name: "テキスト抽出エラー",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			textractResult:    nil,
			textractErr:       errors.New("テキスト抽出エラー"),
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
		{
			name: "テキスト抽出は成功したが空文字",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			textractResult: &aws.TextractResult{
				Text:       "",
				DocumentID: "test.pdf",
				S3Key:      "documents/test.pdf",
				Pages:      1,
			},
			textractErr:       nil,
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
		{
			name: "要約エラーでも処理は成功する",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			textractResult: &aws.TextractResult{
				Text:       longErrorText,
				DocumentID: "test.pdf",
				S3Key:      "documents/test.pdf",
				Pages:      1,
			},
			textractErr:       nil,
			summarizeResult:   &SummarizeResult{},
			summarizeErr:      errors.New("要約エラー"),
			shouldCallSummary: true,
			expectedResult: &DocumentProcessResult{
				OriginalText: longErrorText,
				Summary:      "",
				DocumentInfo: aws.TextractResult{
					Text:       longErrorText,
					DocumentID: "test.pdf",
					S3Key:      "documents/test.pdf",
					Pages:      1,
				},
				FileType: "pdf",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックを作成
			mockTextract := new(mockTextractClient)
			mockSummarize := new(mockSummarizeService)

			// DocumentServiceを作成
			docService := &DocumentService{
				textractClient:   mockTextract,
				summarizeService: mockSummarize,
			}

			// モックの振る舞いを設定
			if tc.fileHeader.Filename != "test.docx" {
				mockTextract.On("ExtractTextFromDocument", mock.Anything, tc.fileHeader).Return(tc.textractResult, tc.textractErr)
			}

			if tc.shouldCallSummary {
				mockSummarize.On("SummarizeText", mock.Anything, tc.textractResult.Text).Return(tc.summarizeResult, tc.summarizeErr)
			}

			// テスト実行
			result, err := docService.ProcessDocument(context.Background(), tc.fileHeader)

			// 検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedResult.OriginalText, result.OriginalText)
				assert.Equal(t, tc.expectedResult.Summary, result.Summary)
				assert.Equal(t, tc.expectedResult.FileType, result.FileType)
				// DocumentInfoの詳細な検証は省略
			}

			// すべての期待されるメソッド呼び出しが行われたことを確認
			mockTextract.AssertExpectations(t)
			mockSummarize.AssertExpectations(t)
		})
	}
}

func TestProcessDocumentByS3Key(t *testing.T) {
	longText := "これはS3上のテストドキュメントの内容です。十分な長さがあるので要約も生成されます。これはテストです。テストです。テストです。繰り返しです。長い文章です。"

	testCases := []struct {
		name              string
		s3Key             string
		textractResult    *aws.TextractResult
		textractErr       error
		summarizeResult   *SummarizeResult
		summarizeErr      error
		expectedResult    *DocumentProcessResult
		expectError       bool
		shouldCallSummary bool
	}{
		{
			name:  "S3キーからPDFファイルの処理成功（要約あり）",
			s3Key: "documents/pdf/test.pdf",
			textractResult: &aws.TextractResult{
				Text:       longText,
				DocumentID: "test.pdf",
				S3Key:      "documents/pdf/test.pdf",
				Pages:      1,
			},
			textractErr: nil,
			summarizeResult: &SummarizeResult{
				Summary:    "S3ドキュメントの要約",
				SourceText: "元のテキスト",
			},
			summarizeErr:      nil,
			shouldCallSummary: true,
			expectedResult: &DocumentProcessResult{
				OriginalText: longText,
				Summary:      "S3ドキュメントの要約",
				DocumentInfo: aws.TextractResult{
					Text:       longText,
					DocumentID: "test.pdf",
					S3Key:      "documents/pdf/test.pdf",
					Pages:      1,
				},
				FileType: "pdf",
			},
			expectError: false,
		},
		{
			name:              "サポートされていないファイル形式",
			s3Key:             "documents/docx/test.docx",
			textractResult:    nil,
			textractErr:       nil,
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
		{
			name:              "S3キーからのテキスト抽出エラー",
			s3Key:             "documents/pdf/error.pdf",
			textractResult:    nil,
			textractErr:       errors.New("テキスト抽出エラー"),
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックを作成
			mockTextract := new(mockTextractClient)
			mockSummarize := new(mockSummarizeService)

			// DocumentServiceを作成
			docService := &DocumentService{
				textractClient:   mockTextract,
				summarizeService: mockSummarize,
			}

			// モックの振る舞いを設定
			if tc.s3Key != "documents/docx/test.docx" {
				mockTextract.On("ExtractTextFromS3Key", mock.Anything, tc.s3Key).Return(tc.textractResult, tc.textractErr)
			}

			if tc.shouldCallSummary {
				mockSummarize.On("SummarizeText", mock.Anything, tc.textractResult.Text).Return(tc.summarizeResult, tc.summarizeErr)
			}

			// テスト実行
			result, err := docService.ProcessDocumentByS3Key(context.Background(), tc.s3Key)

			// 検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedResult.OriginalText, result.OriginalText)
				assert.Equal(t, tc.expectedResult.Summary, result.Summary)
				assert.Equal(t, tc.expectedResult.FileType, result.FileType)
				// DocumentInfoの詳細な検証は省略
			}

			// すべての期待されるメソッド呼び出しが行われたことを確認
			mockTextract.AssertExpectations(t)
			mockSummarize.AssertExpectations(t)
		})
	}
}
