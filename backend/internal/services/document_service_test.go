package services_test

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"testing"

	// models は ProcessDocument/ProcessDocumentByS3Key の戻り値型 DocumentProcessResult で使われるため必要

	// repositorymocks と repository は不要
	// repositorymocks "bedrock-rag-sample/backend/internal/repository/mock"
	// "bedrock-rag-sample/backend/internal/repository"
	"bedrock-rag-sample/backend/internal/services"
	servicemocks "bedrock-rag-sample/backend/internal/services/mocks"
	"bedrock-rag-sample/backend/pkg/aws"
	awsmock "bedrock-rag-sample/backend/pkg/aws/mock"

	"github.com/golang/mock/gomock"
	// uuid は不要になった
	// "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// summarizeServiceAdapter は SummarizeServiceInterface のモックアダプター
// DocumentService のテストで SummarizeService のモックを使うために必要
type summarizeServiceAdapter struct {
	mockService *servicemocks.MockSummarizeServiceInterface // モックサービスを保持 (正しい型を参照)
}

// newSummarizeServiceAdapter は新しいアダプターを作成する
func newSummarizeServiceAdapter(mockService *servicemocks.MockSummarizeServiceInterface) services.SummarizeServiceInterface {
	return &summarizeServiceAdapter{mockService: mockService}
}

// SummarizeText はモックサービスに委譲する (型変換は不要)
func (a *summarizeServiceAdapter) SummarizeText(ctx context.Context, text string) (*services.SummarizeResult, error) {
	return a.mockService.SummarizeText(ctx, text) // モックが *services.SummarizeResult を返す想定
}

// SummarizeFile はインターフェースを満たすためのダミー実装 (新しいシグネチャ)
// このテストファイルでは呼び出されない想定
func (a *summarizeServiceAdapter) SummarizeFile(ctx context.Context, fileContent io.Reader, fileName string) (*services.SummarizeResult, error) {
	return nil, errors.New("SummarizeFile not expected to be called via adapter")
}

// SummarizeFileByS3Key はモックサービスに委譲する (型変換は不要)
func (a *summarizeServiceAdapter) SummarizeFileByS3Key(ctx context.Context, s3Key string) (*services.SummarizeResult, error) {
	return a.mockService.SummarizeFileByS3Key(ctx, s3Key) // モックが *services.SummarizeResult を返す想定
}

