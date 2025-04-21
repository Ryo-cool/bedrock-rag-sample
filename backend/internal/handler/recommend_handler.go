package handler

import (
	"bedrock-rag-sample/backend/internal/domain" // domain をインポート
	"bedrock-rag-sample/backend/internal/services"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// RecommendHandler は類似文書推薦に関するハンドラー
type RecommendHandler struct {
	recommendService *services.RecommendService
	dbHandler        *domain.DBHandler // domain.DBHandler を参照
}

// NewRecommendHandler は新しいRecommendHandlerを生成する
func NewRecommendHandler(recommendService *services.RecommendService, dbHandler *domain.DBHandler) *RecommendHandler { // domain.DBHandler を引数に取る
	return &RecommendHandler{
		recommendService: recommendService,
		dbHandler:        dbHandler,
	}
}

// RecommendRequest は推薦リクエストの構造体
type RecommendRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// HandleRecommend は類似文書の推薦リクエストを処理する
func (h *RecommendHandler) HandleRecommend(c echo.Context) error {
	var req RecommendRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "不正なリクエストです")
	}

	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "検索クエリを指定してください")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 5 // デフォルト値
	}

	result, err := h.recommendService.FindSimilarDocuments(c.Request().Context(), req.Query, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("推薦処理に失敗しました: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}
