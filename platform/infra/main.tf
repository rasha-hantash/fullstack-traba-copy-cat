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

# todo update this to have domain_name, fe_subdomain, be_subdomain

# Variables
variable "aws_region" {
  description = "AWS region to deploy to"
  type        = string
  default     = "us-east-1"
}

data "aws_availability_zones" "az_availables" {
  state = "available"
}



locals {
  environment      = terraform.workspace
  create_resources = local.environment == "staging" || local.environment == "prod" ? 1 : 0
  domain_name      = "fs0ciety.dev"
  frontend_domain  = "traba-${local.environment}.${local.domain_name}"
  backend_domain   = "api-traba-${local.environment}.${local.domain_name}"
}


variable "availability_zone" {
  description = "Availability zone for resources"
  type        = string
  default     = "us-east-1a"
}


variable "frontend_image_tag" {
  description = "Tag for frontend container image"
  type        = string
}

variable "backend_image_tag" {
  description = "Tag for backend container image"
  type        = string
}

variable "ecr_repository_url" {
  description = "ECR repository URL"
  type        = string
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

variable "health_check_path_frontend" {
  description = "Health check path for frontend traba service"
  type        = string
  default     = "/" # todo do ${base_url}/api/health
}

variable "health_check_path_backend" {
  description = "Health check path for backend traba service"
  type        = string
  default     = "/health"
}

variable "auth0_us_ips" {
  description = "Auth0 US region outbound IP addresses"
  type        = list(string)
  default = [
    "174.129.105.183",
    "18.116.79.126",
    "18.117.64.128",
    "18.191.46.63",
    "18.218.26.94",
    "18.232.225.224",
    "18.233.90.226",
    "3.131.238.180",
    "3.131.55.63",
    "3.132.201.78",
    "3.133.18.220",
    "3.134.176.17",
    "3.19.44.88",
    "3.20.244.231",
    "3.21.254.195",
    "3.211.189.167",
    "34.211.191.214",
    "34.233.19.82",
    "34.233.190.223",
    "35.160.3.103",
    "35.162.47.8",
    "35.166.202.113",
    "35.167.74.121",
    "35.171.156.124",
    "35.82.131.220",
    "44.205.93.104",
    "44.218.235.21",
    "44.219.52.110",
    "52.12.243.90",
    "52.2.61.131",
    "52.204.128.250",
    "52.206.34.127",
    "52.43.255.209",
    "52.88.192.232",
    "52.89.116.72",
    "54.145.227.59",
    "54.157.101.160",
    "54.200.12.78",
    "54.209.32.202",
    "54.245.16.146",
    "54.68.157.8",
    "54.69.107.228"
  ]
}

# VPC and Network Configuration
resource "aws_vpc" "main" {
  count = local.create_resources

  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "traba-${local.environment}-vpc"
    Environment = local.environment
  }
}


/*
The cidrsubnet function takes:

Parent CIDR block (VPC's CIDR)
Number of additional bits to add to the prefix (8 gives us /24 subnets)
The subnet number (we offset private subnets by 2 to avoid overlap)

*/

resource "aws_subnet" "public" {
  count             = 2
  availability_zone = data.aws_availability_zones.az_availables.names[count.index]
  vpc_id            = aws_vpc.main[0].id
  cidr_block        = cidrsubnet(aws_vpc.main[0].cidr_block, 7, count.index + 1) //"10.0.1.0/24"
  # availability_zone       = var.availability_zone
  map_public_ip_on_launch = true

  tags = {
    Name        = "traba-${local.environment}-public"
    Environment = local.environment
  }
}

resource "aws_subnet" "private" {
  count             = 2
  availability_zone = data.aws_availability_zones.az_availables.names[count.index]
  vpc_id            = aws_vpc.main[0].id
  cidr_block        = cidrsubnet(aws_vpc.main[0].cidr_block, 7, count.index + 3) //"10.0.2.0/24"
  # availability_zone = var.availability_zone

  tags = {
    Name        = "traba-${local.environment}-private"
    Environment = local.environment
  }
}

