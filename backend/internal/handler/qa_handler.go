package handler

import (
	"bedrock-rag-sample/backend/internal/services"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// QAHandler はQAに関するハンドラー
type QAHandler struct {
	qaService *services.QAService
}

// NewQAHandler は新しいQAHandlerを生成する
func NewQAHandler(qaService *services.QAService) *QAHandler {
	return &QAHandler{
		qaService: qaService,
	}
}

// QARequest はQAリクエストの構造体
type QARequest struct {
	Query string `json:"query"`
}

// HandleQA はQAリクエストを処理する
func (h *QAHandler) HandleQA(c echo.Context) error {
	var req QARequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "不正なリクエストです")
	}

	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "質問を入力してください")
	}

	result, err := h.qaService.SimpleRAG(c.Request().Context(), req.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("QA処理に失敗しました: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}
