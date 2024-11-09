# modules/security/variables.tf
variable "environment" {
  description = "Environment name (e.g., staging, prod)"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
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

variable "bastion_allowed_cidrs" {
  description = "List of CIDR blocks allowed to connect to bastion"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # Should be restricted in production
}

# modules/security/main.tf
# Frontend ALB Security Group
resource "aws_security_group" "frontend_alb" {
  name_prefix = "traba-${var.environment}-frontend-alb-sg"
  vpc_id      = var.vpc_id

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
    Name        = "traba-${var.environment}-frontend-alb-sg"
    Environment = var.environment
  }
}

# Backend ALB Security Group
resource "aws_security_group" "backend_alb" {
  name_prefix = "traba-${var.environment}-backend-alb-sg"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = var.backend_container_port
    to_port     = var.backend_container_port
    protocol    = "tcp"
    self        = true
    description = "Allow health check traffic from ALB"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] // Need to be open to all IP for the webhook
    description = "Allow HTTPS traffic from frontend and Auth0 webhooks"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "traba-${var.environment}-backend-alb-sg"
    Environment = var.environment
  }
}

# Frontend ECS Service Security Group
resource "aws_security_group" "frontend" {
  name_prefix = "traba-${var.environment}-frontend-sg"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = var.frontend_container_port
    to_port         = var.frontend_container_port
    protocol        = "tcp"
    security_groups = [aws_security_group.frontend_alb.id] // todo: is this inherinting sg of alb? and if so couldn't I just specify ALB? 
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "traba-${var.environment}-frontend-sg"
    Environment = var.environment
  }
}

# Backend ECS Service Security Group
resource "aws_security_group" "backend" {
  name_prefix = "traba-${var.environment}-backend-sg"
  vpc_id      = var.vpc_id

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
    Name        = "traba-${var.environment}-backend-sg"
    Environment = var.environment
  }
}

# Aurora Security Group
resource "aws_security_group" "aurora" {
  name_prefix = "traba-${var.environment}-aurora-sg"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.backend.id]
    description     = "Allow PostgreSQL access from backend service"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "traba-${var.environment}-aurora-sg"
    Environment = var.environment
    Service     = "database"
  }
}

# Bastion Security Group
resource "aws_security_group" "bastion" {
  name_prefix = "traba-${var.environment}-bastion-sg"
  vpc_id      = var.vpc_id

  ingress {
    description = "SSH from allowed IPs"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.bastion_allowed_cidrs
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "traba-${var.environment}-bastion-sg"
    Environment = var.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Allow bastion to access Aurora
resource "aws_security_group_rule" "aurora_from_bastion" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.bastion.id
  security_group_id        = aws_security_group.aurora.id
  description             = "Allow PostgreSQL access from bastion host"
}

# modules/security/outputs.tf
output "frontend_alb_sg_id" {
  description = "ID of the frontend ALB security group"
  value       = aws_security_group.frontend_alb.id
}

output "backend_alb_sg_id" {
  description = "ID of the backend ALB security group"
  value       = aws_security_group.backend_alb.id
}

output "frontend_sg_id" {
  description = "ID of the frontend service security group"
  value       = aws_security_group.frontend.id
}

output "backend_sg_id" {
  description = "ID of the backend service security group"
  value       = aws_security_group.backend.id
}

output "aurora_sg_id" {
  description = "ID of the Aurora security group"
  value       = aws_security_group.aurora.id
}

output "bastion_sg_id" {
  description = "ID of the bastion security group"
  value       = aws_security_group.bastion.id
}