resource "aws_internet_gateway" "main" {
  count = local.create_resources

  vpc_id = aws_vpc.main[count.index].id

  tags = {
    Name        = "traba-${local.environment}-igw"
    Environment = local.environment
  }
}

resource "aws_eip" "nat" {
  count  = local.create_resources
  domain = "vpc"
  tags = {
    Name        = "traba-${local.environment}-nat-eip"
    Environment = local.environment
  }
}

resource "aws_nat_gateway" "main" {
  count = local.create_resources

  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = {
    Name        = "traba-${local.environment}-nat"
    Environment = local.environment
  }
}

resource "aws_route_table" "public" {
  count = local.create_resources

  vpc_id = aws_vpc.main[count.index].id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main[count.index].id
  }

  tags = {
    Name        = "traba-${local.environment}-public-rt"
    Environment = local.environment
  }
}

resource "aws_route_table" "private" {
  count = local.create_resources

  vpc_id = aws_vpc.main[count.index].id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }

  tags = {
    Name        = "traba-${local.environment}-private-rt"
    Environment = local.environment
  }
}

resource "aws_route_table_association" "public" {
  count = 2

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public[0].id
}

resource "aws_route_table_association" "private" {
  count = 2

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[0].id
}

# Security Groups
resource "aws_security_group" "frontend_alb" {
  count = local.create_resources

  name_prefix = "traba-${local.environment}-frontend-alb-sg"
  vpc_id      = aws_vpc.main[count.index].id

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
    Name        = "traba-${local.environment}-frontend-alb-sg"
    Environment = local.environment
  }
}

resource "aws_security_group" "backend_alb" {
  count = local.create_resources

  name_prefix = "traba-${local.environment}-backend-alb-sg"
  vpc_id      = aws_vpc.main[count.index].id

  # Rule for health checks
  ingress {
    from_port   = 3000
    to_port     = 3000
    protocol    = "tcp"
    self        = true # Allows traffic from the ALB itself
    description = "Allow health check traffic from ALB"
    // todo add security group here? 
  }

  # Rule for frontend service and public access (for Auth0 webhooks)
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow HTTPS traffic from frontend and Auth0 webhooks"
  }

  # Rule for frontend service
  # ingress {
  #   from_port       = 443
  #   to_port         = 443
  #   protocol        = "tcp"
  #   security_groups = [aws_security_group.frontend[count.index].id]
  # }

  # # Rule for Auth0 webhook
  # ingress {
  #   from_port   = 443
  #   to_port     = 443
  #   protocol    = "tcp"
  #   cidr_blocks = [for ip in var.auth0_us_ips : "${ip}/32"]
  #   description = "Allow traffic from Auth0 webhooks"
  # }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "traba-${local.environment}-backend-alb-sg"
    Environment = local.environment
  }
}

