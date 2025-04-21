package services

import (
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"
	"testing"

	"bedrock-rag-sample/backend/pkg/aws/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUploadFile(t *testing.T) {
	testCases := []struct {
		name           string
		fileHeader     *multipart.FileHeader
		s3UploadKey    string
		s3UploadErr    error
		s3GetURLResult string
		s3GetURLErr    error
		expectedResult *UploadFileResult
		expectError    bool
	}{
		{
			name: "PDFファイルのアップロード成功",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			s3UploadKey:    "documents/pdf/test.pdf",
			s3UploadErr:    nil,
			s3GetURLResult: "https://example.com/test.pdf",
			s3GetURLErr:    nil,
			expectedResult: &UploadFileResult{
				Key:      "documents/pdf/test.pdf",
				Filename: "test.pdf",
				URL:      "https://example.com/test.pdf",
			},
			expectError: false,
		},
		{
			name: "画像ファイルのアップロード成功",
			fileHeader: &multipart.FileHeader{
				Filename: "test.jpg",
				Size:     1024,
			},
			s3UploadKey:    "documents/images/test.jpg",
			s3UploadErr:    nil,
			s3GetURLResult: "https://example.com/test.jpg",
			s3GetURLErr:    nil,
			expectedResult: &UploadFileResult{
				Key:      "documents/images/test.jpg",
				Filename: "test.jpg",
				URL:      "https://example.com/test.jpg",
			},
			expectError: false,
		},
		{
			name: "その他のファイルのアップロード成功",
			fileHeader: &multipart.FileHeader{
				Filename: "test.txt",
				Size:     1024,
			},
			s3UploadKey:    "documents/others/test.txt",
			s3UploadErr:    nil,
			s3GetURLResult: "https://example.com/test.txt",
			s3GetURLErr:    nil,
			expectedResult: &UploadFileResult{
				Key:      "documents/others/test.txt",
				Filename: "test.txt",
				URL:      "https://example.com/test.txt",
			},
			expectError: false,
		},
		{
			name: "S3アップロードエラー",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			s3UploadKey:    "",
			s3UploadErr:    errors.New("アップロードエラー"),
			s3GetURLResult: "",
			s3GetURLErr:    nil,
			expectedResult: nil,
			expectError:    true,
		},
		{
			name: "URL取得エラーでもアップロード自体は成功",
			fileHeader: &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     1024,
			},
			s3UploadKey:    "documents/pdf/test.pdf",
			s3UploadErr:    nil,
			s3GetURLResult: "",
			s3GetURLErr:    errors.New("URL取得エラー"),
			expectedResult: &UploadFileResult{
				Key:      "documents/pdf/test.pdf",
				Filename: "test.pdf",
				URL:      "",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// gomockコントローラーの作成
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// モックS3クライアントの作成
			mockS3 := mock.NewMockS3ClientInterface(ctrl)

			// テスト対象のサービスを作成
			uploadService := &UploadService{
				s3Client: mockS3,
			}

			// S3クライアントのメソッド呼び出しを設定
			var folderPath string
			switch filepath.Ext(tc.fileHeader.Filename) {
			case ".pdf":
				folderPath = "pdf"
			case ".png", ".jpg", ".jpeg":
				folderPath = "images"
			default:
				folderPath = "others"
			}

			// モックの期待値を設定
			mockS3.EXPECT().
				UploadFile(gomock.Any(), tc.fileHeader, folderPath).
				Return(tc.s3UploadKey, tc.s3UploadErr)

			if tc.s3UploadErr == nil {
				mockS3.EXPECT().
					GetFileURL(gomock.Any(), tc.s3UploadKey).
					Return(tc.s3GetURLResult, tc.s3GetURLErr)
			}

			// テスト実行
			result, err := uploadService.UploadFile(context.Background(), tc.fileHeader)

			// 検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedResult.Key, result.Key)
				assert.Equal(t, tc.expectedResult.Filename, result.Filename)
				assert.Equal(t, tc.expectedResult.URL, result.URL)
			}
		})
	}
}
