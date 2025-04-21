package services

import (
	"context"
)

// QAServiceInterface はQAサービスのインターフェース
type QAServiceInterface interface {
	SimpleRAG(ctx context.Context, query string) (*QAResult, error)
	// 他の QAService メソッドが必要であればここに追加
}

// TODO: QAService が QAServiceInterface を実装していることを静的にチェックする
// 例: var _ QAServiceInterface = (*QAService)(nil)
