package route

import (
	"bedrock-rag-sample/backend/internal/handler"

	"github.com/labstack/echo/v4"
)

// SetupRoutes はAPIのルートを設定する
func SetupRoutes(e *echo.Echo,
	uploadHandler *handler.UploadHandler,
	summarizeHandler *handler.SummarizeHandler,
	qaHandler *handler.QAHandler,
	documentHandler *handler.DocumentHandler,
	recommendHandler *handler.RecommendHandler) {

	api := e.Group("/api/v1")

	// アップロードエンドポイント
	api.POST("/upload", uploadHandler.HandleUpload)

	// 要約エンドポイント
	api.POST("/summarize/text", summarizeHandler.HandleTextSummarize)
	api.POST("/summarize/file", summarizeHandler.HandleFileSummarize)

	// QAエンドポイント
	if qaHandler != nil {
		api.POST("/qa", qaHandler.HandleQA)
	}

	// ドキュメント処理エンドポイント
	api.POST("/document/process", documentHandler.HandleProcessDocument)

	// レコメンドエンドポイント
	if recommendHandler != nil {
		api.POST("/recommend", recommendHandler.HandleRecommend)
	}
}
