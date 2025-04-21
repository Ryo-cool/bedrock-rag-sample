# --- Application Load Balancer (ALB) --- #
resource "aws_lb" "main" {
  name               = "${local.project_name}-alb"
  internal           = false # Set to true if you only want internal access
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id # Place ALB in public subnets

  enable_deletion_protection = false # Set to true for production environments

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-alb"
    }
  )
}

# --- Target Group for ECS Service --- #
resource "aws_lb_target_group" "app" {
  name        = "${local.project_name}-app-tg"
  port        = 8080 # Port the backend application listens on
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip" # Required for Fargate

  health_check {
    enabled             = true
    interval            = 30
    path                = "/" # Adjust to your backend health check endpoint (e.g., /health)
    protocol            = "HTTP"
    matcher             = "200" # Expect HTTP 200 OK
    timeout             = 5
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-app-tg"
    }
  )
}

# --- ALB Listener for HTTP --- #
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

# --- Optional: ALB Listener for HTTPS --- #
# Requires an ACM certificate ARN
# variable "acm_certificate_arn" {
#   description = "ARN of the ACM certificate for HTTPS"
#   type        = string
#   default     = "" # Provide your certificate ARN here
# }
#
# resource "aws_lb_listener" "https" {
#   count             = var.acm_certificate_arn != "" ? 1 : 0
#   load_balancer_arn = aws_lb.main.arn
#   port              = "443"
#   protocol          = "HTTPS"
#   ssl_policy        = "ELBSecurityPolicy-2016-08" # Choose an appropriate security policy
#   certificate_arn   = var.acm_certificate_arn
#
#   default_action {
#     type             = "forward"
#     target_group_arn = aws_lb_target_group.app.arn
#   }
# }

# --- Output the DNS name of the ALB --- #
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
} 