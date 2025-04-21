package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"bedrock-rag-sample/backend/config"
	"bedrock-rag-sample/backend/internal/domain"

	_ "github.com/lib/pq"             // PostgreSQL driver
	"github.com/pgvector/pgvector-go" // pgvector
)

// PostgresDocumentRepository は PostgreSQL を使用したドキュメントリポジトリの実装
type PostgresDocumentRepository struct {
	db *sql.DB
}

// NewPostgresDocumentRepository は新しい PostgresDocumentRepository インスタンスを作成する
func NewPostgresDocumentRepository(cfg *config.Config) (*PostgresDocumentRepository, error) {
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
	return &PostgresDocumentRepository{db: db}, nil
}

// Close はデータベース接続を閉じる
func (r *PostgresDocumentRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// GetDocumentByID はIDでドキュメントを取得する (Contentは含まない)
func (r *PostgresDocumentRepository) GetDocumentByID(ctx context.Context, documentID int64) (*domain.Document, error) {
	query := `SELECT id, filename, s3_key, created_at FROM documents WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, documentID)

	var doc domain.Document
	if err := row.Scan(&doc.ID, &doc.Filename, &doc.S3Key, &doc.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found with id %d", documentID)
		}
		return nil, fmt.Errorf("failed to scan document row: %w", err)
	}
	return &doc, nil
}

// SaveDocument はドキュメント情報をデータベースに保存する
func (r *PostgresDocumentRepository) SaveDocument(ctx context.Context, doc *domain.Document) (int64, error) {
	query := `
		INSERT INTO documents (filename, s3_key, content)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var docID int64
	err := r.db.QueryRowContext(ctx, query, doc.Filename, doc.S3Key, doc.Content).Scan(&docID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert document: %w", err)
	}
	return docID, nil
}

// SaveDocumentEmbedding はドキュメントチャンクのEmbeddingを保存する
func (r *PostgresDocumentRepository) SaveDocumentEmbedding(ctx context.Context, documentID int64, chunkContent string, chunkIndex int, embedding []float32) (int64, error) {
	query := `
        INSERT INTO document_chunks (document_id, chunk_index, content, embedding)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	var chunkID int64
	err := r.db.QueryRowContext(ctx, query, documentID, chunkIndex, chunkContent, pgvector.NewVector(embedding)).Scan(&chunkID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert document chunk: %w", err)
	}
	return chunkID, nil
}

// FindSimilarChunks は指定されたEmbeddingに類似したチャンクを検索する (L2距離)
func (r *PostgresDocumentRepository) FindSimilarChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]domain.DocumentChunk, error) {
	query := `
        SELECT id, document_id, chunk_index, content, embedding <-> $1 AS similarity
        FROM document_chunks
        ORDER BY similarity
        LIMIT $2
    `
	rows, err := r.db.QueryContext(ctx, query, pgvector.NewVector(queryEmbedding), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar chunks: %w", err)
	}
	defer rows.Close()

	var chunks []domain.DocumentChunk
	for rows.Next() {
		var chunk domain.DocumentChunk
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
