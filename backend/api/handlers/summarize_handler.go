package handlers

import (
	"net/http"

	"bedrock-rag-sample/backend/internal/services"

	"github.com/labstack/echo/v4"
)

// SummarizeHandler は要約関連のエンドポイントを処理するハンドラー
type SummarizeHandler struct {
	summarizeService *services.SummarizeService
}

// NewSummarizeHandler は新しいSummarizeHandlerを作成する
func NewSummarizeHandler(summarizeService *services.SummarizeService) *SummarizeHandler {
	return &SummarizeHandler{
		summarizeService: summarizeService,
	}
}

// TextRequest はテキスト要約リクエストの構造体
type TextRequest struct {
	Text string `json:"text"`
}

// SummarizeText はテキストを要約するエンドポイントを処理する
func (h *SummarizeHandler) SummarizeText(c echo.Context) error {
	var req TextRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストボディの解析に失敗しました",
		})
	}

	if req.Text == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "テキストが空です",
		})
	}

	result, err := h.summarizeService.SummarizeText(c.Request().Context(), req.Text)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "要約の生成に失敗しました: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

// SummarizeFile はファイルを要約するエンドポイントを処理する
func (h *SummarizeHandler) SummarizeFile(c echo.Context) error {
	// multipart/form-dataからファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ファイルが見つかりませんでした",
		})
	}

	// ファイルサイズの検証（例: 最大5MB）
	if file.Size > 5*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ファイルサイズが大きすぎます（最大5MB）",
		})
	}

	result, err := h.summarizeService.SummarizeFile(c.Request().Context(), file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "ファイルの要約に失敗しました: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
