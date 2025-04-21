terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  required_version = ">= 1.0"
}

provider "aws" {
  region = var.aws_region
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