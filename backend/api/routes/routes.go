package routes

import (
	"bedrock-rag-sample/backend/api/handlers"

	"github.com/labstack/echo/v4"
)

// SetupRoutes はAPIルートを設定する
func SetupRoutes(e *echo.Echo, uploadHandler *handlers.UploadHandler, summarizeHandler *handlers.SummarizeHandler, qaHandler *handlers.QAHandler) {
	// API v1 グループ
	v1 := e.Group("/api/v1")

	// アップロードエンドポイント
	v1.POST("/upload", uploadHandler.UploadFile)

	// 要約エンドポイント
	v1.POST("/summarize/text", summarizeHandler.SummarizeText)
	v1.POST("/summarize/file", summarizeHandler.SummarizeFile)

	// QAエンドポイント
	v1.POST("/qa", qaHandler.AskQuestion)
}
