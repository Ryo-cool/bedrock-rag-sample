package handlers

import (
	"net/http"

	"bedrock-rag-sample/backend/internal/services"

	"github.com/labstack/echo/v4"
)

// QAHandler はQA関連のエンドポイントを処理するハンドラー
type QAHandler struct {
	qaService *services.QAService
}

// NewQAHandler は新しいQAHandlerを作成する
func NewQAHandler(qaService *services.QAService) *QAHandler {
	return &QAHandler{
		qaService: qaService,
	}
}

// QueryRequest はQ&Aリクエストの構造体
// @Description Q&A APIのリクエストボディ
type QueryRequest struct {
	Query string `json:"query" validate:"required" example:"BedrockのKnowledge Baseについて教えてください。"` // RAGシステムへの質問
}

// AskQuestion は質問をRAGシステムに送信し回答を生成するエンドポイントを処理する
// @Summary RAG Q&A
// @Description Knowledge Baseを使用して質問に回答します。
// @Tags QA
// @Accept json
// @Produce json
// @Param query body QueryRequest true "質問を含むリクエストボディ"
// @Success 200 {object} services.QAResult "回答生成成功"
// @Failure 400 {object} models.ErrorResponse "リクエスト不正 (質問が空など)"
// @Failure 500 {object} models.ErrorResponse "サーバー内部エラー (回答生成失敗)"
// @Router /api/v1/qa [post]
func (h *QAHandler) AskQuestion(c echo.Context) error {
	var req QueryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディの解析に失敗しました").SetInternal(err)
	}

	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "質問が空です")
	}

	// RAG処理を実行
	result, err := h.qaService.SimpleRAG(c.Request().Context(), req.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "質問への回答生成に失敗しました").SetInternal(err)
	}

	return c.JSON(http.StatusOK, result)
}
