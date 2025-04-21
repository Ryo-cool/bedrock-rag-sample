# --- ECR Repository for Backend Application --- #
resource "aws_ecr_repository" "app" {
  name                 = "${local.project_name}/app" # Repository name format: project/app
  image_tag_mutability = "MUTABLE"                 # Or IMMUTABLE if you prefer

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-app-repository"
    }
  )
} 