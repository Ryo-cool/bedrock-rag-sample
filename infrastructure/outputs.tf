output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.app.repository_url
}

output "s3_documents_bucket_name" {
  description = "Name of the S3 bucket for documents"
  value       = aws_s3_bucket.documents.bucket
}

output "bedrock_knowledge_base_id" {
  description = "The ID of the created Bedrock Knowledge Base"
  # awscc リソースの ID は `.id` でアクセスできますが、実際の Knowledge Base ID とは異なる場合があります。
  # Knowledge Base ID は ARN の末尾部分に含まれることが多いため、ARN から抽出するか、
  # awscc_bedrock_knowledge_base リソースの属性を確認してください。
  # 現状では ID を出力しますが、実際の値を確認してください。
  value       = awscc_bedrock_knowledge_base.main.id
}

output "bedrock_knowledge_base_name" {
  description = "The name of the created Bedrock Knowledge Base"
  value       = awscc_bedrock_knowledge_base.main.name
}

output "opensearch_collection_id" {
  description = "ID of the OpenSearch Serverless collection"
  value       = awscc_opensearchserverless_collection.kb_vector_store.id
}

output "opensearch_collection_name" {
  description = "Name of the OpenSearch Serverless collection"
  value       = awscc_opensearchserverless_collection.kb_vector_store.name
}

output "opensearch_collection_endpoint" {
  description = "Endpoint for the OpenSearch Serverless collection"
  value       = try(awscc_opensearchserverless_collection.kb_vector_store.dashboard_endpoint, "エンドポイントはまだ利用できません")
} 