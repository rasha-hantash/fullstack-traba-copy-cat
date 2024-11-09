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


locals {
  domain_name     = "fs0ciety.dev"
  frontend_domain = "traba-${var.environment}.${local.domain_name}"
  backend_domain  = "api-traba-${var.environment}.${local.domain_name}"
}

variable "aws_region" {
  description = "AWS region to deploy to"
  type        = string
  default     = "us-east-1"
}


variable "environment" {
  description = "AWS region to deploy to"
  type        = string
}

variable "ecr_repository_url" {
  description = "ECR repository URL"
  type        = string
}

variable "frontend_image_tag" {
  description = "Tag for frontend container image"
  type        = string
}

variable "backend_image_tag" {
  description = "Tag for backend container image"
  type        = string
}


resource "aws_iam_role" "ecs_task_execution_role" {
  name = "traba-${var.environment}-ecs-task-execution-role"

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
    Environment = var.environment
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Task Role - for your application to access AWS services
resource "aws_iam_role" "ecs_task_role" {
  name = "traba-${var.environment}-ecs-task-role"

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

  name = "traba-${var.environment}-secrets-policy"
  role = aws_iam_role.ecs_task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          "arn:aws:secretsmanager:us-east-1:${data.aws_caller_identity.current.account_id}:secret:traba-${var.environment}-*"
        ]
      }
    ]
  })
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "traba-${var.environment}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Environment = var.environment
  }
}


# environments/staging/main.tf
module "networking" {
  source = "./modules/networking"

  environment          = "staging"
  vpc_cidr             = "10.0.0.0/16"
  public_subnet_count  = 2
  private_subnet_count = 2
}

module "dns" {
  source = "./modules/dns"

  environment          = "staging"
  domain_name          = local.domain_name
  frontend_domain      = local.frontend_domain
  backend_domain       = local.backend_domain
  frontend_lb_dns_name = module.frontend.alb_dns_name
  frontend_lb_zone_id  = module.frontend.alb_zone_id
  backend_lb_dns_name  = module.backend.alb_dns_name
  backend_lb_zone_id   = module.backend.alb_zone_id
}

module "security" {
  source = "./modules/security-group"

  environment             = "staging"
  vpc_id                  = module.networking.vpc_id
  frontend_container_port = 80
  backend_container_port  = 3000
  bastion_allowed_cidrs   = ["0.0.0.0/0"] # Replace with your IP
}

# environments/staging/main.tf

# ... after networking and security modules ...

module "database" {
  source = "./modules/database"

  environment        = "staging"
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_ids = [module.security.aurora_sg_id]

  instance_class = "db.t4g.medium"
  engine_version = "15.4"
  database_name  = "traba"

  # Additional configurations for staging
  backup_retention_period = 7
  deletion_protection     = false
}

module "bastion" {
  source = "./modules/bastion"

  environment   = "staging"
  vpc_id        = module.networking.vpc_id
  vpc_security_group_ids = [module.security.bastion_sg_id]
  subnet_id     = module.networking.public_subnet_ids[0]
  allowed_cidrs = ["0.0.0.0/0"] # Replace with your IP
  instance_type = "t3.micro"
  key_name      = "bastion-key-pair-2"
}



# environments/staging/main.tf

module "frontend" {
  source = "./modules/frontend"

  environment        = "staging"
  vpc_id             = module.networking.vpc_id
  public_subnet_ids  = module.networking.public_subnet_ids
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_ids = [module.security.frontend_sg_id]
  certificate_arn    = module.dns.certificate_arn
  container_image    = "${var.ecr_repository_url}:${var.frontend_image_tag}"
  cluster_id         = aws_ecs_cluster.main.id
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn      = aws_iam_role.ecs_task_role.arn

  container_port    = 80
  health_check_path = "/"
}

module "backend" {
  source = "./modules/backend"

  environment        = "staging"
  vpc_id             = module.networking.vpc_id
  public_subnet_ids  = module.networking.public_subnet_ids
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_ids = [module.security.backend_sg_id]
  certificate_arn    = module.dns.certificate_arn
  container_image    = "${var.ecr_repository_url}:${var.backend_image_tag}"
  cluster_id         = aws_ecs_cluster.main.id
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn      = aws_iam_role.ecs_task_role.arn

  container_port    = 3000
  health_check_path = "/health"

  depends_on = [
    module.database
  ]
}

data "aws_secretsmanager_secret" "backend_config" {
  name  = "traba-${var.environment}-backend-config" # Use name instead of secret_id
}

data "aws_secretsmanager_secret_version" "backend_config" {
  secret_id = data.aws_secretsmanager_secret.backend_config.id
}

resource "aws_secretsmanager_secret_version" "backend_config" {
  secret_id = data.aws_secretsmanager_secret.backend_config.id

  secret_string = jsonencode(
    merge(
      jsondecode(data.aws_secretsmanager_secret_version.backend_config.secret_string),
      {
        CONN_STRING = "postgresql://${module.database.master_username}:${module.database.master_password}@${module.database.cluster_endpoint}:5432/${module.database.database_name}"
      }
    )
  )

  depends_on = [
    module.database
  ]
}


output "ecs_cluster_name" {
  value       = aws_ecs_cluster.main.name
  description = "Name of the ECS cluster"
}
