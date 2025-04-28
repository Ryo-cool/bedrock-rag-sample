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
  # Use the standard 'id' attribute provided by the aws provider resource
  value       = aws_bedrockagent_knowledge_base.main.id
}

output "bedrock_knowledge_base_name" {
  description = "The name of the created Bedrock Knowledge Base"
  value       = aws_bedrockagent_knowledge_base.main.name
}