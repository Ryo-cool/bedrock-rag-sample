package handler

import (
	"bedrock-rag-sample/backend/internal/services"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// UploadHandler はファイルのアップロードに関するハンドラー
type UploadHandler struct {
	uploadService services.UploadServiceInterface
}

// NewUploadHandler は新しいUploadHandlerを生成する
func NewUploadHandler(uploadService services.UploadServiceInterface) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

// HandleUpload はファイルのアップロードリクエストを処理する
func (h *UploadHandler) HandleUpload(c echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		// エラーレスポンスを返す (例: 不正なリクエスト)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ファイルの取得に失敗しました: %v", err))
	}

	// サービス層を呼び出してファイルをアップロード
	result, err := h.uploadService.UploadFile(c.Request().Context(), fileHeader)
	if err != nil {
		// エラーレスポンスを返す (例: サーバー内部エラー)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("ファイルのアップロードに失敗しました: %v", err))
	}

	// 成功レスポンスを返す
	return c.JSON(http.StatusOK, map[string]string{
		"message":  "ファイルが正常にアップロードされました",
		"location": result.URL,
	})
}
