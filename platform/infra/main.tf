# Provider and Terraform Configuration
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

# Variables
variable "aws_region" {
  description = "AWS region to deploy to"
  type        = string
  default     = "us-east-1"
}

locals {
  environment = terraform.workspace
  create_resources = local.environment == "staging" || local.environment == "prod" ? 1 : 0
}

variable "domain_name" {
  description = "Domain name for the traba application"
  type        = string
  default     = "traba-${local.environment}.fs0ciety.dev"
}

variable "availability_zone" {
  description = "Availability zone for resources"
  type        = string
  default     = "us-east-1a"
}

variable "frontend_container_image" {
  description = "Frontend container image"
  type        = string
  validation {
    condition     = length(var.frontend_container_image) > 0
    error_message = "Frontend container image must be specified"
  }
}

variable "backend_container_image" {
  description = "Backend container image"
  type        = string
  validation {
    condition     = length(var.backend_container_image) > 0
    error_message = "Backend container image must be specified"
  }
}

variable "frontend_container_port" {
  description = "Port the frontend container listens on"
  type        = number
  default     = 80
}

variable "backend_container_port" {
  description = "Port the backend container listens on"
  type        = number
  default     = 3000
}

variable "auth0_domain" {
  description = "Auth0 domain"
  type        = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9-]+\\.auth0\\.com$", var.auth0_domain))
    error_message = "Auth0 domain must be a valid auth0.com domain"
  }
}

variable "health_check_path_frontend" {
  description = "Health check path for frontend traba service"
  type        = string
  default     = "/"
}

variable "health_check_path_backend" {
  description = "Health check path for backend traba service"
  type        = string
  default     = "/health"
}

# VPC and Network Configuration
resource "aws_vpc" "main" {
  count = local.create_resources

  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${local.environment}-traba-vpc"
    Environment = local.environment
  }
}

resource "aws_subnet" "public" {
  count = local.create_resources

  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = var.availability_zone
  map_public_ip_on_launch = true

  tags = {
    Name        = "${local.environment}-traba-public"
    Environment = local.environment
  }
}

resource "aws_subnet" "private" {
  count = local.create_resources

  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = var.availability_zone

  tags = {
    Name        = "${local.environment}-traba-private"
    Environment = local.environment
  }
}

resource "aws_internet_gateway" "main" {
  count = local.create_resources

  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "${local.environment}-traba-igw"
    Environment = local.environment
  }
}

resource "aws_eip" "nat" {
  domain = "vpc"
  tags = {
    Name        = "${local.environment}-traba-nat-eip"
    Environment = local.environment
  }
}

resource "aws_nat_gateway" "main" {
  count = local.create_resources

  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public.id

  tags = {
    Name        = "${local.environment}-traba-nat"
    Environment = local.environment
  }
}

resource "aws_route_table" "public" {
  count = local.create_resources

  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name        = "${local.environment}-traba-public-rt"
    Environment = local.environment
  }
}

resource "aws_route_table" "private" {
  count = local.create_resources

  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = {
    Name        = "${local.environment}-traba-private-rt"
    Environment = local.environment
  }
}

resource "aws_route_table_association" "public" {
  count = local.create_resources

  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private.id
}

# Security Groups
resource "aws_security_group" "frontend_alb" {
  count = local.create_resources

  name_prefix = "${local.environment}-traba-frontend-alb-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
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

  tags = {
    Name        = "${local.environment}-traba-frontend-alb-sg"
    Environment = local.environment
  }
}

resource "aws_security_group" "backend_alb" {
  count = local.create_resources

  name_prefix = "${local.environment}-traba-backend-alb-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [aws_security_group.frontend.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.environment}-traba-backend-alb-sg"
    Environment = local.environment
  }
}

resource "aws_security_group" "frontend" {
  count = local.create_resources

  name_prefix = "${local.environment}-traba-frontend-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = var.frontend_container_port
    to_port         = var.frontend_container_port
    protocol        = "tcp"
    security_groups = [aws_security_group.frontend_alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.environment}-traba-frontend-sg"
    Environment = local.environment
  }
}

resource "aws_security_group" "backend" {
  count = local.create_resources

  name_prefix = "${local.environment}-traba-backend-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = var.backend_container_port
    to_port         = var.backend_container_port
    protocol        = "tcp"
    security_groups = [aws_security_group.backend_alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.environment}-traba-backend-sg"
    Environment = local.environment
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  count = local.create_resources

  name = "${local.environment}-traba-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Environment = local.environment
  }
}

