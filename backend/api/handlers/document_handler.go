package handlers

import (
	"net/http"

	"bedrock-rag-sample/backend/internal/services"

	"github.com/labstack/echo/v4"
)

// DocumentHandler はドキュメント処理関連のエンドポイントを処理するハンドラー
type DocumentHandler struct {
	documentService *services.DocumentService
}

// NewDocumentHandler は新しいDocumentHandlerを作成する
func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// ProcessDocument はドキュメントを処理するエンドポイントを処理する
func (h *DocumentHandler) ProcessDocument(c echo.Context) error {
	// multipart/form-dataからファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ファイルが見つかりませんでした",
		})
	}

	// ファイルサイズの検証（例: 最大15MB）
	if file.Size > 15*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ファイルサイズが大きすぎます（最大15MB）",
		})
	}

	// ドキュメント処理サービスにファイルを渡す
	result, err := h.documentService.ProcessDocument(c.Request().Context(), file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "ドキュメント処理に失敗しました: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
