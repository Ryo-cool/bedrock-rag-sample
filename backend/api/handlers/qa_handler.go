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
type QueryRequest struct {
	Query string `json:"query"`
}

// AskQuestion は質問をRAGシステムに送信し回答を生成するエンドポイントを処理する
func (h *QAHandler) AskQuestion(c echo.Context) error {
	var req QueryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストボディの解析に失敗しました",
		})
	}

	if req.Query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "質問が空です",
		})
	}

	// RAG処理を実行
	result, err := h.qaService.SimpleRAG(c.Request().Context(), req.Query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "質問への回答生成に失敗しました: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
