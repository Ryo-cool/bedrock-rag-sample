package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bedrock-rag-sample/backend/internal/handler"
	"bedrock-rag-sample/backend/internal/services"
	servicemocks "bedrock-rag-sample/backend/internal/services/mocks"
	"bedrock-rag-sample/backend/pkg/aws" // aws.TextractResult のために必要

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentHandler_HandleProcessDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDocumentService := servicemocks.NewMockDocumentServiceInterface(ctrl)
	documentHandler := handler.NewDocumentHandler(mockDocumentService)

	e := echo.New()

	s3Key := "path/to/document.pdf"
	reqBody := handler.ProcessDocumentRequest{S3Key: s3Key}
	reqBytes, _ := json.Marshal(reqBody)

	serviceResult := &services.DocumentProcessResult{
		OriginalText: "抽出されたテキスト",
		Summary:      "要約結果",
		DocumentInfo: aws.TextractResult{Text: "抽出されたテキスト" /* 他の TextractResult フィールドも必要に応じて設定 */},
		FileType:     "pdf",
	}

	t.Run("正常系", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/documents/process", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// モックの設定
		mockDocumentService.EXPECT().
			ProcessDocumentByS3Key(gomock.Any(), s3Key).
			Return(serviceResult, nil).
			Times(1)

		// ハンドラーの実行
		err := documentHandler.HandleProcessDocument(c)

		// アサーション
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp services.DocumentProcessResult
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, *serviceResult, resp) // 結果が一致するか確認
	})

	t.Run("異常系_リクエストボディ不正", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/documents/process", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := documentHandler.HandleProcessDocument(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "不正なリクエストです")
	})

	t.Run("異常系_S3キー未指定", func(t *testing.T) {
		emptyReqBody := handler.ProcessDocumentRequest{S3Key: ""}
		emptyReqBytes, _ := json.Marshal(emptyReqBody)
		req := httptest.NewRequest(http.MethodPost, "/documents/process", bytes.NewReader(emptyReqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := documentHandler.HandleProcessDocument(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "S3キーを指定してください")
	})

	t.Run("異常系_サービスエラー", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/documents/process", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		serviceError := errors.New("service processing failed")
		mockDocumentService.EXPECT().
			ProcessDocumentByS3Key(gomock.Any(), s3Key).
			Return(nil, serviceError).
			Times(1)

		err := documentHandler.HandleProcessDocument(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Contains(t, httpError.Message, "ドキュメント処理に失敗しました")
		// fmt.Sprintf されたエラーメッセージを確認
		assert.Contains(t, httpError.Message.(string), serviceError.Error()) // 元のエラーが含まれているか
	})
}
