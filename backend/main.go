package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"bedrock-rag-sample/backend/api/handlers"
	apiModels "bedrock-rag-sample/backend/api/models"
	"bedrock-rag-sample/backend/api/routes"
	"bedrock-rag-sample/backend/config"
	internalModels "bedrock-rag-sample/backend/internal/models"
	"bedrock-rag-sample/backend/internal/services"
	"bedrock-rag-sample/backend/pkg/aws"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// customHTTPErrorHandler はEchoのカスタムエラーハンドラー
func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return // レスポンスが既にコミットされている場合は何もしない
	}

	var (
		statusCode = http.StatusInternalServerError
		errorCode  = "INTERNAL_SERVER_ERROR"
		message    = "内部サーバーエラーが発生しました"
		details    = ""
	)

	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		statusCode = httpError.Code
		// httpError.Messageがstringの場合のみメッセージとして使用
		if msg, ok := httpError.Message.(string); ok {
			message = msg
		} else {
			// それ以外（JSONなど）の場合はステータスコードに基づいた汎用メッセージ
			message = http.StatusText(statusCode)
			if message == "" {
				message = fmt.Sprintf("HTTPエラー %d", statusCode)
			}
		}
		// エラーコードもステータスコードから推測
		switch statusCode {
		case http.StatusBadRequest:
			errorCode = "BAD_REQUEST"
		case http.StatusUnauthorized:
			errorCode = "UNAUTHORIZED"
		case http.StatusForbidden:
			errorCode = "FORBIDDEN"
		case http.StatusNotFound:
			errorCode = "NOT_FOUND"
			// 他のステータスコードに対応するエラーコードを追加可能
		}

		if httpError.Internal != nil {
			details = httpError.Internal.Error()
		}

	} else {
		// echo.HTTPError以外のエラーは内部サーバーエラーとして扱う
		// 元のエラーメッセージはログに出力し、レスポンスには含めない（本番環境を想定）
		c.Logger().Error(err)
		// 開発中はdetailsにエラーメッセージを含めることも可能
		// details = err.Error()
	}

	errorResponse := apiModels.ErrorResponse{
		Error: apiModels.ErrorDetail{
			Code:    errorCode,
			Message: message,
			Details: details, // detailsは開発時のみ含める等の制御も可能
		},
	}

	// Send response
	if !c.Response().Committed {
		if err := c.JSON(statusCode, errorResponse); err != nil {
			c.Logger().Error("カスタムエラーハンドラーでのJSON送信エラー:", err)
		}
	}
}

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

	// Textractクライアントを初期化
	textractClient, err := aws.NewTextractClient(cfg, s3Client)
	if err != nil {
		log.Fatalf("Textractクライアントの初期化に失敗しました: %v", err)
	}

	// データベースハンドラーを初期化
	dbHandler, err := internalModels.NewDBHandler(cfg)
	if err != nil {
		log.Printf("警告: データベース接続に失敗しました: %v", err)
		log.Printf("レコメンド機能は利用できません")
		dbHandler = nil
	} else {
		// プログラム終了時にDBコネクションを閉じる
		defer dbHandler.Close()
	}

	// サービスを初期化
	uploadService := services.NewUploadService(s3Client)
	summarizeService := services.NewSummarizeService(bedrockClient, uploadService)

	// ドキュメント処理サービスを初期化
	documentService := services.NewDocumentService(textractClient, summarizeService)

	// レコメンドサービスを初期化
	var recommendService *services.RecommendService
	if dbHandler != nil {
		recommendService = services.NewRecommendService(bedrockClient, dbHandler)
	}

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
	documentHandler := handlers.NewDocumentHandler(documentService)

	// レコメンドハンドラーの初期化
	var recommendHandler *handlers.RecommendHandler
	if recommendService != nil && dbHandler != nil {
		recommendHandler = handlers.NewRecommendHandler(recommendService, dbHandler)
	}

	// QAハンドラーの初期化（サービスが初期化できなかった場合はnilが渡される）
	var qaHandler *handlers.QAHandler
	if qaService != nil {
		qaHandler = handlers.NewQAHandler(qaService)
	}

	// Echo instance
	e := echo.New()

	// カスタムエラーハンドラーを設定
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// ルートを設定
	routes.SetupRoutes(e, uploadHandler, summarizeHandler, qaHandler, documentHandler, recommendHandler)

	// ヘルスチェック用のエンドポイント
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
