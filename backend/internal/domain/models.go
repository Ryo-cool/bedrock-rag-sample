package domain

import (
	"time"

	"github.com/pgvector/pgvector-go" // pgvector
)

// Document はドキュメント情報を表す構造体
type Document struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	S3Key     string    `json:"s3_key"`
	Content   string    `json:"content,omitempty"` // 必要に応じて読み込む
	CreatedAt time.Time `json:"created_at"`
}

// DocumentChunk はドキュメントのチャンクとEmbeddingを表す構造体
type DocumentChunk struct {
	ID         int64           `json:"id"`
	DocumentID int64           `json:"document_id"`
	ChunkIndex int             `json:"chunk_index"`
	Content    string          `json:"content"`
	Embedding  pgvector.Vector `json:"-"`                    // JSONには含めない
	Similarity float64         `json:"similarity,omitempty"` // 類似度検索の結果で使用
}
