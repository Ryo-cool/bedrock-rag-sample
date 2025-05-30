// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repository/repository.go

// Package mock is a generated GoMock package.
package mock

import (
	domain "bedrock-rag-sample/backend/internal/domain"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDocumentRepository is a mock of DocumentRepository interface.
type MockDocumentRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDocumentRepositoryMockRecorder
}

// MockDocumentRepositoryMockRecorder is the mock recorder for MockDocumentRepository.
type MockDocumentRepositoryMockRecorder struct {
	mock *MockDocumentRepository
}

// NewMockDocumentRepository creates a new mock instance.
func NewMockDocumentRepository(ctrl *gomock.Controller) *MockDocumentRepository {
	mock := &MockDocumentRepository{ctrl: ctrl}
	mock.recorder = &MockDocumentRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDocumentRepository) EXPECT() *MockDocumentRepositoryMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockDocumentRepository) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockDocumentRepositoryMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDocumentRepository)(nil).Close))
}

// FindSimilarChunks mocks base method.
func (m *MockDocumentRepository) FindSimilarChunks(ctx context.Context, queryEmbedding []float32, limit int) ([]domain.DocumentChunk, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindSimilarChunks", ctx, queryEmbedding, limit)
	ret0, _ := ret[0].([]domain.DocumentChunk)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindSimilarChunks indicates an expected call of FindSimilarChunks.
func (mr *MockDocumentRepositoryMockRecorder) FindSimilarChunks(ctx, queryEmbedding, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindSimilarChunks", reflect.TypeOf((*MockDocumentRepository)(nil).FindSimilarChunks), ctx, queryEmbedding, limit)
}

// GetDocumentByID mocks base method.
func (m *MockDocumentRepository) GetDocumentByID(ctx context.Context, documentID int64) (*domain.Document, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDocumentByID", ctx, documentID)
	ret0, _ := ret[0].(*domain.Document)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDocumentByID indicates an expected call of GetDocumentByID.
func (mr *MockDocumentRepositoryMockRecorder) GetDocumentByID(ctx, documentID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDocumentByID", reflect.TypeOf((*MockDocumentRepository)(nil).GetDocumentByID), ctx, documentID)
}

// SaveDocument mocks base method.
func (m *MockDocumentRepository) SaveDocument(ctx context.Context, doc *domain.Document) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveDocument", ctx, doc)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveDocument indicates an expected call of SaveDocument.
func (mr *MockDocumentRepositoryMockRecorder) SaveDocument(ctx, doc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveDocument", reflect.TypeOf((*MockDocumentRepository)(nil).SaveDocument), ctx, doc)
}

// SaveDocumentEmbedding mocks base method.
func (m *MockDocumentRepository) SaveDocumentEmbedding(ctx context.Context, documentID int64, chunkContent string, chunkIndex int, embedding []float32) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveDocumentEmbedding", ctx, documentID, chunkContent, chunkIndex, embedding)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveDocumentEmbedding indicates an expected call of SaveDocumentEmbedding.
func (mr *MockDocumentRepositoryMockRecorder) SaveDocumentEmbedding(ctx, documentID, chunkContent, chunkIndex, embedding interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveDocumentEmbedding", reflect.TypeOf((*MockDocumentRepository)(nil).SaveDocumentEmbedding), ctx, documentID, chunkContent, chunkIndex, embedding)
}
