package handler_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bedrock-rag-sample/backend/internal/handler"
	"bedrock-rag-sample/backend/internal/services"
	servicemocks "bedrock-rag-sample/backend/internal/services/mocks"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadHandler_HandleUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUploadService := servicemocks.NewMockUploadServiceInterface(ctrl)
	uploadHandler := handler.NewUploadHandler(mockUploadService)

	e := echo.New()

	// --- ヘルパー: multipart/form-data リクエストを作成 ---
	createMultipartRequest := func(fieldName string, fileName string, fileContent string) (*http.Request, error) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile(fieldName, fileName)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, strings.NewReader(fileContent))
		if err != nil {
			return nil, err
		}
		err = writer.Close()
		if err != nil {
			return nil, err
		}

		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		return req, nil
	}
	// --- ここまでヘルパー ---

	fieldName := "file"
	fileName := "upload.txt"
	fileContent := "アップロードするファイルの内容"

	serviceResult := &services.UploadFileResult{
		Key:      "uploaded/upload.txt",
		Filename: fileName,
		URL:      "http://example.com/uploaded/upload.txt",
	}

	t.Run("正常系", func(t *testing.T) {
		req, err := createMultipartRequest(fieldName, fileName, fileContent)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// モックの設定: UploadFile が呼ばれることを期待
		// fileHeader の比較は難しいので gomock.Any() を使う
		mockUploadService.EXPECT().
			UploadFile(gomock.Any(), gomock.Any()). // fileHeader の比較は困難なため Any
			DoAndReturn(func(ctx context.Context, header *multipart.FileHeader) (*services.UploadFileResult, error) {
				// 呼び出された際のヘッダーのファイル名が期待通りかくらいはチェックできる
				assert.Equal(t, fileName, header.Filename)
				return serviceResult, nil
			}).Times(1)

		// ハンドラーの実行
		err = uploadHandler.HandleUpload(c)

		// アサーション
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		// レスポンスボディの確認 (map[string]string)
		assert.Contains(t, rec.Body.String(), "\"message\":\"ファイルが正常にアップロードされました\"")
		assert.Contains(t, rec.Body.String(), "\"location\":\""+serviceResult.URL+"\"")
	})

	t.Run("異常系_ファイルなし", func(t *testing.T) {
		// ファイルを含まないリクエストを作成
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.Close() // ファイルを追加せずに閉じる
		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := uploadHandler.HandleUpload(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "ファイルの取得に失敗しました")
	})

	t.Run("異常系_サービスエラー", func(t *testing.T) {
		req, err := createMultipartRequest(fieldName, fileName, fileContent)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		serviceError := errors.New("upload service failed")
		mockUploadService.EXPECT().
			UploadFile(gomock.Any(), gomock.Any()).
			Return(nil, serviceError).
			Times(1)

		err = uploadHandler.HandleUpload(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Contains(t, httpError.Message, "ファイルのアップロードに失敗しました")
		assert.Contains(t, httpError.Message.(string), serviceError.Error())
	})
}
