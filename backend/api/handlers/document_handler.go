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
// @Summary ドキュメント処理 (Textract + 要約)
// @Description アップロードされたドキュメントからTextractでテキストを抽出し、その内容を要約します。
// @Tags Document
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "処理するドキュメント (PDF/画像, 最大15MB)"
// @Success 200 {object} services.DocumentProcessResult "ドキュメント処理成功"
// @Failure 400 {object} models.ErrorResponse "リクエスト不正 (ファイルなし、サイズ超過、サポート外形式など)"
// @Failure 500 {object} models.ErrorResponse "サーバー内部エラー (処理失敗)"
// @Router /api/v1/document/process [post]
func (h *DocumentHandler) ProcessDocument(c echo.Context) error {
	// multipart/form-dataからファイルを取得
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ファイルが見つかりませんでした")
	}

	// ファイルサイズの検証（例: 最大15MB）
	if file.Size > 15*1024*1024 {
		return echo.NewHTTPError(http.StatusBadRequest, "ファイルサイズが大きすぎます（最大15MB）")
	}

	// ドキュメント処理サービスにファイルを渡す
	result, err := h.documentService.ProcessDocument(c.Request().Context(), file)
	if err != nil {
		// エラーの種類によってステータスコードを変えることも検討可能
		// 例: サポートされていないファイル形式の場合は 400 Bad Request
		// if errors.Is(err, services.ErrUnsupportedFileType) {
		// 	 return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		// }
		return echo.NewHTTPError(http.StatusInternalServerError, "ドキュメント処理に失敗しました").SetInternal(err)
	}

	return c.JSON(http.StatusOK, result)
}
