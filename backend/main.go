// @title Bedrock RAG Sample API
// @version 1.0
// @description This is a sample API server for Bedrock RAG.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"bedrock-rag-sample/backend/config"
	_ "bedrock-rag-sample/backend/docs"                   // docs パッケージをインポート (init()を実行するため)
	domain "bedrock-rag-sample/backend/internal/domain"   // エイリアス domain を指定
	"bedrock-rag-sample/backend/internal/handler"         // 修正
	dto "bedrock-rag-sample/backend/internal/handler/dto" // エイリアス dto を指定
	"bedrock-rag-sample/backend/internal/route"           // 修正

	// 修正 (エイリアス domain)
	"bedrock-rag-sample/backend/internal/services"
	"bedrock-rag-sample/backend/pkg/aws"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger" // echo-swagger をインポート
)

// init は main より先に実行される
func init() {
	// zerolog の設定
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs // Unixミリ秒形式
	zerolog.SetGlobalLevel(zerolog.InfoLevel)          // デフォルトはINFO
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		level, err := zerolog.ParseLevel(levelStr)
		if err == nil {
			zerolog.SetGlobalLevel(level)
		}
	}

	// 環境に応じて出力を変更 (例: 開発環境ではコンソールフレンドリーに)
	if os.Getenv("ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	log.Info().Msg("Logger initialized")
}

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
			// 内部エラーは詳細ログに記録
			log.Error().Err(httpError.Internal).Str("details", details).Msg("HTTP error with internal error")
		} else {
			log.Warn().Int("status", statusCode).Str("code", errorCode).Msg(message)
		}

	} else {
		// echo.HTTPError以外のエラーは内部サーバーエラーとして扱う
		// 元のエラーメッセージはログに出力し、レスポンスには含めない
		log.Error().Err(err).Msg("Unhandled internal server error")
		// 開発中はdetailsにエラーメッセージを含めることも可能
		// details = err.Error()
	}

	errorResponse := dto.ErrorResponse{ // apiModels -> dto に修正
		Error: dto.ErrorDetail{ // apiModels -> dto に修正
			Code:    errorCode,
			Message: message,
			// details は本番では基本返さない方針。
			// 開発用に返す場合は環境変数などで制御する。
			// Details: details,
		},
	}

	// Send response
	if !c.Response().Committed {
		if err := c.JSON(statusCode, errorResponse); err != nil {
			// エラーレスポンス送信自体のエラーはログに記録
			log.Error().Err(err).Msg("Failed to send error response JSON")
		}
	}
}

// zerologLoggerMiddleware は zerolog を使用したリクエストロギングミドルウェア
func zerologLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		req := c.Request()
		res := c.Response()

		// リクエストIDを取得または生成 (X-Request-ID ヘッダーがあればそれを使う)
		reqID := req.Header.Get(echo.HeaderXRequestID)
		if reqID == "" {
			reqID = fmt.Sprintf("%d", time.Now().UnixNano())
			res.Header().Set(echo.HeaderXRequestID, reqID)
		}

		// ログに基本情報を付与
		logCtx := log.With().
			Str("request_id", reqID).
			Str("remote_ip", c.RealIP()).
			Str("host", req.Host).
			Str("method", req.Method).
			Str("uri", req.RequestURI).Logger()

		// コンテキストにロガーを設定 (ハンドラー内で利用可能にする)
		c.Set("logger", logCtx)

		logCtx.Info().Msg("Request started")

		err := next(c)
		if err != nil {
			// エラーハンドラーに処理を任せるが、エラーが発生したことをログに残す
			logCtx.Error().Err(err).Msg("Request error occurred")
			c.Error(err) // 必ずカスタムエラーハンドラーに渡す
		}

		stop := time.Now()
		duration := stop.Sub(start)

		// レスポンス情報をログに追加
		logCtx.Info().
			Int("status", res.Status).
			Dur("duration", duration).
			Int64("response_size", res.Size).
			Msg("Request completed")

		return nil // エラーは既に処理済み or 発生していない
	}
}

