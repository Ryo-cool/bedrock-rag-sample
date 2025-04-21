variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "ap-northeast-1" # Tokyo Region
}

variable "project_name" {
  description = "Name of the project, used for resource tagging and naming"
  type        = string
  default     = "bedrock-rag-sample"
}

variable "environment" {
  description = "Deployment environment (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
} 