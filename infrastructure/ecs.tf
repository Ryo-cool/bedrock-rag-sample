# --- ECS Cluster --- #
resource "aws_ecs_cluster" "main" {
  name = "${local.project_name}-cluster"

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-cluster"
    }
  )
}

# --- CloudWatch Log Group for ECS Task --- #
resource "aws_cloudwatch_log_group" "ecs_log_group" {
  name              = "/ecs/${local.project_name}-app"
  retention_in_days = 30 # Adjust retention as needed

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-ecs-log-group"
    }
  )
}

# --- ECS Task Definition --- #
resource "aws_ecs_task_definition" "app" {
  family                   = "${local.project_name}-app-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"  # Adjust CPU units as needed (e.g., 256 = 0.25 vCPU)
  memory                   = "512"  # Adjust memory in MiB as needed
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  # Define the container for the backend application
  container_definitions = jsonencode([
    {
      name      = "${local.project_name}-app-container"
      image     = aws_ecr_repository.app.repository_url # Will be replaced with specific image tag during deployment
      essential = true
      portMappings = [
        {
          containerPort = 8080 # Port the Go application listens on inside the container
          hostPort      = 8080 # For awsvpc mode, hostPort is typically same as containerPort
          protocol      = "tcp"
        }
      ]
      # Define environment variables if needed
      # environment = [
      #   { name = "EXAMPLE_VAR", value = "example_value" }
      # ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs_log_group.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-app-task-def"
    }
  )
}

# --- ECS Service --- #
# This service will run and maintain the desired number of tasks for the application
resource "aws_ecs_service" "app" {
  name            = "${local.project_name}-app-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 1 # Start with one task, can be adjusted or autoscaled
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id # Run tasks in private subnets
    security_groups = [aws_security_group.ecs_service.id]
    assign_public_ip = false # Tasks in private subnets do not need public IPs
  }

  # Link the service to the Application Load Balancer (defined in alb.tf)
  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn # Reference the target group defined later
    container_name   = "${local.project_name}-app-container"
    container_port   = 8080 # Port defined in the task definition
  }

  # Ensure task definition is created before the service
  depends_on = [aws_lb_target_group.app] # Also depends on the LB target group

  # Optional: Service discovery, deployment configuration, health check grace period etc.
  health_check_grace_period_seconds = 60

  tags = merge(
    local.tags,
    {
      Name = "${local.project_name}-app-service"
    }
  )
} 