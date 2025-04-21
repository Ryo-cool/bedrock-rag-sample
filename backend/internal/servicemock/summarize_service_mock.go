// Code generated manually - mimicking gomock style
package servicemock

import (
	"context"
	"mime/multipart"
	"reflect"

	"github.com/golang/mock/gomock"
)

// 結果の型を直接定義してインポートサイクルを避ける
type SummarizeResult struct {
	Summary    string
	SourceText string
}

// MockSummarizeServiceInterface is a mock of SummarizeServiceInterface interface.
type MockSummarizeServiceInterface struct {
	ctrl     *gomock.Controller
	recorder *MockSummarizeServiceInterfaceMockRecorder
}

// MockSummarizeServiceInterfaceMockRecorder is the mock recorder for MockSummarizeServiceInterface.
type MockSummarizeServiceInterfaceMockRecorder struct {
	mock *MockSummarizeServiceInterface
}

// NewMockSummarizeServiceInterface creates a new mock instance.
func NewMockSummarizeServiceInterface(ctrl *gomock.Controller) *MockSummarizeServiceInterface {
	mock := &MockSummarizeServiceInterface{ctrl: ctrl}
	mock.recorder = &MockSummarizeServiceInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSummarizeServiceInterface) EXPECT() *MockSummarizeServiceInterfaceMockRecorder {
	return m.recorder
}

// SummarizeText mocks base method.
func (m *MockSummarizeServiceInterface) SummarizeText(ctx context.Context, text string) (*SummarizeResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SummarizeText", ctx, text)
	ret0, _ := ret[0].(*SummarizeResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SummarizeText indicates an expected call of SummarizeText.
func (mr *MockSummarizeServiceInterfaceMockRecorder) SummarizeText(ctx, text interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SummarizeText", reflect.TypeOf((*MockSummarizeServiceInterface)(nil).SummarizeText), ctx, text)
}

// SummarizeFile mocks base method.
func (m *MockSummarizeServiceInterface) SummarizeFile(ctx context.Context, file *multipart.FileHeader) (*SummarizeResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SummarizeFile", ctx, file)
	ret0, _ := ret[0].(*SummarizeResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SummarizeFile indicates an expected call of SummarizeFile.
func (mr *MockSummarizeServiceInterfaceMockRecorder) SummarizeFile(ctx, file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SummarizeFile", reflect.TypeOf((*MockSummarizeServiceInterface)(nil).SummarizeFile), ctx, file)
}
