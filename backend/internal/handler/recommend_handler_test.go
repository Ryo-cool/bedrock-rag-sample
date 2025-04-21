package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bedrock-rag-sample/backend/internal/domain"
	domainmocks "bedrock-rag-sample/backend/internal/domain/mocks" // DBHandler モック (引数用に必要)
	"bedrock-rag-sample/backend/internal/handler"
	"bedrock-rag-sample/backend/internal/services"
	servicemocks "bedrock-rag-sample/backend/internal/services/mocks"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommendHandler_HandleRecommend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRecommendService := servicemocks.NewMockRecommendServiceInterface(ctrl)
	mockDBHandler := domainmocks.NewMockDBHandlerInterface(ctrl) // ハンドラー生成用にモックを用意
	recommendHandler := handler.NewRecommendHandler(mockRecommendService, mockDBHandler)

	e := echo.New()

	query := "類似文書を探す"
	limit := 3
	reqBody := handler.RecommendRequest{Query: query, Limit: limit}
	reqBytes, _ := json.Marshal(reqBody)

	serviceResult := &services.RecommendResult{
		Query: query,
		RecommendedChunks: []domain.DocumentChunk{
			{ID: 1, DocumentID: 10, Content: "チャンク1", Similarity: 0.9},
		},
		Documents: map[int64]*domain.Document{
			10: {ID: 10, Filename: "文書10"},
		},
	}

	t.Run("正常系", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/recommend", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// モックの設定
		mockRecommendService.EXPECT().
			FindSimilarDocuments(gomock.Any(), query, limit).
			Return(serviceResult, nil).
			Times(1)

		// ハンドラーの実行
		err := recommendHandler.HandleRecommend(c)

		// アサーション
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp services.RecommendResult
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, *serviceResult, resp)
	})

	t.Run("正常系_limit指定なし", func(t *testing.T) {
		defaultLimit := 5
		reqBodyNoLimit := handler.RecommendRequest{Query: query} // Limit を指定しない
		reqBytesNoLimit, _ := json.Marshal(reqBodyNoLimit)
		req := httptest.NewRequest(http.MethodPost, "/recommend", bytes.NewReader(reqBytesNoLimit))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// モックの設定 (limit がデフォルト値で呼ばれることを期待)
		mockRecommendService.EXPECT().
			FindSimilarDocuments(gomock.Any(), query, defaultLimit).
			Return(serviceResult, nil).
			Times(1)

		err := recommendHandler.HandleRecommend(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		// レスポンス内容は上の正常系と同じと仮定
	})

	t.Run("異常系_リクエストボディ不正", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/recommend", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := recommendHandler.HandleRecommend(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "不正なリクエストです")
	})

	t.Run("異常系_クエリ未指定", func(t *testing.T) {
		emptyReqBody := handler.RecommendRequest{Query: ""}
		emptyReqBytes, _ := json.Marshal(emptyReqBody)
		req := httptest.NewRequest(http.MethodPost, "/recommend", bytes.NewReader(emptyReqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := recommendHandler.HandleRecommend(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Contains(t, httpError.Message, "検索クエリを指定してください")
	})

	t.Run("異常系_サービスエラー", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/recommend", bytes.NewReader(reqBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		serviceError := errors.New("recommend service failed")
		mockRecommendService.EXPECT().
			FindSimilarDocuments(gomock.Any(), query, limit).
			Return(nil, serviceError).
			Times(1)

		err := recommendHandler.HandleRecommend(c)

		require.Error(t, err)
		httpError, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Contains(t, httpError.Message, "推薦処理に失敗しました")
		assert.Contains(t, httpError.Message.(string), serviceError.Error())
	})
}