resource "aws_security_group" "frontend" {
  count = local.create_resources

  name_prefix = "traba-${local.environment}-frontend-sg"
  vpc_id      = aws_vpc.main[count.index].id

  ingress {
    from_port       = var.frontend_container_port
    to_port         = var.frontend_container_port
    protocol        = "tcp"
    security_groups = [aws_security_group.frontend_alb[count.index].id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  #   egress {
  #   from_port       = 443
  #   to_port         = 443
  #   protocol        = "tcp"
  #   security_groups = [aws_security_group.backend_alb[count.index].id]
  # }

  tags = {
    Name        = "traba-${local.environment}-frontend-sg"
    Environment = local.environment
  }
}

resource "aws_security_group" "backend" {
  count = local.create_resources

  name_prefix = "traba-${local.environment}-backend-sg"
  vpc_id      = aws_vpc.main[count.index].id

  ingress {
    from_port       = var.backend_container_port
    to_port         = var.backend_container_port
    protocol        = "tcp"
    security_groups = [aws_security_group.backend_alb[count.index].id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "traba-${local.environment}-backend-sg"
    Environment = local.environment
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  count = local.create_resources

  name = "traba-${local.environment}-cluster"

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

  name = "traba-${local.environment}-ecs-task-execution-role"

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

  role       = aws_iam_role.ecs_task_execution_role[count.index].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Frontend Resources
resource "aws_lb" "frontend" {
  count = local.create_resources

  name               = "traba-${local.environment}-frontend-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.frontend_alb[count.index].id]
  subnets            = aws_subnet.public[*].id

  tags = {
    Name        = "traba-${local.environment}-frontend-alb"
    Environment = local.environment
  }
}

resource "aws_lb_target_group" "frontend" {
  count = local.create_resources

  name        = "traba-${local.environment}-frontend-tg"
  port        = var.frontend_container_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main[count.index].id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200,302,404" # Accept more status codes
    path                = "/"           # Check root path var.health_pathcheck_frontend
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

  load_balancer_arn = aws_lb.frontend[count.index].arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.main[count.index].arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.frontend[count.index].arn
  }
}

resource "aws_lb_listener" "frontend_http" {
  count = local.create_resources

  load_balancer_arn = aws_lb.frontend[count.index].arn
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

# Task Role - for your application to access AWS services
resource "aws_iam_role" "ecs_task_role" {
  count = local.create_resources
  name  = "traba-${local.environment}-ecs-task-role"

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
}

data "aws_caller_identity" "current" {}
# Policy to allow access to Secrets Manager
resource "aws_iam_role_policy" "ecs_task_secrets" {
  count = local.create_resources
  name  = "traba-${local.environment}-secrets-policy"
  role  = aws_iam_role.ecs_task_role[count.index].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          "arn:aws:secretsmanager:us-east-1:${data.aws_caller_identity.current.account_id}:secret:traba-${local.environment}-*"
        ]
      }
    ]
  })
}

