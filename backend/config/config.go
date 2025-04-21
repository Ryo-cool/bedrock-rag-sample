package config

import (
	"os"
)

// AWSConfig はAWS関連の設定を保持する構造体
type AWSConfig struct {
	Region          string
	S3BucketName    string
	S3DocumentsPath string
	BedrockModelID  string
	KnowledgeBaseID string
}

// Config はアプリケーション全体の設定を保持する構造体
type Config struct {
	AWS AWSConfig
}

// NewConfig は新しい設定オブジェクトを作成する
func NewConfig() *Config {
	return &Config{
		AWS: AWSConfig{
			Region:          getEnvOrDefault("AWS_REGION", "us-west-2"),
			S3BucketName:    getEnvOrDefault("S3_BUCKET_NAME", "bedrock-rag-documents"),
			S3DocumentsPath: getEnvOrDefault("S3_DOCUMENTS_PATH", "documents/"),
			BedrockModelID:  getEnvOrDefault("BEDROCK_MODEL_ID", "anthropic.claude-3-haiku-20240307-v1:0"),
			KnowledgeBaseID: getEnvOrDefault("BEDROCK_KB_ID", ""),
		},
	}
}

// getEnvOrDefault は環境変数から値を取得し、存在しなければデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