# IAM Roles
resource "aws_iam_role" "ecs_task_execution_role" {
  count = local.create_resources

  name = "${local.environment}-traba-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Environment = local.environment
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  count = local.create_resources

  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Frontend Resources
resource "aws_lb" "frontend" {
  count = local.create_resources

  name               = "${local.environment}-traba-frontend-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.frontend_alb.id]
  subnets           = [aws_subnet.public.id]

  tags = {
    Name        = "${local.environment}-traba-frontend-alb"
    Environment = local.environment
  }
}

resource "aws_lb_target_group" "frontend" {
  count = local.create_resources

  name        = "${local.environment}-traba-frontend-tg"
  port        = var.frontend_container_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = var.health_check_path_frontend
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 3
  }

  tags = {
    Environment = local.environment
  }
}

resource "aws_lb_listener" "frontend_https" {
  count = local.create_resources

  load_balancer_arn = aws_lb.frontend.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.frontend.arn
  }
}

resource "aws_lb_listener" "frontend_http" {
  count = local.create_resources

  load_balancer_arn = aws_lb.frontend.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_ecs_task_definition" "frontend" {
  count = local.create_resources

  family                   = "${local.environment}-traba-frontend"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name  = "frontend"
      image = var.frontend_container_image
      portMappings = [
        {
          containerPort = var.frontend_container_port
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "NODE_ENV"
          value = local.environment
        },
        {
          name  = "BACKEND_URL"
          value = "https://api.${var.domain_name}"
        },
        {
          name  = "AUTH0_DOMAIN"
          value = var.auth0_domain
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-region"        = var.aws_region
          "awslogs-group"         = aws_cloudwatch_log_group.frontend.name
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = {
    Environment = local.environment
  }
}

resource "aws_ecs_service" "frontend" {
  count = local.create_resources

  name            = "${local.environment}-traba-frontend"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.frontend.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = [aws_subnet.private.id]
    security_groups  = [aws_security_group.frontend.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.frontend.arn
    container_name   = "frontend"
    container_port   = var.frontend_container_port
  }

  depends_on = [aws_lb_listener.frontend_https]

  tags = {
    Environment = local.environment
  }
}

# Backend Resources
resource "aws_lb" "backend" {
  count = local.create_resources

  name               = "${local.environment}-traba-backend-alb"
  internal           = true
  load_balancer_type = "application"
  security_groups    = [aws_security_group.backend_alb.id]
  subnets           = [aws_subnet.public.id]

  tags = {
    Name        = "${local.environment}-traba-backend-alb"
    Environment = local.environment
  }
}

resource "aws_lb_target_group" "backend" {
  count = local.create_resources

  name        = "${local.environment}-traba-backend-tg"
  port        = var.backend_container_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = var.health_check_path_backend
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 3
  }

  tags = {
    Environment = local.environment
  }
}

resource "aws_lb_listener" "backend_https" {
  count = local.create_resources
  
  load_balancer_arn = aws_lb.backend.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend.arn
  }
}

# Continuing with Backend Resources...

