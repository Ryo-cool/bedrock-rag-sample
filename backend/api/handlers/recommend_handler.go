package handlers

import (
	"net/http"
	"strconv"

	"bedrock-rag-sample/backend/internal/models"
	"bedrock-rag-sample/backend/internal/services"

	"github.com/labstack/echo/v4"
)

// RecommendHandler はレコメンド関連のエンドポイントを処理するハンドラー
type RecommendHandler struct {
	recommendService *services.RecommendService
	dbHandler        *models.DBHandler
}

// NewRecommendHandler は新しいRecommendHandlerを作成する
func NewRecommendHandler(recommendService *services.RecommendService, dbHandler *models.DBHandler) *RecommendHandler {
	return &RecommendHandler{
		recommendService: recommendService,
		dbHandler:        dbHandler,
	}
}

// RecommendRequest はレコメンドリクエストの構造体
type RecommendRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// FindSimilarDocuments はクエリに類似したドキュメントを検索するエンドポイントを処理する
func (h *RecommendHandler) FindSimilarDocuments(c echo.Context) error {
	var req RecommendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストボディの解析に失敗しました",
		})
	}

	if req.Query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "検索クエリが空です",
		})
	}

	// デフォルト値または最大値の設定
	if req.Limit <= 0 {
		req.Limit = 5
	} else if req.Limit > 20 {
		req.Limit = 20
	}

	// 類似ドキュメントを検索
	result, err := h.recommendService.FindSimilarDocuments(c.Request().Context(), req.Query, req.Limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "類似ドキュメントの検索に失敗しました: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

// ProcessDocumentForEmbedding はドキュメントのEmbeddingを処理するエンドポイントを処理する
func (h *RecommendHandler) ProcessDocumentForEmbedding(c echo.Context) error {
	// URLからドキュメントIDを取得
	docIDStr := c.Param("id")
	if docIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ドキュメントIDが指定されていません",
		})
	}

	// ドキュメントIDを数値に変換
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "無効なドキュメントIDです",
		})
	}

	// DBからドキュメントを取得
	doc, err := h.dbHandler.GetDocumentByID(c.Request().Context(), docID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "ドキュメントが見つかりません: " + err.Error(),
		})
	}

	// Embedding処理を実行
	if err := h.recommendService.ProcessDocumentForEmbedding(c.Request().Context(), doc); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "ドキュメントのEmbedding処理に失敗しました: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ドキュメントのEmbedding処理が完了しました",
	})
}
