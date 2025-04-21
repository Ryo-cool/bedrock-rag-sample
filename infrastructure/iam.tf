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
  description = "Policy for ECS tasks to access S3, Bedrock, and Textract"

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action   = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket"
          # Add more S3 actions if needed
        ]
        Effect   = "Allow"
        Resource = [
          aws_s3_bucket.documents.arn,
          "${aws_s3_bucket.documents.arn}/*"
          # Add Textract results bucket ARN if created
        ]
      },
      {
        Action = [
          "bedrock:InvokeModel", # For direct model invocation (summarization)
          "bedrock:RetrieveAndGenerate", # For Bedrock Agents RAG
          "bedrock:Retrieve"             # For Bedrock Agents RAG
          # Add specific agent/knowledge base ARN permissions if needed for more granularity
        ]
        Effect   = "Allow"
        Resource = "*" # Scope down if possible, e.g., specific model ARNs
      },
      {
        Action = [
          "textract:DetectDocumentText",
          "textract:AnalyzeDocument"
          # Add StartDocumentTextDetection, GetDocumentTextDetection etc. if using async API
        ]
        Effect   = "Allow"
        Resource = "*" # Scope down if necessary
      }
      # Add permissions for DynamoDB/OpenSearch if used for recommendations/KB metadata
    ]
  })

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-ecs-task-policy"
    }
  )
}

# Attach the custom policy to the ECS task role
resource "aws_iam_role_policy_attachment" "ecs_task_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_task_policy.arn
} 