resource "aws_ecs_task_definition" "frontend" {
  count = local.create_resources

  task_role_arn            = aws_iam_role.ecs_task_role[count.index].arn # Add this line
  family                   = "traba-${local.environment}-frontend"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role[count.index].arn

  container_definitions = jsonencode([
    {
      name  = "frontend"
      image = "${var.ecr_repository_url}:${var.frontend_image_tag}"
      portMappings = [
        {
          containerPort = var.frontend_container_port
          protocol      = "tcp"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-region"        = var.aws_region
          "awslogs-group"         = aws_cloudwatch_log_group.frontend[count.index].name
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

  name            = "traba-${local.environment}-frontend"
  cluster         = aws_ecs_cluster.main[count.index].id
  task_definition = aws_ecs_task_definition.frontend[count.index].arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private[*].id
    security_groups  = [aws_security_group.frontend[count.index].id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.frontend[count.index].arn
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

  name               = "traba-${local.environment}-backend-alb"
  internal           = true
  load_balancer_type = "application"
  security_groups    = [aws_security_group.backend_alb[count.index].id]
  subnets            = aws_subnet.public[*].id

  tags = {
    Name        = "traba-${local.environment}-backend-alb"
    Environment = local.environment
  }
}

resource "aws_lb_target_group" "backend" {
  count = local.create_resources

  name        = "traba-${local.environment}-backend-tg"
  port        = var.backend_container_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main[count.index].id
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

resource "aws_lb_listener" "backend_http" {
  count = local.create_resources

  load_balancer_arn = aws_lb.backend[count.index].arn
  port              = "3000"
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

resource "aws_lb_listener" "backend_https" {
  count = local.create_resources

  load_balancer_arn = aws_lb.backend[count.index].arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.main[count.index].arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend[count.index].arn
  }
}

# Continuing with Backend Resources...

resource "aws_ecs_task_definition" "backend" {
  count = local.create_resources

  task_role_arn            = aws_iam_role.ecs_task_role[count.index].arn # Add this line
  family                   = "traba-${local.environment}-backend"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role[count.index].arn

  container_definitions = jsonencode([
    {
      name  = "backend"
      image = "${var.ecr_repository_url}:${var.backend_image_tag}"
      portMappings = [
        {
          containerPort = var.backend_container_port
          protocol      = "tcp"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-region"        = var.aws_region
          "awslogs-group"         = aws_cloudwatch_log_group.backend[count.index].name
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

  name            = "traba-${local.environment}-backend"
  cluster         = aws_ecs_cluster.main[count.index].id
  task_definition = aws_ecs_task_definition.backend[count.index].arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private[*].id
    security_groups  = [aws_security_group.backend[count.index].id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.backend[count.index].arn
    container_name   = "backend"
    container_port   = var.backend_container_port
  }

  depends_on = [
    aws_lb_listener.backend_https,
    aws_rds_cluster.aurora_cluster,
    aws_rds_cluster_instance.aurora_instances,
  ]

  tags = {
    Environment = local.environment
  }
}

# DNS and SSL Configuration
resource "aws_acm_certificate" "main" {
  count = local.create_resources

  domain_name               = local.domain_name
  subject_alternative_names = ["*.fs0ciety.dev"]
  validation_method         = "DNS"

  tags = {
    Environment = local.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

data "aws_route53_zone" "main" {
  name         = "fs0ciety.dev" //local.domain_name
  private_zone = false
}

resource "aws_route53_record" "acm_validation" {
  for_each = local.create_resources > 0 ? {
    for dvo in aws_acm_certificate.main[0].domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.main.zone_id
}

resource "aws_acm_certificate_validation" "main" {
  count = local.create_resources

  certificate_arn         = aws_acm_certificate.main[0].arn
  validation_record_fqdns = [for record in aws_route53_record.acm_validation : record.fqdn]
}

resource "aws_route53_record" "frontend" {
  count = local.create_resources

  zone_id = data.aws_route53_zone.main.zone_id
  name    = local.frontend_domain
  type    = "A"

  alias {
    name                   = aws_lb.frontend[count.index].dns_name
    zone_id                = aws_lb.frontend[count.index].zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "backend" {
  count = local.create_resources

  zone_id = data.aws_route53_zone.main.zone_id
  name    = local.backend_domain
  type    = "A"

  alias {
    name                   = aws_lb.backend[count.index].dns_name
    zone_id                = aws_lb.backend[count.index].zone_id
    evaluate_target_health = true
  }
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "frontend" {
  count = local.create_resources

  name              = "/ecs/traba-${local.environment}-frontend"
  retention_in_days = 30

  tags = {
    Environment = local.environment
  }
}

resource "aws_cloudwatch_log_group" "backend" {
  count = local.create_resources

  name              = "/ecs/traba-${local.environment}-backend"
  retention_in_days = 30

  tags = {
    Environment = local.environment
  }
}

# Outputs
output "vpc_id" {
  value       = local.create_resources > 0 ? aws_vpc.main[0].id : null
  description = "ID of the VPC"
}

output "public_subnet_id" {
  value       = local.create_resources > 0 ? aws_subnet.public[0].id : null
  description = "ID of the public subnet"
}

output "private_subnet_id" {
  value       = local.create_resources > 0 ? aws_subnet.private[0].id : null
  description = "ID of the private subnet"
}

output "frontend_url" {
  value       = "https://${local.frontend_domain}"
  description = "URL of the frontend application"
}

output "backend_url" {
  value       = "https://${local.backend_domain}"
  description = "URL of the backend API"
}

output "frontend_alb_dns" {
  value       = local.create_resources > 0 ? aws_lb.frontend[0].dns_name : null
  description = "DNS name of the frontend load balancer"
}

output "backend_alb_dns" {
  value       = local.create_resources > 0 ? aws_lb.backend[0].dns_name : null
  description = "DNS name of the backend load balancer"
}

output "ecs_cluster_name" {
  value       = local.create_resources > 0 ? aws_ecs_cluster.main[0].name : null
  description = "Name of the ECS cluster"
}

output "frontend_security_group_id" {
  value       = local.create_resources > 0 ? aws_security_group.frontend[0].id : null
  description = "ID of the frontend security group"
}

output "backend_security_group_id" {
  value       = local.create_resources > 0 ? aws_security_group.backend[0].id : null
  description = "ID of the backend security group"
}

resource "aws_db_subnet_group" "aurora" {
  count       = local.create_resources
  name        = "traba-${local.environment}-aurora"
  description = "Subnet group for Aurora cluster"
  subnet_ids  = aws_subnet.private[*].id # Use all private subnets

  tags = {
    Environment = local.environment
  }
}


resource "aws_rds_cluster" "aurora_cluster" {
  count              = local.create_resources
  cluster_identifier = "traba-${local.environment}"
  engine             = "aurora-postgresql"
  engine_version     = "15.4"
  database_name      = "traba" # Add this line to create the database
  master_username    = "traba_admin"
  master_password    = random_password.master_password[0].result

  # Add other configuration as needed
  skip_final_snapshot = true

  # You might also want to add these configurations
  # backup_retention_period = 7
  # preferred_backup_window = "07:00-09:00"
  # deletion_protection    = local.environment == "prod" ? true : false  # Protect prod from accidental deletion

  # Add these lines to put Aurora in the correct VPC
  db_subnet_group_name   = aws_db_subnet_group.aurora[0].name
  vpc_security_group_ids = [aws_security_group.aurora_sg[0].id]

  depends_on = [random_password.master_password]
}


# Create Aurora instance(s)
resource "aws_rds_cluster_instance" "aurora_instances" {
  count = local.create_resources

  identifier         = "traba-${local.environment}-aurora-${count.index + 1}"
  cluster_identifier = aws_rds_cluster.aurora_cluster[count.index].id
  instance_class     = "db.t4g.medium" //local.environment == "prod" ? "db.r6g.large" : "db.r6g.medium"
  engine             = aws_rds_cluster.aurora_cluster[count.index].engine
  engine_version     = aws_rds_cluster.aurora_cluster[count.index].engine_version

  tags = {
    Environment = local.environment
    Service     = "database"
  }
}

# Generate random password for database
resource "random_password" "master_password" {
  count = local.create_resources

  length           = 16
  special = false
  upper = true
  lower = true
  numeric = true
}

// todo create aurora security group

# Security group for Aurora
resource "aws_security_group" "aurora_sg" {
  count = local.create_resources

  name_prefix = "traba-${local.environment}-aurora-sg"
  description = "Security group for Aurora PostgreSQL cluster"

  # Add your VPC ID here
  vpc_id = aws_vpc.main[count.index].id

  # ingress {
  #   from_port       = 5432
  #   to_port         = 5432
  #   protocol        = "tcp"
  #   security_groups = [aws_security_group.backend[count.index].id]
  # }

  # Add new rule for public access
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Allow from anywhere
    description = "Allow PostgreSQL access from anywhere"
  }

  # Add egress rule (recommended)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Environment = local.environment
    Service     = "database"
  }
}

# First, add a data source to fetch the existing secret
# Get existing secret value if it exists
# Read existing secret
data "aws_secretsmanager_secret" "backend_config" {
  count = local.create_resources
  name  = "traba-${local.environment}-backend-config" # Use name instead of secret_id
}

data "aws_secretsmanager_secret_version" "backend_config" {
  count     = local.create_resources
  secret_id = data.aws_secretsmanager_secret.backend_config[0].id
}

# Update existing secret with new values
resource "aws_secretsmanager_secret_version" "backend_config" {
  count     = local.create_resources
  secret_id = data.aws_secretsmanager_secret.backend_config[0].id

  secret_string = jsonencode(
    merge(
      jsondecode(data.aws_secretsmanager_secret_version.backend_config[0].secret_string),
      {
        CONN_STRING = "postgresql://${aws_rds_cluster.aurora_cluster[0].master_username}:${aws_rds_cluster.aurora_cluster[0].master_password}@${aws_rds_cluster.aurora_cluster[0].endpoint}:5432/${aws_rds_cluster.aurora_cluster[0].database_name}"
      }
    )
  )

  depends_on = [
    aws_rds_cluster.aurora_cluster,
    aws_rds_cluster_instance.aurora_instances
  ]
}

# Output the endpoint for reference
output "aurora_endpoint" {
  value       = local.create_resources > 0 ? aws_rds_cluster.aurora_cluster[0].endpoint : null
  description = "The endpoint of the Aurora cluster"
}

