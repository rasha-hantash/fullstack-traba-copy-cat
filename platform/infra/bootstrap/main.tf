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
  name                 = "traba-${local.environment}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

# Create AWS Secrets Manager secret for frontend config
resource "aws_secretsmanager_secret" "frontend_config" {
  name = "traba-${local.environment}-frontend-config"

  tags = {
    Environment = local.environment
    Service     = "frontend"
  }
}

# Create AWS Secrets Manager secret for backend config
resource "aws_secretsmanager_secret" "backend_config" {
  name = "traba-${local.environment}-backend-config"
  tags = {
    Environment = local.environment
    Service     = "backend"
  }
}

resource "aws_route53_zone" "main" {
  name = "fs0ceity.dev" # Your base domain

  tags = {
    Name = "traba-zone"
  }
}

// note this will take a while to create 
/*
aws_route53_zone.main: Still creating... [10s elapsed]
aws_route53_zone.main: Still creating... [20s elapsed]
aws_route53_zone.main: Still creating... [30s elapsed]
aws_route53_zone.main: Creation complete after 40s [id=BLAHBLAHBLAH]
 */


module "terraform_state" {
  source = "../modules/terraform-state"

  state_bucket_name   = "traba-terraform-states"
  dynamodb_table_name = "terraform-state-locks"
}

output "route53_zone_id" {
  value       = aws_route53_zone.main.zone_id
  description = "The ID of the Route53 hosted zone"
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
