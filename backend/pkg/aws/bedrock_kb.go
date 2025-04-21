package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"bedrock-rag-sample/backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"

	// bedrockagenttypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types" // 不要なインポートをコメントアウト
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types" // runtimeの型のみ使用
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockKBClient はBedrock Knowledge Base機能を扱うクライアント
type BedrockKBClient struct {
	agentClient        *bedrockagent.Client
	runtimeClient      *bedrockruntime.Client
	agentRuntimeClient *bedrockagentruntime.Client
	region             string
	kbId               string // Knowledge Base ID
	modelId            string // 使用するモデルID
}

// NewBedrockKBClient は新しいBedrockKBClientを作成する
func NewBedrockKBClient(cfg *config.Config) (*BedrockKBClient, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(cfg.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	agentClient := bedrockagent.NewFromConfig(awsCfg)
	runtimeClient := bedrockruntime.NewFromConfig(awsCfg)
	agentRuntimeClient := bedrockagentruntime.NewFromConfig(awsCfg)

	// 環境変数からKnowledge Base IDを取得
	kbId := os.Getenv("BEDROCK_KB_ID")
	if kbId == "" {
		return nil, fmt.Errorf("BEDROCK_KB_IDが設定されていません")
	}

	// デフォルトでClaude 3 Haikuを使用
	modelId := getEnvOrDefault("BEDROCK_MODEL_ID", "anthropic.claude-3-haiku-20240307-v1:0")

	return &BedrockKBClient{
		agentClient:        agentClient,
		runtimeClient:      runtimeClient,
		agentRuntimeClient: agentRuntimeClient,
		region:             cfg.AWS.Region,
		kbId:               kbId,
		modelId:            modelId,
	}, nil
}

// ClaudeRAGInput はClaude RAGリクエストの入力形式
type ClaudeRAGInput struct {
	Prompt            string  `json:"prompt"`
	MaxTokensToSample int     `json:"max_tokens_to_sample"`
	Temperature       float64 `json:"temperature"`
	TopP              float64 `json:"top_p"`
	TopK              int     `json:"top_k"`
}

// ClaudeOutput はClaude出力形式
// type ClaudeOutput struct {
// 	Completion string `json:"completion"`
// }

// RAGRetrieveResult はRetrieveオペレーションの結果
type RAGRetrieveResult struct {
	RetrievedReferences []RetrievedReference `json:"retrieved_references"`
	Query               string               `json:"query"`
}

// RetrievedReference は検索された参照情報
type RetrievedReference struct {
	Content    string  `json:"content"`
	Location   string  `json:"location"`
	Metadata   string  `json:"metadata"`
	Score      float64 `json:"score"`
	DocumentId string  `json:"document_id"` // Retrieve APIのレスポンスに直接IDが含まれない場合がある
}

// RetrieveFromKB はKnowledge Baseからクエリに関連するドキュメントを検索する
func (b *BedrockKBClient) RetrieveFromKB(ctx context.Context, query string) (*RAGRetrieveResult, error) {
	// agentRuntime.Retrieve APIを呼び出す
	resp, err := b.agentRuntimeClient.Retrieve(ctx, &bedrockagentruntime.RetrieveInput{
		KnowledgeBaseId: aws.String(b.kbId),
		RetrievalQuery: &types.KnowledgeBaseQuery{ // ← types.RetrievalQuery → types.KnowledgeBaseQuery
			Text: aws.String(query),
		},
		RetrievalConfiguration: &types.KnowledgeBaseRetrievalConfiguration{
			VectorSearchConfiguration: &types.KnowledgeBaseVectorSearchConfiguration{
				NumberOfResults: aws.Int32(5),
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("knowledge Baseからの情報取得に失敗しました: %w", err)
	}

	result := &RAGRetrieveResult{
		Query:               query,
		RetrievedReferences: make([]RetrievedReference, 0),
	}

	// 検索結果をマッピング
	for _, item := range resp.RetrievalResults {
		if item.Content == nil || item.Content.Text == nil {
			continue
		}

		ref := RetrievedReference{
			Content: *item.Content.Text,
		}

		if item.Location != nil && item.Location.Type != "" {
			if item.Location.Type == types.RetrievalResultLocationTypeS3 {
				ref.Location = *item.Location.S3Location.Uri
			}
		}

		if item.Metadata != nil {
			metadataBytes, _ := json.Marshal(item.Metadata)
			ref.Metadata = string(metadataBytes)
		}

		// DocumentIdは直接アクセスできない場合があるため、コメントアウトまたは別の方法で取得
		// if item.DocumentId != nil {
		// 	ref.DocumentId = *item.DocumentId
		// }

		result.RetrievedReferences = append(result.RetrievedReferences, ref)
	}

	return result, nil
}

// RAGQueryWithKB はKnowledge Baseを使用したRAGベースのクエリを実行する
func (b *BedrockKBClient) RAGQueryWithKB(ctx context.Context, query string, references []RetrievedReference) (string, error) {
	// Claude 3 Haiku モデルID
	modelID := b.modelId

	// 参考情報をプロンプトに組み込む
	var sb strings.Builder
	sb.WriteString("以下は関連するドキュメントからの情報です：\n\n")

	for i, ref := range references {
		sb.WriteString(fmt.Sprintf("文書[%d]: %s\n", i+1, ref.Content))
	}

	// 入力プロンプトの作成
	prompt := fmt.Sprintf(`<human>%s

質問: %s</human>

<assistant>`, sb.String(), query)

	input := ClaudeRAGInput{
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
	response, err := b.runtimeClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Body:        inputBytes,
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		return "", fmt.Errorf("bedrockの呼び出しに失敗しました: %w", err)
	}

	// レスポンスの解析
	var output ClaudeOutput // このClaudeOutputはbedrock.goで定義されている想定
	if err := json.Unmarshal(response.Body, &output); err != nil {
		return "", fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}

	return strings.TrimSpace(output.Completion), nil
}

// RAGQueryWithRetrieveAndGenerate はBedrockのRetrieveAndGenerate APIを使用してKnowledge Baseに基づく回答を生成する
func (b *BedrockKBClient) RAGQueryWithRetrieveAndGenerate(ctx context.Context, query string) (string, error) {
	// RetrieveAndGenerate APIの呼び出し
	input := &bedrockagentruntime.RetrieveAndGenerateInput{
		Input: &types.RetrieveAndGenerateInput{
			Text: aws.String(query),
		},
		RetrieveAndGenerateConfiguration: &types.RetrieveAndGenerateConfiguration{
			Type: types.RetrieveAndGenerateTypeKnowledgeBase,
			KnowledgeBaseConfiguration: &types.KnowledgeBaseRetrieveAndGenerateConfiguration{
				KnowledgeBaseId: aws.String(b.kbId),
				ModelArn:        aws.String(b.modelId),
				GenerationConfiguration: &types.GenerationConfiguration{
					PromptTemplate: &types.PromptTemplate{
						TextPromptTemplate: aws.String("以下の質問に日本語で答えてください。\n質問: ${question}"),
					},
					InferenceConfig: &types.InferenceConfig{
						TextInferenceConfig: &types.TextInferenceConfig{
							MaxTokens:   aws.Int32(2048),
							Temperature: aws.Float32(0.7),
							TopP:        aws.Float32(0.9),
						},
					},
				},
			},
		},
	}

	resp, err := b.agentRuntimeClient.RetrieveAndGenerate(ctx, input)
	if err != nil {
		return "", fmt.Errorf("RetrieveAndGenerateの呼び出しに失敗しました: %w", err)
	}

	if resp.Output == nil || resp.Output.Text == nil {
		return "", fmt.Errorf("回答の生成に失敗しました: 出力がありません")
	}

	return *resp.Output.Text, nil
}

// getEnvOrDefault は環境変数から値を取得し、存在しなければデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
