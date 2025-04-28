# --- IAM Role for ECS Task Execution --- #
# Allows ECS agent to make calls to AWS APIs on your behalf (e.g., pull ECR image, send logs to CloudWatch)
resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "${local.project_name}-ecs-task-execution-role"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-ecs-task-execution-role"
    }
  )
}

# Attach the standard managed policy for ECS task execution
resource "aws_iam_role_policy_attachment" "ecs_task_execution_policy_attachment" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# --- IAM Role for ECS Task --- #
# Allows the application running in the ECS task to interact with other AWS services
resource "aws_iam_role" "ecs_task_role" {
  name               = "${local.project_name}-ecs-task-role"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-ecs-task-role"
    }
  )
}

# --- IAM Policy for ECS Task --- #
# Defines permissions for S3, Bedrock, Textract, etc.
resource "aws_iam_policy" "ecs_task_policy" {
  name        = "${local.project_name}-ecs-task-policy"
  description = "Policy for ECS tasks to access S3, Bedrock, Textract, and KB"

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Sid    = "S3Access"
        Action   = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket"
        ]
        Effect   = "Allow"
        Resource = [
          aws_s3_bucket.documents.arn,
          "${aws_s3_bucket.documents.arn}/*"
        ]
      },
      {
        Sid    = "BedrockModelAccess"
        Action = [
          "bedrock:InvokeModel", # For direct summarization
        ]
        Effect   = "Allow"
        # Limit to specific models if possible
        Resource = [
          "arn:aws:bedrock:${var.aws_region}::foundation-model/anthropic.claude-3-haiku-20240307-v1:0", # Example Haiku
           var.embedding_model_arn # For potential embedding generation in backend
        ]
      },
      {
        Sid    = "TextractAccess"
        Action = [
          "textract:DetectDocumentText",
          "textract:AnalyzeDocument"
        ]
        Effect   = "Allow"
        Resource = "*" # Scope down if necessary
      },
      /* # コメントアウト開始
      { # Add Bedrock Knowledge Base Access
        Sid    = "BedrockKBAccess"
        Action = [
          "bedrock:Retrieve",
          "bedrock:RetrieveAndGenerate"
        ]
        Effect   = "Allow"
        # Resource should be the ARN of the Knowledge Base created in bedrock.tf
        # Using a wildcard for now, refine later if needed.
        # Consider using depends_on if referencing awscc_bedrock_knowledge_base.main.arn directly
        Resource = ["*"] # Example: "arn:aws:bedrock:${var.aws_region}:${data.aws_caller_identity.current.account_id}:knowledge-base/${awscc_bedrock_knowledge_base.main.id}"
      }
      */ # コメントアウト終了
    ]
  })

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-ecs-task-policy"
    }
  )

  # Ensure this depends on the knowledge base if referencing its ARN directly
  # depends_on = [
  #   awscc_bedrock_knowledge_base.main
  # ]
}

# Attach the custom policy to the ECS task role
resource "aws_iam_role_policy_attachment" "ecs_task_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_task_policy.arn
}

# --- IAM Role for Bedrock Knowledge Base --- #
resource "aws_iam_role" "bedrock_kb_role" {
  name = "${local.project_name}-bedrock-kb-role-${random_pet.suffix.id}"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "bedrock.amazonaws.com"
        }
      }
    ]
  })
  tags = merge(
    local.tags,
    { Name = "${local.project_name}-bedrock-kb-role" }
  )
}

# --- IAM Policy for Bedrock Knowledge Base --- #
resource "aws_iam_policy" "bedrock_kb_policy" {
  name        = "${local.project_name}-bedrock-kb-policy-${random_pet.suffix.id}"
  description = "Policy for Bedrock Knowledge Base to access S3, OpenSearch, and Embedding model"

  # Bedrock KB Role needs access to the S3 bucket where documents are stored
  # and the embedding model.
  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Sid      = "S3Access"
        Action   = ["s3:GetObject", "s3:ListBucket"]
        Effect   = "Allow"
        Resource = [
          aws_s3_bucket.documents.arn,
          "${aws_s3_bucket.documents.arn}/*"
        ]
      },
      {
        Sid      = "EmbeddingModelAccess"
        Action   = ["bedrock:InvokeModel"]
        Effect   = "Allow"
        # Use the variable defined in bedrock.tf (or create a new one)
        Resource = [var.embedding_model_arn]
      },
    ]
  })
  tags = merge(
    local.tags,
    { Name = "${local.project_name}-bedrock-kb-policy" }
  )
}

# Attach the policy to the Bedrock KB role
resource "aws_iam_role_policy_attachment" "bedrock_kb_policy_attachment" {
  role       = aws_iam_role.bedrock_kb_role.name
  policy_arn = aws_iam_policy.bedrock_kb_policy.arn
} 