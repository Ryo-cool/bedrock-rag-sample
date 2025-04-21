package routes

import (
	"bedrock-rag-sample/backend/api/handlers"

	"github.com/labstack/echo/v4"
)

// SetupRoutes はAPIルートを設定する
func SetupRoutes(e *echo.Echo, uploadHandler *handlers.UploadHandler, summarizeHandler *handlers.SummarizeHandler, qaHandler *handlers.QAHandler, documentHandler *handlers.DocumentHandler, recommendHandler *handlers.RecommendHandler) {
	// API v1 グループ
	v1 := e.Group("/api/v1")

	// アップロードエンドポイント
	v1.POST("/upload", uploadHandler.UploadFile)

	// 要約エンドポイント
	v1.POST("/summarize/text", summarizeHandler.SummarizeText)
	v1.POST("/summarize/file", summarizeHandler.SummarizeFile)

	// QAエンドポイント
	v1.POST("/qa", qaHandler.AskQuestion)

	// ドキュメント処理エンドポイント
	v1.POST("/document/process", documentHandler.ProcessDocument)

	// レコメンドエンドポイント
	v1.POST("/recommend", recommendHandler.FindSimilarDocuments)
	v1.POST("/document/:id/embedding", recommendHandler.ProcessDocumentForEmbedding)
}
