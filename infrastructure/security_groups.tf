# --- Security Group for Application Load Balancer (ALB) --- #
resource "aws_security_group" "alb" {
  name        = "${local.project_name}-alb-sg"
  description = "Allow HTTP/HTTPS inbound traffic to ALB"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "HTTP from anywhere"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS from anywhere"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-alb-sg"
    }
  )
}

# --- Security Group for ECS Service (Fargate) --- #
resource "aws_security_group" "ecs_service" {
  name        = "${local.project_name}-ecs-service-sg"
  description = "Allow traffic from ALB and within VPC"
  vpc_id      = aws_vpc.main.id

  # Allow traffic from the ALB security group on the application port (e.g., 8080)
  ingress {
    description     = "Allow traffic from ALB"
    from_port       = 8080 # Adjust port if your backend listens on a different port
    to_port         = 8080 # Adjust port if your backend listens on a different port
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  # Optional: Allow all traffic from within the VPC for easier service communication
  # ingress {
  #   description = "Allow all traffic within VPC"
  #   from_port   = 0
  #   to_port     = 0
  #   protocol    = "-1"
  #   cidr_blocks = [aws_vpc.main.cidr_block]
  # }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-ecs-service-sg"
    }
  )
} 