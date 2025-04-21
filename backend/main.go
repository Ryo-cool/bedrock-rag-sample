package main

import (
	"log"
	"net/http"

	"bedrock-rag-sample/backend/api/handlers"
	"bedrock-rag-sample/backend/api/routes"
	"bedrock-rag-sample/backend/config"
	"bedrock-rag-sample/backend/internal/services"
	"bedrock-rag-sample/backend/pkg/aws"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 設定を読み込む
	cfg := config.NewConfig()

	// S3クライアントを初期化
	s3Client, err := aws.NewS3Client(cfg)
	if err != nil {
		log.Fatalf("S3クライアントの初期化に失敗しました: %v", err)
	}

	// Bedrockクライアントを初期化
	bedrockClient, err := aws.NewBedrockClient(cfg)
	if err != nil {
		log.Fatalf("Bedrockクライアントの初期化に失敗しました: %v", err)
	}

	// サービスを初期化
	uploadService := services.NewUploadService(s3Client)
	summarizeService := services.NewSummarizeService(bedrockClient, uploadService)

	// QAサービスの初期化
	qaService, err := services.NewQAService(bedrockClient, cfg)
	if err != nil {
		// KBIDが設定されていない場合はエラーログを出すが、サーバー起動はブロックしない
		log.Printf("警告: QAサービスの初期化に失敗しました: %v", err)
		log.Printf("Knowledge Base機能を使用するには環境変数BEDROCK_KB_IDを設定してください")
	}

	// ハンドラーを初期化
	uploadHandler := handlers.NewUploadHandler(uploadService)
	summarizeHandler := handlers.NewSummarizeHandler(summarizeService)

	// QAハンドラーの初期化（サービスが初期化できなかった場合はnilが渡される）
	var qaHandler *handlers.QAHandler
	if qaService != nil {
		qaHandler = handlers.NewQAHandler(qaService)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// ルートを設定
	routes.SetupRoutes(e, uploadHandler, summarizeHandler, qaHandler)

	// ヘルスチェック用のエンドポイント
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
