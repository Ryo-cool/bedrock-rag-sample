package handler

import (
	"bedrock-rag-sample/backend/internal/services"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// DocumentHandler はドキュメント処理に関するハンドラー
type DocumentHandler struct {
	documentService services.DocumentServiceInterface
}

// NewDocumentHandler は新しいDocumentHandlerを生成する
func NewDocumentHandler(documentService services.DocumentServiceInterface) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// ProcessDocumentRequest はドキュメント処理リクエストの構造体
type ProcessDocumentRequest struct {
	S3Key string `json:"s3_key"`
}

// HandleProcessDocument は指定されたS3ファイルの処理リクエストを処理する
func (h *DocumentHandler) HandleProcessDocument(c echo.Context) error {
	var req ProcessDocumentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "不正なリクエストです")
	}

	if req.S3Key == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "処理するファイルのS3キーを指定してください")
	}

	result, err := h.documentService.ProcessDocumentByS3Key(c.Request().Context(), req.S3Key)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ドキュメント処理に失敗しました: %v", err))
	}

	// 成功レスポンスを返す (例: 抽出されたテキストや要約を含む)
	return c.JSON(http.StatusOK, result)
}
