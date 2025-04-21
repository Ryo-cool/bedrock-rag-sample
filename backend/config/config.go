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

// DBConfig はデータベース関連の設定を保持する構造体
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Config はアプリケーション全体の設定を保持する構造体
type Config struct {
	AWS AWSConfig
	DB  DBConfig
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
		DB: DBConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			Name:     getEnvOrDefault("DB_NAME", "bedrock_rag"),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
	}
}

// GetDBConnectionString はデータベース接続文字列を返す
func (c *Config) GetDBConnectionString() string {
	return "host=" + c.DB.Host +
		" port=" + c.DB.Port +
		" user=" + c.DB.User +
		" password=" + c.DB.Password +
		" dbname=" + c.DB.Name +
		" sslmode=" + c.DB.SSLMode
}

// getEnvOrDefault は環境変数から値を取得し、存在しなければデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
