package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"bedrock-rag-sample/backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockClient はAmazon Bedrockとの連携を行うクライアント
type BedrockClient struct {
	client  *bedrockruntime.Client
	region  string
	modelID string
}

// NewBedrockClient は新しいBedrockClientを作成する
func NewBedrockClient(cfg *config.Config) (*BedrockClient, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	client := bedrockruntime.NewFromConfig(awsCfg)

	return &BedrockClient{
		client:  client,
		region:  cfg.AWS.Region,
		modelID: cfg.AWS.BedrockModelID,
	}, nil
}

// ClaudeInput はClaude Anthropicモデルへの入力形式
type ClaudeInput struct {
	Prompt            string  `json:"prompt"`
	MaxTokensToSample int     `json:"max_tokens_to_sample"`
	Temperature       float64 `json:"temperature"`
	TopP              float64 `json:"top_p"`
	TopK              int     `json:"top_k"`
}

// ClaudeOutput はClaude Anthropicモデルからの出力形式
type ClaudeOutput struct {
	Completion string `json:"completion"`
}

// TitanEmbeddingInput はAmazon Titan Embeddingモデルへの入力形式
type TitanEmbeddingInput struct {
	InputText string `json:"inputText"`
}

// TitanEmbeddingOutput はAmazon Titan Embeddingモデルからの出力形式
type TitanEmbeddingOutput struct {
	Embedding []float32 `json:"embedding"`
}

// GenerateSummary はテキストの要約を生成する
func (b *BedrockClient) GenerateSummary(ctx context.Context, text string) (string, error) {
	// 入力プロンプトの作成
	prompt := fmt.Sprintf(`<human>以下のテキストを100-200文字程度の日本語で要約してください。要約のみを返してください。

%s</human>

<assistant>`, text)

	return b.GenerateText(ctx, prompt)
}

// GenerateText はテキスト生成を行う
func (b *BedrockClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	input := ClaudeInput{
		Prompt:            prompt,
		MaxTokensToSample: 2048,
		Temperature:       0.7,
		TopP:              0.9,
		TopK:              250,
	}

	// リクエストボディの作成
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("入力JSONの作成に失敗しました: %w", err)
	}

	// Bedrockにリクエスト
	response, err := b.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.modelID),
		Body:        inputBytes,
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		return "", fmt.Errorf("bedrockの呼び出しに失敗しました: %w", err)
	}

	// レスポンスの解析
	var output ClaudeOutput
	if err := json.Unmarshal(response.Body, &output); err != nil {
		return "", fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}

	// 余分な空白や改行を削除
	result := strings.TrimSpace(output.Completion)

	return result, nil
}

// GenerateEmbedding はテキストからEmbeddingを生成する
func (b *BedrockClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Titan Embeddingモデル ID
	embeddingModelID := "amazon.titan-embed-text-v1"

	// 入力を準備
	input := TitanEmbeddingInput{
		InputText: text,
	}

	// リクエストボディの作成
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("入力JSONの作成に失敗しました: %w", err)
	}

	// Bedrockにリクエスト
	response, err := b.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(embeddingModelID),
		Body:        inputBytes,
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		return nil, fmt.Errorf("bedrock Embeddingの呼び出しに失敗しました: %w", err)
	}

	// レスポンスの解析
	var output TitanEmbeddingOutput
	if err := json.Unmarshal(response.Body, &output); err != nil {
		return nil, fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}

	return output.Embedding, nil
}
