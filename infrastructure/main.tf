terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    awscc = {
      source  = "hashicorp/awscc"
      version = "~> 0.60"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }

  required_version = ">= 1.0"
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}

provider "awscc" {
  region  = var.aws_region
  profile = var.aws_profile
}

# --- Resource definitions will go here --- #

# Example: Random Pet name for unique resource naming
resource "random_pet" "suffix" {
  length = 2
}

locals {
  project_name = var.project_name
  tags = {
    Project     = var.project_name
    Environment = var.environment
    Terraform   = "true"
  }
}

variable "aws_profile" {
  description = "AWS profile to use for authentication"
  type        = string
  default     = "bedrock-sso"
} 