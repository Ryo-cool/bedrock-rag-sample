package handlers

import (
	"net/http"

	"bedrock-rag-sample/backend/internal/services"

	"github.com/labstack/echo/v4"
)

// UploadHandler はファイルアップロード関連のエンドポイントを処理するハンドラー
type UploadHandler struct {
	uploadService *services.UploadService
}

// NewUploadHandler は新しいUploadHandlerを作成する
func NewUploadHandler(uploadService *services.UploadService) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

// UploadFile はファイルをアップロードするエンドポイントを処理する
func (h *UploadHandler) UploadFile(c echo.Context) error {
	// multipart/form-dataからファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ファイルが見つかりませんでした",
		})
	}

	// ファイルサイズの検証（例: 最大20MB）
	if file.Size > 20*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ファイルサイズが大きすぎます（最大20MB）",
		})
	}

	// アップロードサービスにファイルを渡す
	result, err := h.uploadService.UploadFile(c.Request().Context(), file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "ファイルのアップロードに失敗しました: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
