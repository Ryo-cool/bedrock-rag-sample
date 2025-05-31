# bedrock.tf

# Using the standard AWS provider resource now
variable "embedding_model_arn" {
  description = "ARN of the embedding model to use for the Knowledge Base"
  type        = string
  # The aws provider might require the version suffix, adjust if needed based on apply errors
  default     = "arn:aws:bedrock:ap-northeast-1::foundation-model/amazon.titan-embed-text-v2:0" 
}

/* # コメントアウト開始
resource "aws_bedrockagent_knowledge_base" "main" {
  name        = "${local.project_name}-kb-${random_pet.suffix.id}"
  description = "Knowledge Base for the Bedrock RAG Sample application"
  role_arn    = aws_iam_role.bedrock_kb_role.arn

  knowledge_base_configuration {
    type = "VECTOR"
    vector_knowledge_base_configuration {
      embedding_model_arn = var.embedding_model_arn
    }
  }

  storage_configuration {
    # S3Configuration も必要に応じて追加 (例: オリジナルドキュメントの場所を示すメタデータなど)
    # type = "S3"
    # s3_configuration = {
    #   bucket_arn = aws_s3_bucket.documents.arn
    #   # inclusion_prefixes = ["docs/"] # オプション:特定のプレフィックスのみ対象とする場合
    # }
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-knowledge-base"
    }
  )

  # provider = awscc # Removed, using default aws provider now

  # Dependencies remain similar, but check if awscc resources are still needed
  depends_on = [
    aws_iam_role.bedrock_kb_role,
    time_sleep.wait_for_aoss_policy # Add dependency on the sleep resource
  ]
}
*/ # コメントアウト終了

# Aurora + pgvector を使用した Knowledge Base
resource "aws_bedrockagent_knowledge_base" "main" {
  name        = "${local.project_name}-kb-${random_pet.suffix.id}"
  description = "Knowledge Base for the Bedrock RAG Sample application with Aurora PostgreSQL + pgvector"
  role_arn    = aws_iam_role.bedrock_kb_role.arn

  knowledge_base_configuration {
    type = "VECTOR"
    vector_knowledge_base_configuration {
      embedding_model_arn = var.embedding_model_arn
    }
  }

  storage_configuration {
    type = "RDS"
    rds_configuration {
      resource_arn        = aws_rds_cluster.aurora.arn
      credentials_secret_arn = aws_secretsmanager_secret.aurora_credentials.arn
      database_name       = aws_rds_cluster.aurora.database_name
      table_name          = "knowledge_chunks"
      vector_field        = "embedding"
      text_field          = "text"
      metadata_field      = "metadata"
      # 必要に応じて primary_key_field も指定可能
      field_mapping = {
        primary_key_field = "id"
      }
    }
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-kb-aurora-pgvector"
    }
  )

  # Auroraクラスターが完全に準備できていることを確認するため
  depends_on = [
    aws_iam_role.bedrock_kb_role,
    aws_rds_cluster.aurora,
    aws_rds_cluster_instance.aurora_instances,
    aws_secretsmanager_secret_version.aurora_credentials
  ]
} 