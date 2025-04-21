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
// @Description テキスト要約APIのリクエストボディ
type TextRequest struct {
	Text string `json:"text" validate:"required" example:"これは要約するテキストです。"` // 要約するテキスト
}

// SummarizeText はテキストを要約するエンドポイントを処理する
// @Summary テキスト要約
// @Description 指定されたテキストを要約します。
// @Tags Summarize
// @Accept json
// @Produce json
// @Param text body TextRequest true "要約するテキストを含むリクエストボディ"
// @Success 200 {object} services.SummarizeResult "要約成功"
// @Failure 400 {object} models.ErrorResponse "リクエスト不正 (テキストが空など)"
// @Failure 500 {object} models.ErrorResponse "サーバー内部エラー (要約失敗)"
// @Router /api/v1/summarize/text [post]
func (h *SummarizeHandler) SummarizeText(c echo.Context) error {
	var req TextRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "リクエストボディの解析に失敗しました").SetInternal(err)
	}

	if req.Text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "テキストが空です")
	}

	result, err := h.summarizeService.SummarizeText(c.Request().Context(), req.Text)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "要約の生成に失敗しました").SetInternal(err)
	}

	return c.JSON(http.StatusOK, result)
}

// SummarizeFile はファイルを要約するエンドポイントを処理する
// @Summary ファイル要約
// @Description アップロードされたファイル (PDF, 画像) の内容を要約します。
// @Tags Summarize
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要約するファイル (PDF/画像, 最大5MB)"
// @Success 200 {object} services.SummarizeResult "要約成功"
// @Failure 400 {object} models.ErrorResponse "リクエスト不正 (ファイルなし、サイズ超過など)"
// @Failure 500 {object} models.ErrorResponse "サーバー内部エラー (要約失敗)"
// @Router /api/v1/summarize/file [post]
func (h *SummarizeHandler) SummarizeFile(c echo.Context) error {
	// multipart/form-dataからファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ファイルが見つかりませんでした")
	}

	// ファイルサイズの検証（例: 最大5MB）
	if file.Size > 5*1024*1024 {
		return echo.NewHTTPError(http.StatusBadRequest, "ファイルサイズが大きすぎます（最大5MB）")
	}

	result, err := h.summarizeService.SummarizeFile(c.Request().Context(), file)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ファイルの要約に失敗しました").SetInternal(err)
	}

	return c.JSON(http.StatusOK, result)
}
