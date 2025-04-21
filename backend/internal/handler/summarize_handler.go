package handler

import (
	"bedrock-rag-sample/backend/internal/services"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// SummarizeHandler はテキスト要約に関するハンドラー
type SummarizeHandler struct {
	summarizeService services.SummarizeServiceInterface
}

// NewSummarizeHandler は新しいSummarizeHandlerを生成する
func NewSummarizeHandler(summarizeService services.SummarizeServiceInterface) *SummarizeHandler {
	return &SummarizeHandler{
		summarizeService: summarizeService,
	}
}

// TextSummarizeRequest はテキスト要約リクエストの構造体
type TextSummarizeRequest struct {
	Text string `json:"text"`
}

// HandleTextSummarize は自由テキストの要約リクエストを処理する
func (h *SummarizeHandler) HandleTextSummarize(c echo.Context) error {
	var req TextSummarizeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "不正なリクエストです")
	}

	if req.Text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "要約するテキストを指定してください")
	}

	result, err := h.summarizeService.SummarizeText(c.Request().Context(), req.Text)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("要約処理に失敗しました: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"summary": result.Summary,
	})
}

// FileSummarizeRequest はファイル要約リクエストの構造体
type FileSummarizeRequest struct {
	S3Key string `json:"s3_key"`
}

// HandleFileSummarize は指定されたS3ファイルの要約リクエストを処理する
func (h *SummarizeHandler) HandleFileSummarize(c echo.Context) error {
	var req FileSummarizeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "不正なリクエストです")
	}

	if req.S3Key == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "要約するファイルのS3キーを指定してください")
	}

	result, err := h.summarizeService.SummarizeFileByS3Key(c.Request().Context(), req.S3Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ファイル要約処理に失敗しました: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"summary": result.Summary,
	})
}