func TestProcessDocument(t *testing.T) {
	longText := "これはテストドキュメントの内容です。十分な長さがあるので要約も生成されます。これはテストです。テストです。テストです。繰り返しです。長い文章です。"
	longErrorText := "これはテストドキュメントの内容です。十分な長さがあるので要約も生成されますが、要約に失敗します。これはテストです。テストです。テストです。繰り返しです。長い文章です。"

	testCases := []struct {
		name              string
		fileHeader        *multipart.FileHeader
		textractResult    *aws.TextractResult
		textractErr       error
		summarizeResult   *services.SummarizeResult
		summarizeErr      error
		expectedResult    *services.DocumentProcessResult
		expectError       bool
		shouldCallSummary bool
	}{
		{
			name:       "PDFファイルの処理成功（要約あり）",
			fileHeader: createTestFileHeader("test.pdf", longText),
			textractResult: &aws.TextractResult{
				Text:       longText,
				DocumentID: "test.pdf",
				S3Key:      "documents/test.pdf",
				Pages:      1,
			},
			textractErr: nil,
			summarizeResult: &services.SummarizeResult{
				Summary:    "テストドキュメントの要約",
				SourceText: "元のテキスト",
			},
			summarizeErr:      nil,
			shouldCallSummary: true,
			expectedResult: &services.DocumentProcessResult{
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
			name:       "PDFファイルの処理成功（テキストが短いので要約なし）",
			fileHeader: createTestFileHeader("test.pdf", "短いテキスト"),
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
			expectedResult: &services.DocumentProcessResult{
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
			name:              "サポートされていないファイル形式",
			fileHeader:        createTestFileHeader("test.docx", "dummy"),
			textractResult:    nil,
			textractErr:       nil,
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
		{
			name:              "テキスト抽出エラー",
			fileHeader:        createTestFileHeader("test.pdf", "dummy"),
			textractResult:    nil,
			textractErr:       errors.New("テキスト抽出エラー"),
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
		{
			name:       "テキスト抽出は成功したが空文字",
			fileHeader: createTestFileHeader("test.pdf", ""),
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
			name:       "要約エラーでも処理は成功する",
			fileHeader: createTestFileHeader("test.pdf", longErrorText),
			textractResult: &aws.TextractResult{
				Text:       longErrorText,
				DocumentID: "test.pdf",
				S3Key:      "documents/test.pdf",
				Pages:      1,
			},
			textractErr:       nil,
			summarizeResult:   &services.SummarizeResult{},
			summarizeErr:      errors.New("要約エラー"),
			shouldCallSummary: true,
			expectedResult: &services.DocumentProcessResult{
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTextract := awsmock.NewMockTextractClientInterface(ctrl)
			mockSummarize := servicemocks.NewMockSummarizeServiceInterface(ctrl)
			summarizeAdapter := newSummarizeServiceAdapter(mockSummarize)

			documentService := services.NewDocumentService(mockTextract, summarizeAdapter)

			ctx := context.Background()

			if tc.fileHeader != nil {
				mockTextract.EXPECT().
					ExtractTextFromDocument(ctx, tc.fileHeader).
					Return(tc.textractResult, tc.textractErr).
					MaxTimes(1)
			}

			if tc.shouldCallSummary && tc.textractResult != nil {
				mockSummarize.EXPECT().
					SummarizeText(ctx, tc.textractResult.Text).
					Return(tc.summarizeResult, tc.summarizeErr).
					MaxTimes(1)
			}

			result, err := documentService.ProcessDocument(ctx, tc.fileHeader)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.textractErr != nil {
					assert.ErrorIs(t, err, tc.textractErr)
				} else if tc.fileHeader != nil && filepath.Ext(tc.fileHeader.Filename) == ".docx" {
					assert.Contains(t, err.Error(), "サポートされていないファイル形式です")
				} else if tc.textractResult != nil && tc.textractResult.Text == "" {
					assert.Contains(t, err.Error(), "テキストを抽出できませんでした")
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
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
		summarizeResult   *services.SummarizeResult
		summarizeErr      error
		expectedResult    *services.DocumentProcessResult
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
			summarizeResult: &services.SummarizeResult{
				Summary:    "S3ドキュメントの要約",
				SourceText: "元のテキスト",
			},
			summarizeErr:      nil,
			shouldCallSummary: true,
			expectedResult: &services.DocumentProcessResult{
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
			textractErr:       errors.New("S3テキスト抽出エラー"),
			summarizeResult:   nil,
			summarizeErr:      nil,
			shouldCallSummary: false,
			expectedResult:    nil,
			expectError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTextract := awsmock.NewMockTextractClientInterface(ctrl)
			mockSummarize := servicemocks.NewMockSummarizeServiceInterface(ctrl)
			summarizeAdapter := newSummarizeServiceAdapter(mockSummarize)

			documentService := services.NewDocumentService(mockTextract, summarizeAdapter)

			ctx := context.Background()

			mockTextract.EXPECT().
				ExtractTextFromS3Key(ctx, tc.s3Key).
				Return(tc.textractResult, tc.textractErr).
				MaxTimes(1)

			if tc.shouldCallSummary && tc.textractResult != nil {
				mockSummarize.EXPECT().
					SummarizeText(ctx, tc.textractResult.Text).
					Return(tc.summarizeResult, tc.summarizeErr).
					MaxTimes(1)
			}

			result, err := documentService.ProcessDocumentByS3Key(ctx, tc.s3Key)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tc.textractErr != nil {
					assert.ErrorIs(t, err, tc.textractErr)
				} else if filepath.Ext(tc.s3Key) == ".docx" {
					assert.Contains(t, err.Error(), "サポートされていないファイル形式です")
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.expectedResult.OriginalText, result.OriginalText)
				assert.Equal(t, tc.expectedResult.Summary, result.Summary)
				assert.Equal(t, tc.expectedResult.FileType, result.FileType)
			}
		})
	}
}

// createTestFileHeader はテスト用の multipart.FileHeader を作成するヘルパー関数
// (document_service.go で必要になる可能性があるため残しておく)
func createTestFileHeader(filename string, content string) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: filename,
		Header:   textproto.MIMEHeader{},
		Size:     int64(len(content)),
	}
}
