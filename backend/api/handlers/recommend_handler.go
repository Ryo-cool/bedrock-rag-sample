package handlers

import (
	"net/http"
	"strconv"

	internalModels "bedrock-rag-sample/backend/internal/models"
	"bedrock-rag-sample/backend/internal/services"

	"github.com/labstack/echo/v4"
)

// RecommendHandler はレコメンド関連のエンドポイントを処理するハンドラー
type RecommendHandler struct {
	recommendService *services.RecommendService
	dbHandler        *internalModels.DBHandler
}

// NewRecommendHandler は新しいRecommendHandlerを作成する
func NewRecommendHandler(recommendService *services.RecommendService, dbHandler *internalModels.DBHandler) *RecommendHandler {
	return &RecommendHandler{
		recommendService: recommendService,
		dbHandler:        dbHandler,
	}
}

// RecommendRequest はレコメンドリクエストの構造体
// @Description 類似ドキュメント検索APIのリクエストボディ
type RecommendRequest struct {
	Query string `json:"query" validate:"required" example:"Bedrockについて"` // 検索クエリ
	Limit int    `json:"limit,omitempty" example:"5"`                     // 取得するドキュメントの最大数 (デフォルト5, 最大20)
}

// FindSimilarDocuments はクエリに類似したドキュメントを検索するエンドポイントを処理する
// @Summary 類似ドキュメント検索
// @Description 指定されたクエリにベクトル検索で類似するドキュメントを検索します。
// @Tags Recommend
// @Accept json
// @Produce json
// @Param query body RecommendRequest true "検索クエリと取得件数"
// @Success 200 {array} internalModels.DocumentChunk "類似ドキュメントチャンクの配列"
// @Failure 400 {object} models.ErrorResponse "リクエスト不正 (クエリが空など)"
// @Failure 500 {object} models.ErrorResponse "サーバー内部エラー (検索失敗)"
// @Router /api/v1/recommend [post]
func (h *RecommendHandler) FindSimilarDocuments(c echo.Context) error {
	var req RecommendRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディの解析に失敗しました").SetInternal(err)
	}

	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "検索クエリが空です")
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
		return echo.NewHTTPError(http.StatusInternalServerError, "類似ドキュメントの検索に失敗しました").SetInternal(err)
	}

	return c.JSON(http.StatusOK, result)
}

// ProcessDocumentForEmbedding はドキュメントのEmbeddingを処理するエンドポイントを処理する
// @Summary ドキュメントEmbedding処理
// @Description 指定されたIDのドキュメントを取得し、Embeddingを生成・保存します。
// @Tags Recommend
// @Accept json
// @Produce json
// @Param id path int true "処理対象ドキュメントのID"
// @Success 200 {object} map[string]string "処理成功メッセージ"
// @Failure 400 {object} models.ErrorResponse "リクエスト不正 (無効なIDなど)"
// @Failure 404 {object} models.ErrorResponse "ドキュメントが見つからない"
// @Failure 500 {object} models.ErrorResponse "サーバー内部エラー (処理失敗)"
// @Router /api/v1/document/{id}/embedding [post]
func (h *RecommendHandler) ProcessDocumentForEmbedding(c echo.Context) error {
	// URLからドキュメントIDを取得
	docIDStr := c.Param("id")
	if docIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ドキュメントIDが指定されていません")
	}

	// ドキュメントIDを数値に変換
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なドキュメントIDです").SetInternal(err)
	}

	// DBからドキュメントを取得
	doc, err := h.dbHandler.GetDocumentByID(c.Request().Context(), docID)
	if err != nil {
		// エラーの種類によってStatus Codeを変える
		// TODO: DB層で NotFound 専用のエラーを返すようにするとより良い
		return echo.NewHTTPError(http.StatusNotFound, "ドキュメントが見つかりません").SetInternal(err)
	}

	// Embedding処理を実行
	if err := h.recommendService.ProcessDocumentForEmbedding(c.Request().Context(), doc); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ドキュメントのEmbedding処理に失敗しました").SetInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ドキュメントのEmbedding処理が完了しました",
	})
}
