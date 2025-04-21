package aws

import (
	"context"
)

// BedrockClientInterface はBedrockクライアントのインターフェース
type BedrockClientInterface interface {
	GenerateSummary(ctx context.Context, text string) (string, error)
	// その他のBedrock関連メソッドをここに追加
}

// インターフェースを実装していることを静的にチェック
var _ BedrockClientInterface = (*BedrockClient)(nil)
