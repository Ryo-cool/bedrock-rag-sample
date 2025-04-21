package services

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"bedrock-rag-sample/backend/internal/servicemock"
	"bedrock-rag-sample/backend/pkg/aws"
	awsmock "bedrock-rag-sample/backend/pkg/aws/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// モックアダプター: servicemock.SummarizeResult から services.SummarizeResult への変換
type summarizeServiceAdapter struct {
	mockService *servicemock.MockSummarizeServiceInterface
}

func newSummarizeServiceAdapter(mockService *servicemock.MockSummarizeServiceInterface) SummarizeServiceInterface {
	return &summarizeServiceAdapter{mockService: mockService}
}

func (a *summarizeServiceAdapter) SummarizeText(ctx context.Context, text string) (*SummarizeResult, error) {
	result, err := a.mockService.SummarizeText(ctx, text)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &SummarizeResult{
		Summary:    result.Summary,
		SourceText: result.SourceText,
	}, nil
}

func (a *summarizeServiceAdapter) SummarizeFile(ctx context.Context, file *multipart.FileHeader) (*SummarizeResult, error) {
	result, err := a.mockService.SummarizeFile(ctx, file)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &SummarizeResult{
		Summary:    result.Summary,
		SourceText: result.SourceText,
	}, nil
}

func TestProcessDocument(t *testing.T) {
	longText := "これはテストドキュメントの内容です。十分な長さがあるので要約も生成されます。これはテストです。テストです。テストです。繰り返しです。長い文章です。"
	longErrorText := "これはテストドキュメントの内容です。十分な長さがあるので要約も生成されますが、要約に失敗します。これはテストです。テストです。テストです。繰り返しです。長い文章です。"

	testCases := []struct {
		name              string
		fileHeader        *multipart.FileHeader
		textractResult    *aws.TextractResult
		textractErr       error
		summarizeResult   *servicemock.SummarizeResult
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
			summarizeResult: &servicemock.SummarizeResult{
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
			summarizeResult:   &servicemock.SummarizeResult{},
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
			// gomockコントローラーの作成
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// モックの作成
			mockTextract := awsmock.NewMockTextractClientInterface(ctrl)
			mockSummarize := servicemock.NewMockSummarizeServiceInterface(ctrl)

			// モックの振る舞いを設定
			mockTextract.EXPECT().
				ExtractTextFromDocument(gomock.Any(), tc.fileHeader).
				Return(tc.textractResult, tc.textractErr).
				AnyTimes()

			if tc.shouldCallSummary {
				mockSummarize.EXPECT().
					SummarizeText(gomock.Any(), longText).
					Return(tc.summarizeResult, tc.summarizeErr).
					AnyTimes()

				if tc.textractResult != nil && tc.textractResult.Text == longErrorText {
					mockSummarize.EXPECT().
						SummarizeText(gomock.Any(), longErrorText).
						Return(tc.summarizeResult, tc.summarizeErr).
						AnyTimes()
				}
			}

			// アダプターを介してテスト対象のサービスを作成
			documentService := &DocumentService{
				textractClient:   mockTextract,
				summarizeService: newSummarizeServiceAdapter(mockSummarize),
			}

			// テスト実行
			result, err := documentService.ProcessDocument(context.Background(), tc.fileHeader)

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
			}
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
		summarizeResult   *servicemock.SummarizeResult
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
			summarizeResult: &servicemock.SummarizeResult{
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
			// gomockコントローラーの作成
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// モックの作成
			mockTextract := awsmock.NewMockTextractClientInterface(ctrl)
			mockSummarize := servicemock.NewMockSummarizeServiceInterface(ctrl)

			// モックの振る舞いを設定
			mockTextract.EXPECT().
				ExtractTextFromS3Key(gomock.Any(), tc.s3Key).
				Return(tc.textractResult, tc.textractErr).
				AnyTimes()

			if tc.shouldCallSummary {
				mockSummarize.EXPECT().
					SummarizeText(gomock.Any(), gomock.Any()).
					Return(tc.summarizeResult, tc.summarizeErr).
					AnyTimes()
			}

			// アダプターを介してテスト対象のサービスを作成
			documentService := &DocumentService{
				textractClient:   mockTextract,
				summarizeService: newSummarizeServiceAdapter(mockSummarize),
			}

			// テスト実行
			result, err := documentService.ProcessDocumentByS3Key(context.Background(), tc.s3Key)

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
			}
		})
	}
}
