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
// @Summary ファイルアップロード
// @Description 指定されたファイルをS3にアップロードします。
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "アップロードするファイル (最大20MB)"
// @Success 200 {object} services.UploadFileResult "アップロード成功。S3のキーとURLを含む"
// @Failure 400 {object} map[string]string "リクエスト不正 (ファイルなし、サイズ超過など)"
// @Failure 500 {object} map[string]string "サーバー内部エラー (アップロード失敗)"
// @Router /api/v1/upload [post]
func (h *UploadHandler) UploadFile(c echo.Context) error {
	// multipart/form-dataからファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		// カスタムエラーハンドラーに処理を任せる
		return echo.NewHTTPError(http.StatusBadRequest, "ファイルが見つかりませんでした")
	}

	// ファイルサイズの検証（例: 最大20MB）
	if file.Size > 20*1024*1024 {
		// カスタムエラーハンドラーに処理を任せる
		return echo.NewHTTPError(http.StatusBadRequest, "ファイルサイズが大きすぎます（最大20MB）")
	}

	// アップロードサービスにファイルを渡す
	result, err := h.uploadService.UploadFile(c.Request().Context(), file)
	if err != nil {
		// カスタムエラーハンドラーに処理を任せる
		return echo.NewHTTPError(http.StatusInternalServerError, "ファイルのアップロードに失敗しました").SetInternal(err)
	}

	return c.JSON(http.StatusOK, result)
}