resource "aws_ecs_task_definition" "backend" {
  count = local.create_resources

  family                   = "${local.environment}-traba-backend"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name  = "backend"
      image = var.backend_container_image
      portMappings = [
        {
          containerPort = var.backend_container_port
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "NODE_ENV"
          value = local.environment
        },
        {
          name  = "AUTH0_DOMAIN"
          value = var.auth0_domain
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-region"        = var.aws_region
          "awslogs-group"         = aws_cloudwatch_log_group.backend.name
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = {
    Environment = local.environment
  }
}

resource "aws_ecs_service" "backend" {
  count = local.create_resources

  name            = "${local.environment}-traba-backend"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.backend.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = [aws_subnet.private.id]
    security_groups  = [aws_security_group.backend.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.backend.arn
    container_name   = "backend"
    container_port   = var.backend_container_port
  }

  depends_on = [aws_lb_listener.backend_https]

  tags = {
    Environment = local.environment
  }
}

# Create Aurora PostgreSQL cluster
resource "aws_rds_cluster" "aurora_cluster" {
  count = local.create_resources

  cluster_identifier     = "${local.environment}-traba-aurora"
  engine                = "aurora-postgresql"
  engine_version        = "15.4"
  database_name         = "traba"
  master_username       = "trabadmin"
  master_password       = random_password.master_password[0].result
  skip_final_snapshot   = local.environment == "staging" ? true : false
  deletion_protection   = local.environment == "prod" ? true : false
  
  vpc_security_group_ids = [aws_security_group.aurora_sg[0].id]
  
  tags = {
    Environment = local.environment
    Service     = "database"
  }
}

# Create Aurora instance(s)
resource "aws_rds_cluster_instance" "aurora_instances" {
  count = local.environment == "staging" || local.environment == "prod" ? (local.environment == "prod" ? 2 : 1) : 0

  identifier         = "${local.environment}-traba-aurora-${count.index + 1}"
  cluster_identifier = aws_rds_cluster.aurora_cluster[0].id
  instance_class     = local.environment == "prod" ? "db.r6g.large" : "db.r6g.medium"
  engine             = aws_rds_cluster.aurora_cluster[0].engine
  engine_version     = aws_rds_cluster.aurora_cluster[0].engine_version

  tags = {
    Environment = local.environment
    Service     = "database"
  }
}

# Generate random password for database
resource "random_password" "master_password" {
  count = local.create_resources

  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

// todo create aurora security group

# Security group for Aurora
resource "aws_security_group" "aurora_sg" {
  count = local.create_resources

  name_prefix = "${local.environment}-traba-aurora-sg"
  description = "Security group for Aurora PostgreSQL cluster"
  
  # Add your VPC ID here
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.backend[0].id]
  }

  tags = {
    Environment = local.environment
    Service     = "database"
  }
}

# Store database credentials in Secrets Manager
resource "aws_secretsmanager_secret_version" "aurora_credentials" {
  count = local.create_resources

  secret_id = aws_secretsmanager_secret.backend_config.id
  secret_string = jsonencode({
    database_host     = aws_rds_cluster.aurora_cluster[0].endpoint
    database_name     = aws_rds_cluster.aurora_cluster[0].database_name
    database_username = aws_rds_cluster.aurora_cluster[0].master_username
    database_password = random_password.master_password[0].result
  })
}


# DNS and SSL Configuration
resource "aws_acm_certificate" "main" {
  count = local.create_resources

  domain_name               = var.domain_name
  subject_alternative_names = ["*.${var.domain_name}"]
  validation_method        = "DNS"

  tags = {
    Environment = local.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

data "aws_route53_zone" "main" {
  name         = var.domain_name
  private_zone = false
}

resource "aws_route53_record" "acm_validation" {
  count = local.create_resources

  for_each = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.main.zone_id
}

resource "aws_acm_certificate_validation" "main" {
  count = local.create_resources

  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [for record in aws_route53_record.acm_validation : record.fqdn]
}

resource "aws_route53_record" "frontend" {
  count = local.create_resources

  zone_id = data.aws_route53_zone.main.zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = aws_lb.frontend.dns_name
    zone_id               = aws_lb.frontend.zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "backend" {
  count = local.create_resources

  zone_id = data.aws_route53_zone.main.zone_id
  name    = "api.${var.domain_name}"
  type    = "A"

  alias {
    name                   = aws_lb.backend.dns_name
    zone_id               = aws_lb.backend.zone_id
    evaluate_target_health = true
  }
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "frontend" {
  count = local.create_resources

  name              = "/ecs/${local.environment}-traba-frontend"
  retention_in_days = 30

  tags = {
    Environment = local.environment
  }
}

resource "aws_cloudwatch_log_group" "backend" {
  count = local.create_resources

  name              = "/ecs/${local.environment}-traba-backend"
  retention_in_days = 30

  tags = {
    Environment = local.environment
  }
}

# Outputs
output "vpc_id" {
  value       = aws_vpc.main.id
  description = "ID of the VPC"
}

output "public_subnet_id" {
  value       = aws_subnet.public.id
  description = "ID of the public subnet"
}

output "private_subnet_id" {
  value       = aws_subnet.private.id
  description = "ID of the private subnet"
}

output "frontend_url" {
  value       = "https://${var.domain_name}"
  description = "URL of the frontend application"
}

output "backend_url" {
  value       = "https://api.${var.domain_name}"
  description = "URL of the backend API"
}

output "frontend_alb_dns" {
  value       = aws_lb.frontend.dns_name
  description = "DNS name of the frontend load balancer"
}

output "backend_alb_dns" {
  value       = aws_lb.backend.dns_name
  description = "DNS name of the backend load balancer"
}

output "ecs_cluster_name" {
  value       = aws_ecs_cluster.main.name
  description = "Name of the ECS cluster"
}

output "frontend_security_group_id" {
  value       = aws_security_group.frontend.id
  description = "ID of the frontend security group"
}

output "backend_security_group_id" {
  value       = aws_security_group.backend.id
  description = "ID of the backend security group"
}

output "aurora_endpoint" {
  value       = local.environment == "staging" || local.environment == "prod" ? aws_rds_cluster.aurora_cluster[0].endpoint : null
  description = "The endpoint of the Aurora cluster"
}

output "aurora_reader_endpoint" {
  value       = local.environment == "staging" || local.environment == "prod" ? aws_rds_cluster.aurora_cluster[0].reader_endpoint : null
  description = "The reader endpoint of the Aurora cluster"
}