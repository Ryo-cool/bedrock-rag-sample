package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"bedrock-rag-sample/backend/config"

	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// DBHandler はデータベース操作を管理する構造体
type DBHandler struct {
	db     *sql.DB
	config *config.Config
}

// NewDBHandler は新しいDBHandlerを作成する
func NewDBHandler(cfg *config.Config) (*DBHandler, error) {
	// PostgreSQL接続
	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		return nil, fmt.Errorf("データベース接続の作成に失敗しました: %w", err)
	}

	// 接続確認
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("データベース接続の確認に失敗しました: %w", err)
	}

	handler := &DBHandler{
		db:     db,
		config: cfg,
	}

	// テーブルの初期化
	if err := handler.initializeTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("テーブル初期化に失敗しました: %w", err)
	}

	// pgvectorの初期化
	if err := handler.initializePgVector(); err != nil {
		log.Printf("警告: pgvectorの初期化に失敗しました: %v", err)
		log.Printf("類似検索機能は利用できません")
	}

	return handler, nil
}

// Close はDBの接続を閉じる
func (h *DBHandler) Close() error {
	return h.db.Close()
}

// initializeTables はテーブルを初期化する
func (h *DBHandler) initializeTables() error {
	// ドキュメントテーブルの作成
	_, err := h.db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			summary TEXT,
			file_type TEXT,
			s3_key TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("documentsテーブルの作成に失敗しました: %w", err)
	}

	// ベクトルテーブルの作成（pgvector拡張が必要）
	_, err = h.db.Exec(`
		CREATE TABLE IF NOT EXISTS document_embeddings (
			id SERIAL PRIMARY KEY,
			document_id INTEGER REFERENCES documents(id) ON DELETE CASCADE,
			embedding vector(1536),
			chunk_text TEXT NOT NULL,
			chunk_index INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("警告: document_embeddingsテーブルの作成に失敗しました: %v", err)
		log.Printf("pgvector拡張がインストールされていない可能性があります")
	}

	return nil
}

// initializePgVector はpgvector拡張を初期化する
func (h *DBHandler) initializePgVector() error {
	// pgvector拡張の作成（管理者権限が必要）
	_, err := h.db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("pgvector拡張の作成に失敗しました: %w", err)
	}

	// インデックスの作成
	_, err = h.db.Exec("CREATE INDEX IF NOT EXISTS document_embeddings_embedding_idx ON document_embeddings USING ivfflat (embedding vector_l2_ops)")
	if err != nil {
		return fmt.Errorf("embeddingインデックスの作成に失敗しました: %w", err)
	}

	return nil
}

// Document はドキュメント情報を表す構造体
type Document struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Summary   string `json:"summary,omitempty"`
	FileType  string `json:"file_type,omitempty"`
	S3Key     string `json:"s3_key,omitempty"`
	CreatedAt string `json:"created_at"`
}

// DocumentChunk はドキュメントのチャンク（部分）を表す構造体
type DocumentChunk struct {
	ID         int64     `json:"id"`
	DocumentID int64     `json:"document_id"`
	ChunkText  string    `json:"chunk_text"`
	ChunkIndex int       `json:"chunk_index"`
	Embedding  []float32 `json:"-"` // レスポンスには含めない
	Similarity float64   `json:"similarity,omitempty"`
}

// SaveDocument はドキュメントをデータベースに保存する
func (h *DBHandler) SaveDocument(ctx context.Context, doc *Document) (int64, error) {
	var id int64
	err := h.db.QueryRowContext(
		ctx,
		`INSERT INTO documents (title, content, summary, file_type, s3_key) 
		 VALUES ($1, $2, $3, $4, $5) 
		 RETURNING id`,
		doc.Title, doc.Content, doc.Summary, doc.FileType, doc.S3Key,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ドキュメントの保存に失敗しました: %w", err)
	}

	return id, nil
}

// SaveDocumentEmbedding はドキュメントEmbeddingをデータベースに保存する
func (h *DBHandler) SaveDocumentEmbedding(ctx context.Context, docID int64, chunkText string, chunkIndex int, embedding []float32) (int64, error) {
	var id int64
	err := h.db.QueryRowContext(
		ctx,
		`INSERT INTO document_embeddings (document_id, chunk_text, chunk_index, embedding) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id`,
		docID, chunkText, chunkIndex, pgvector.NewVector(embedding),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ドキュメントEmbeddingの保存に失敗しました: %w", err)
	}

	return id, nil
}

// FindSimilarChunks は与えられたEmbeddingに類似したチャンクを検索する
func (h *DBHandler) FindSimilarChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]DocumentChunk, error) {
	if limit <= 0 {
		limit = 5 // デフォルト値
	}

	rows, err := h.db.QueryContext(
		ctx,
		`SELECT 
			e.id, e.document_id, e.chunk_text, e.chunk_index, 
			1 - (e.embedding <-> $1) as similarity
		FROM document_embeddings e
		ORDER BY e.embedding <-> $1
		LIMIT $2`,
		pgvector.NewVector(queryEmbedding), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("類似チャンクの検索に失敗しました: %w", err)
	}
	defer rows.Close()

	chunks := make([]DocumentChunk, 0, limit)
	for rows.Next() {
		var chunk DocumentChunk
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.ChunkText, &chunk.ChunkIndex, &chunk.Similarity); err != nil {
			return nil, fmt.Errorf("検索結果の読み取りに失敗しました: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// GetDocumentByID はIDからドキュメントを取得する
func (h *DBHandler) GetDocumentByID(ctx context.Context, id int64) (*Document, error) {
	row := h.db.QueryRowContext(
		ctx,
		`SELECT id, title, content, summary, file_type, s3_key, created_at 
		 FROM documents 
		 WHERE id = $1`,
		id,
	)

	var doc Document
	err := row.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.Summary, &doc.FileType, &doc.S3Key, &doc.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ドキュメントが見つかりません: %d", id)
		}
		return nil, fmt.Errorf("ドキュメントの取得に失敗しました: %w", err)
	}

	return &doc, nil
}
