package domain

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"bedrock-rag-sample/backend/config"

	_ "github.com/lib/pq"             // PostgreSQL driver
	"github.com/pgvector/pgvector-go" // pgvector
)

// DBHandler はデータベース操作を管理する
type DBHandler struct {
	DB *sql.DB
}

// NewDBHandler は新しいDBHandlerを作成し、データベースに接続する
func NewDBHandler(cfg *config.Config) (*DBHandler, error) {
	connStr := cfg.GetDBConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		db.Close() // エラー時は接続を閉じる
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return &DBHandler{DB: db}, nil
}

// Close はデータベース接続を閉じる
func (h *DBHandler) Close() error {
	if h.DB != nil {
		return h.DB.Close()
	}
	return nil
}

// --- DocumentChunk関連のメソッド ---

// SaveDocumentEmbedding はドキュメントチャンクのEmbeddingを保存する
func (h *DBHandler) SaveDocumentEmbedding(ctx context.Context, documentID int64, chunkContent string, chunkIndex int, embedding []float32) (int64, error) {
	query := `
        INSERT INTO document_chunks (document_id, chunk_index, content, embedding)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	var chunkID int64
	err := h.DB.QueryRowContext(ctx, query, documentID, chunkIndex, chunkContent, pgvector.NewVector(embedding)).Scan(&chunkID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert document chunk: %w", err)
	}
	return chunkID, nil
}

// FindSimilarChunks は指定されたEmbeddingに類似したチャンクを検索する (L2距離)
func (h *DBHandler) FindSimilarChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]DocumentChunk, error) {
	query := `
        SELECT id, document_id, chunk_index, content, embedding <-> $1 AS similarity
        FROM document_chunks
        ORDER BY similarity
        LIMIT $2
    `
	rows, err := h.DB.QueryContext(ctx, query, pgvector.NewVector(queryEmbedding), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar chunks: %w", err)
	}
	defer rows.Close()

	var chunks []DocumentChunk
	for rows.Next() {
		var chunk DocumentChunk
		var embedding pgvector.Vector
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.ChunkIndex, &chunk.Content, &chunk.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan chunk row: %w", err)
		}
		chunk.Embedding = embedding // スキャンしたEmbeddingを設定
		chunks = append(chunks, chunk)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return chunks, nil
}

// --- Document関連のメソッド ---

// GetDocumentByID はIDでドキュメントを取得する (Contentは含まない)
func (h *DBHandler) GetDocumentByID(ctx context.Context, documentID int64) (*Document, error) {
	query := `SELECT id, filename, s3_key, created_at FROM documents WHERE id = $1`
	row := h.DB.QueryRowContext(ctx, query, documentID)

	var doc Document
	if err := row.Scan(&doc.ID, &doc.Filename, &doc.S3Key, &doc.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found with id %d", documentID)
		}
		return nil, fmt.Errorf("failed to scan document row: %w", err)
	}
	return &doc, nil
}

// ... 他に必要な Document 関連メソッドがあればここに追加 ...
