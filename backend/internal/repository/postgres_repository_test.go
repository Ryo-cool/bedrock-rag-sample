package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"bedrock-rag-sample/backend/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*PostgresDocumentRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &PostgresDocumentRepository{db: db}
	return repo, mock, func() {
		db.Close()
	}
}

func TestGetDocumentByID(t *testing.T) {
	// テスト用のデータ
	expectedDoc := &domain.Document{
		ID:        123,
		Filename:  "test-document.pdf",
		S3Key:     "documents/test-document.pdf",
		CreatedAt: time.Now(),
	}

	// テストケース
	testCases := []struct {
		name        string
		documentID  int64
		mockSetup   func(sqlmock.Sqlmock)
		expectedDoc *domain.Document
		expectError bool
	}{
		{
			name:       "正常系: ドキュメントが見つかる",
			documentID: 123,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "filename", "s3_key", "created_at"}).
					AddRow(expectedDoc.ID, expectedDoc.Filename, expectedDoc.S3Key, expectedDoc.CreatedAt)
				mock.ExpectQuery("^SELECT (.+) FROM documents WHERE id = \\$1$").
					WithArgs(expectedDoc.ID).
					WillReturnRows(rows)
			},
			expectedDoc: expectedDoc,
			expectError: false,
		},
		{
			name:       "異常系: ドキュメントが見つからない",
			documentID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM documents WHERE id = \\$1$").
					WithArgs(int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			expectedDoc: nil,
			expectError: true,
		},
		{
			name:       "異常系: DB接続エラー",
			documentID: 123,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM documents WHERE id = \\$1$").
					WithArgs(int64(123)).
					WillReturnError(errors.New("db connection error"))
			},
			expectedDoc: nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックDBのセットアップ
			repo, mock, cleanup := setupMockDB(t)
			defer cleanup()

			// モックの振る舞いを設定
			tc.mockSetup(mock)

			// テスト対象メソッドの実行
			doc, err := repo.GetDocumentByID(context.Background(), tc.documentID)

			// 検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.Equal(t, tc.expectedDoc.ID, doc.ID)
				assert.Equal(t, tc.expectedDoc.Filename, doc.Filename)
				assert.Equal(t, tc.expectedDoc.S3Key, doc.S3Key)
			}

			// 期待されるすべてのDBコールが呼び出されたことを確認
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSaveDocument(t *testing.T) {
	// テスト用のデータ
	doc := &domain.Document{
		Filename: "new-document.pdf",
		S3Key:    "documents/new-document.pdf",
		Content:  "This is a test document content",
	}

	// テストケース
	testCases := []struct {
		name        string
		document    *domain.Document
		mockSetup   func(sqlmock.Sqlmock)
		expectedID  int64
		expectError bool
	}{
		{
			name:     "正常系: ドキュメントの保存に成功",
			document: doc,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(42)
				mock.ExpectQuery("^INSERT INTO documents").
					WithArgs(doc.Filename, doc.S3Key, doc.Content).
					WillReturnRows(rows)
			},
			expectedID:  42,
			expectError: false,
		},
		{
			name:     "異常系: ドキュメントの保存に失敗",
			document: doc,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^INSERT INTO documents").
					WithArgs(doc.Filename, doc.S3Key, doc.Content).
					WillReturnError(errors.New("insert failed"))
			},
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックDBのセットアップ
			repo, mock, cleanup := setupMockDB(t)
			defer cleanup()

			// モックの振る舞いを設定
			tc.mockSetup(mock)

			// テスト対象メソッドの実行
			id, err := repo.SaveDocument(context.Background(), tc.document)

			// 検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, int64(0), id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedID, id)
			}

			// 期待されるすべてのDBコールが呼び出されたことを確認
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSaveDocumentEmbedding(t *testing.T) {
	// テスト用のデータ
	documentID := int64(42)
	chunkContent := "This is a chunk of text"
	chunkIndex := 0
	embedding := []float32{0.1, 0.2, 0.3, 0.4}

	// テストケース
	testCases := []struct {
		name        string
		mockSetup   func(sqlmock.Sqlmock)
		expectedID  int64
		expectError bool
	}{
		{
			name: "正常系: Embeddingの保存に成功",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(101)
				mock.ExpectQuery("^INSERT INTO document_chunks").
					WithArgs(documentID, chunkIndex, chunkContent, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedID:  101,
			expectError: false,
		},
		{
			name: "異常系: Embeddingの保存に失敗",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^INSERT INTO document_chunks").
					WithArgs(documentID, chunkIndex, chunkContent, sqlmock.AnyArg()).
					WillReturnError(errors.New("insert failed"))
			},
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックDBのセットアップ
			repo, mock, cleanup := setupMockDB(t)
			defer cleanup()

			// モックの振る舞いを設定
			tc.mockSetup(mock)

			// テスト対象メソッドの実行
			id, err := repo.SaveDocumentEmbedding(context.Background(), documentID, chunkContent, chunkIndex, embedding)

			// 検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, int64(0), id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedID, id)
			}

			// 期待されるすべてのDBコールが呼び出されたことを確認
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFindSimilarChunks(t *testing.T) {
	// テスト用のデータ
	queryEmbedding := []float32{0.1, 0.2, 0.3, 0.4}
	limit := 5

	expectedChunks := []domain.DocumentChunk{
		{
			ID:         101,
			DocumentID: 42,
			ChunkIndex: 0,
			Content:    "Chunk 1 content",
			Similarity: 0.85,
		},
		{
			ID:         102,
			DocumentID: 43,
			ChunkIndex: 1,
			Content:    "Chunk 2 content",
			Similarity: 0.75,
		},
	}

	// テストケース
	testCases := []struct {
		name           string
		mockSetup      func(sqlmock.Sqlmock)
		expectedChunks []domain.DocumentChunk
		expectError    bool
	}{
		{
			name: "正常系: 類似チャンクの検索に成功",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "document_id", "chunk_index", "content", "similarity"})
				for _, chunk := range expectedChunks {
					rows.AddRow(chunk.ID, chunk.DocumentID, chunk.ChunkIndex, chunk.Content, chunk.Similarity)
				}
				mock.ExpectQuery("^SELECT (.+) FROM document_chunks").
					WithArgs(sqlmock.AnyArg(), limit).
					WillReturnRows(rows)
			},
			expectedChunks: expectedChunks,
			expectError:    false,
		},
		{
			name: "異常系: 検索に失敗",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM document_chunks").
					WithArgs(sqlmock.AnyArg(), limit).
					WillReturnError(errors.New("query failed"))
			},
			expectedChunks: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックDBのセットアップ
			repo, mock, cleanup := setupMockDB(t)
			defer cleanup()

			// モックの振る舞いを設定
			tc.mockSetup(mock)

			// テスト対象メソッドの実行
			chunks, err := repo.FindSimilarChunks(context.Background(), queryEmbedding, limit)

			// 検証
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, chunks)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedChunks), len(chunks))

				for i, expected := range tc.expectedChunks {
					assert.Equal(t, expected.ID, chunks[i].ID)
					assert.Equal(t, expected.DocumentID, chunks[i].DocumentID)
					assert.Equal(t, expected.ChunkIndex, chunks[i].ChunkIndex)
					assert.Equal(t, expected.Content, chunks[i].Content)
					assert.Equal(t, expected.Similarity, chunks[i].Similarity)
				}
			}

			// 期待されるすべてのDBコールが呼び出されたことを確認
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
