# bedrock.tf

# 使用するEmbeddingモデルのARNを変数またはローカル変数で定義
# variables.tf に追加するか、ここでローカル変数として定義
variable "embedding_model_arn" {
  description = "ARN of the embedding model to use for the Knowledge Base"
  type        = string
  default     = "arn:aws:bedrock:ap-northeast-1::foundation-model/amazon.titan-embed-text-v2" # Default to Titan Embedding v2 in Tokyo
}

resource "awscc_bedrock_knowledge_base" "main" {
  name        = "${local.project_name}-kb-${random_pet.suffix.id}"
  description = "Knowledge Base for the Bedrock RAG Sample application"
  role_arn    = aws_iam_role.bedrock_kb_role.arn

  knowledge_base_configuration = {
    type = "VECTOR"
    vector_knowledge_base_configuration = {
      embedding_model_arn = var.embedding_model_arn
    }
  }

  storage_configuration = {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration = {
      collection_arn    = awscc_opensearchserverless_collection.kb_vector_store.arn
      vector_index_name = "${local.project_name}-kb-index" # インデックス名
      field_mapping = {
        vector_field   = "vector"
        text_field     = "text"
        metadata_field = "metadata"
      }
    }
    # S3Configuration も必要に応じて追加 (例: オリジナルドキュメントの場所を示すメタデータなど)
    # type = "S3"
    # s3_configuration = {
    #   bucket_arn = aws_s3_bucket.documents.arn
    #   # inclusion_prefixes = ["docs/"] # オプション:特定のプレフィックスのみ対象とする場合
    # }
  }

  tags = merge(
    local.tags,
    { Name = "${local.project_name}-knowledge-base" }
  )

  # awscc プロバイダーを使用することを明示
  provider = awscc

  # 依存関係: IAM Role と OpenSearch Collection が先に作成される必要がある
  depends_on = [
    aws_iam_role.bedrock_kb_role,
    awscc_opensearchserverless_collection.kb_vector_store,
    # データアクセスポリシーが適用された後の方が安全な場合がある
    awscc_opensearchserverless_access_policy.data_access_policy
  ]
} 