func main() {
	// 設定を読み込む
	cfg := config.NewConfig()
	log.Info().Msg("Configuration loaded")

	// S3クライアントを初期化
	s3Client, err := aws.NewS3Client(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("S3クライアントの初期化に失敗しました")
	}
	log.Info().Msg("S3 client initialized")

	// Bedrockクライアントを初期化
	bedrockClient, err := aws.NewBedrockClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Bedrockクライアントの初期化に失敗しました")
	}
	log.Info().Msg("Bedrock client initialized")

	// Textractクライアントを初期化
	textractClient, err := aws.NewTextractClient(cfg, s3Client)
	if err != nil {
		log.Fatal().Err(err).Msg("Textractクライアントの初期化に失敗しました")
	}
	log.Info().Msg("Textract client initialized")

	// DBハンドラーを初期化
	dbHandler, err := domain.NewDBHandler(cfg) // 元の形に戻す
	if err != nil {
		log.Warn().Err(err).Msg("データベース接続に失敗しました。レコメンド機能は利用できません")
		dbHandler = nil // エラーの場合は nil を設定
	} else {
		log.Info().Msg("DB handler initialized")
		// プログラム終了時にDBコネクションを閉じる
		defer func() {
			if dbHandler != nil {
				if err := dbHandler.Close(); err != nil {
					log.Error().Err(err).Msg("Failed to close database connection")
				}
				log.Info().Msg("Database connection closed")
			}
		}()
	}

	// サービスを初期化
	uploadService := services.NewUploadService(s3Client)
	summarizeService := services.NewSummarizeService(bedrockClient, uploadService)
	log.Info().Msg("Upload and Summarize services initialized")

	// ドキュメント処理サービスを初期化
	documentService := services.NewDocumentService(textractClient, summarizeService)
	log.Info().Msg("Document service initialized")

	// レコメンドサービスを初期化
	var recommendService *services.RecommendService
	if dbHandler != nil {
		recommendService = services.NewRecommendService(bedrockClient, dbHandler)
		log.Info().Msg("Recommend service initialized")
	} else {
		log.Warn().Msg("Recommend service skipped due to DB connection failure")
	}

	// QAサービスの初期化
	qaService, err := services.NewQAService(bedrockClient, cfg)
	if err != nil {
		log.Warn().Err(err).Msg("QAサービスの初期化に失敗しました。Knowledge Base機能は利用できません。BEDROCK_KB_IDを確認してください")
	} else {
		log.Info().Msg("QA service initialized")
	}

	// ハンドラーを初期化
	uploadHandler := handler.NewUploadHandler(uploadService)
	summarizeHandler := handler.NewSummarizeHandler(summarizeService)
	documentHandler := handler.NewDocumentHandler(documentService)
	log.Info().Msg("Upload, Summarize, Document handlers initialized")

	// レコメンドハンドラーの初期化
	var recommendHandler *handler.RecommendHandler
	if recommendService != nil && dbHandler != nil {
		recommendHandler = handler.NewRecommendHandler(recommendService, dbHandler)
		log.Info().Msg("Recommend handler initialized")
	}

	// QAハンドラーの初期化（サービスが初期化できなかった場合はnilが渡される）
	var qaHandler *handler.QAHandler
	if qaService != nil {
		qaHandler = handler.NewQAHandler(qaService)
		log.Info().Msg("QA handler initialized")
	} else {
		log.Warn().Msg("QA handler skipped due to QA service initialization failure")
	}

	// Echo instance
	e := echo.New()

	// カスタムエラーハンドラーを設定
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Middleware
	e.Use(zerologLoggerMiddleware) // <- zerologベースのロガーミドルウェアを使用
	// e.Use(middleware.Logger()) // <- コメントアウト
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	log.Info().Msg("Middlewares configured")

	// Swagger UI エンドポイント
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	log.Info().Msg("Swagger UI endpoint configured at /swagger/")

	// ルートを設定
	route.SetupRoutes(e, uploadHandler, summarizeHandler, qaHandler, documentHandler, recommendHandler)
	log.Info().Msg("Routes configured")

	// ヘルスチェック用のエンドポイント
	e.GET("/health", func(c echo.Context) error {
		log.Debug().Msg("Health check requested") // デバッグレベルでログ出力
		return c.String(http.StatusOK, "Healthy")
	})

	// Start server
	serverAddress := ":8080"
	log.Info().Str("address", serverAddress).Msg("Starting server")
	e.Logger.Fatal(e.Start(serverAddress))
}
