terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.50.0"
    }
  }
}

locals {
  environment = terraform.workspace
}

provider "aws" {
  region = var.aws_region
}

# Create ECR repository
resource "aws_ecr_repository" "traba" {
  name                 = "${local.environment}-traba"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

# Create AWS Secrets Manager secret for frontend config
resource "aws_secretsmanager_secret" "frontend_config" {
  name = "${local.environment}-frontend-config"

  tags = {
    Environment = local.environment
    Service     = "frontend"
  }
}

# Create AWS Secrets Manager secret for backend config
resource "aws_secretsmanager_secret" "backend_config" {
  name = "${local.environment}-backend-config"
  tags = {
    Environment = local.environment
    Service     = "backend"
  }
}

# Outputs
output "ecr_repository_url" {
  value       = aws_ecr_repository.traba.repository_url
  description = "The URL of the ECR repository"
}

output "frontend_secret_arn" {
  value       = aws_secretsmanager_secret.frontend_config.arn
  description = "The ARN of the Secrets Manager secret"
}

output "backend_secret_arn" {
  value       = aws_secretsmanager_secret.backend_config.arn
  description = "The ARN of the Secrets Manager secret"
}

# Variables (define these in a separate variables.tf file)
variable "aws_region" {
  description = "The AWS region to create resources in"
  default     = "us-east-1"
}
