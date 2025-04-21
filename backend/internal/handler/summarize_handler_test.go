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

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarizeHandler_HandleTextSummarize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSummarizeService := servicemocks.NewMockSummarizeServiceInterface(ctrl)
	summarizeHandler := handler.NewSummarizeHandler(mockSummarizeService)

	e := echo.New()

	inputText := "これは要約されるテキストです。"
	reqBody := handler.TextSummarizeRequest{Text: inputText}
	reqBytes, _ := json.Marshal(reqBody)

	serviceResult := &services.SummarizeResult{
		Summary: "要約結果",
	}

	t.Run("正常系", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/summarize/text", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// モックの設定
		mockSummarizeService.EXPECT().
			SummarizeText(gomock.Any(), inputText).
			Return(serviceResult, nil).
			Times(1)

		// ハンドラーの実行
		err := summarizeHandler.HandleTextSummarize(c)

		// アサーション
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, serviceResult.Summary, resp["summary"])
	})

	t.Run("異常系_リクエストボディ不正", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/summarize/text", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := summarizeHandler.HandleTextSummarize(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "不正なリクエストです")
	})

	t.Run("異常系_テキスト未指定", func(t *testing.T) {
		emptyReqBody := handler.TextSummarizeRequest{Text: ""}
		emptyReqBytes, _ := json.Marshal(emptyReqBody)
		req := httptest.NewRequest(http.MethodPost, "/summarize/text", bytes.NewReader(emptyReqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := summarizeHandler.HandleTextSummarize(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "要約するテキストを指定してください")
	})

	t.Run("異常系_サービスエラー", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/summarize/text", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		serviceError := errors.New("summarize service failed")
		mockSummarizeService.EXPECT().
			SummarizeText(gomock.Any(), inputText).
			Return(nil, serviceError).
			Times(1)

		err := summarizeHandler.HandleTextSummarize(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Contains(t, httpError.Message, "要約処理に失敗しました")
		assert.Contains(t, httpError.Message.(string), serviceError.Error())
	})
}

func TestSummarizeHandler_HandleFileSummarize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSummarizeService := servicemocks.NewMockSummarizeServiceInterface(ctrl)
	summarizeHandler := handler.NewSummarizeHandler(mockSummarizeService)

	e := echo.New()

	s3Key := "path/to/summarize.txt"
	reqBody := handler.FileSummarizeRequest{S3Key: s3Key}
	reqBytes, _ := json.Marshal(reqBody)

	serviceResult := &services.SummarizeResult{
		Summary: "ファイル要約結果",
	}

	t.Run("正常系", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/summarize/file", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// モックの設定
		mockSummarizeService.EXPECT().
			SummarizeFileByS3Key(gomock.Any(), s3Key).
			Return(serviceResult, nil).
			Times(1)

		// ハンドラーの実行
		err := summarizeHandler.HandleFileSummarize(c)

		// アサーション
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, serviceResult.Summary, resp["summary"])
	})

	t.Run("異常系_リクエストボディ不正", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/summarize/file", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := summarizeHandler.HandleFileSummarize(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "不正なリクエストです")
	})

	t.Run("異常系_S3キー未指定", func(t *testing.T) {
		emptyReqBody := handler.FileSummarizeRequest{S3Key: ""}
		emptyReqBytes, _ := json.Marshal(emptyReqBody)
		req := httptest.NewRequest(http.MethodPost, "/summarize/file", bytes.NewReader(emptyReqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := summarizeHandler.HandleFileSummarize(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "要約するファイルのS3キーを指定してください") // メッセージを確認
	})

	t.Run("異常系_サービスエラー", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/summarize/file", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		serviceError := errors.New("summarize file service failed")
		mockSummarizeService.EXPECT().
			SummarizeFileByS3Key(gomock.Any(), s3Key).
			Return(nil, serviceError).
			Times(1)

		err := summarizeHandler.HandleFileSummarize(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Contains(t, httpError.Message, "ファイル要約処理に失敗しました")
		assert.Contains(t, httpError.Message.(string), serviceError.Error())
	})
}
