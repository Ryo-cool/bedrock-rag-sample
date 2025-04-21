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

func TestQAHandler_HandleQA(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQAService := servicemocks.NewMockQAServiceInterface(ctrl)
	qaHandler := handler.NewQAHandler(mockQAService)

	e := echo.New()

	query := "これはテストの質問です？"
	reqBody := handler.QARequest{Query: query}
	reqBytes, _ := json.Marshal(reqBody)

	serviceResult := &services.QAResult{
		Query:  query,
		Answer: "これはテストの回答です。",
		RetrievedDocuments: []services.RetrievedDocument{
			{Content: "関連ドキュメント1", DocumentID: "doc1"},
		},
	}

	t.Run("正常系", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/qa", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// モックの設定
		mockQAService.EXPECT().
			SimpleRAG(gomock.Any(), query).
			Return(serviceResult, nil).
			Times(1)

		// ハンドラーの実行
		err := qaHandler.HandleQA(c)

		// アサーション
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp services.QAResult
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, *serviceResult, resp)
	})

	t.Run("異常系_リクエストボディ不正", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/qa", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := qaHandler.HandleQA(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "不正なリクエストです")
	})

	t.Run("異常系_クエリ未入力", func(t *testing.T) {
		emptyReqBody := handler.QARequest{Query: ""}
		emptyReqBytes, _ := json.Marshal(emptyReqBody)
		req := httptest.NewRequest(http.MethodPost, "/qa", bytes.NewReader(emptyReqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := qaHandler.HandleQA(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "質問を入力してください")
	})

	t.Run("異常系_サービスエラー", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/qa", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		serviceError := errors.New("qa service failed")
		mockQAService.EXPECT().
			SimpleRAG(gomock.Any(), query).
			Return(nil, serviceError).
			Times(1)

		err := qaHandler.HandleQA(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Contains(t, httpError.Message, "QA処理に失敗しました")
		assert.Contains(t, httpError.Message.(string), serviceError.Error())
	})
}
