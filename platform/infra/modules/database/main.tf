# modules/database/variables.tf
variable "environment" {
  description = "Environment name (e.g., staging, prod)"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs"
  type        = list(string)
}

variable "security_group_ids" {
  description = "List of security group IDs"
  type        = list(string)
}

variable "instance_class" {
  description = "Instance class for Aurora instances"
  type        = string
  default     = "db.t4g.medium"
}

variable "engine_version" {
  description = "Aurora PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

variable "database_name" {
  description = "Name of the database to create"
  type        = string
  default     = "traba"
}

variable "backup_retention_period" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

# modules/database/main.tf
resource "aws_db_subnet_group" "aurora" {
  name        = "traba-${var.environment}-aurora"
  description = "Subnet group for Aurora cluster"
  subnet_ids  = var.private_subnet_ids

  tags = {
    Environment = var.environment
    Service     = "database"
  }
}

resource "random_password" "master_password" {
  length  = 16
  special = false
  upper   = true
  lower   = true
  numeric = true
}

resource "aws_rds_cluster" "aurora_cluster" {
  cluster_identifier = "traba-${var.environment}"
  engine            = "aurora-postgresql"
  engine_version    = var.engine_version
  database_name     = var.database_name
  master_username   = "traba_admin"
  master_password   = random_password.master_password.result

  backup_retention_period = var.backup_retention_period
  preferred_backup_window = "07:00-09:00"
  deletion_protection     = var.deletion_protection
  skip_final_snapshot     = var.environment != "prod"

  db_subnet_group_name   = aws_db_subnet_group.aurora.name
  vpc_security_group_ids = var.security_group_ids

  tags = {
    Environment = var.environment
    Service     = "database"
  }
}

resource "aws_rds_cluster_instance" "aurora_instances" {
  count = 1 # Increase for production

  identifier         = "traba-${var.environment}-aurora-${count.index + 1}"
  cluster_identifier = aws_rds_cluster.aurora_cluster.id
  instance_class     = var.instance_class
  engine             = aws_rds_cluster.aurora_cluster.engine
  engine_version     = aws_rds_cluster.aurora_cluster.engine_version

  tags = {
    Environment = var.environment
    Service     = "database"
  }
}

# Create a secret for the database connection string
resource "aws_secretsmanager_secret" "db_config" {
  name = "traba-${var.environment}-db-config"
  
  tags = {
    Environment = var.environment
    Service     = "database"
  }
}

resource "aws_secretsmanager_secret_version" "db_config" {
  secret_id = aws_secretsmanager_secret.db_config.id
  secret_string = jsonencode({
    CONN_STRING = "postgresql://${aws_rds_cluster.aurora_cluster.master_username}:${aws_rds_cluster.aurora_cluster.master_password}@${aws_rds_cluster.aurora_cluster.endpoint}:5432/${aws_rds_cluster.aurora_cluster.database_name}"
  })
}

# modules/database/outputs.tf
output "cluster_endpoint" {
  description = "The cluster endpoint"
  value       = aws_rds_cluster.aurora_cluster.endpoint
}

output "cluster_identifier" {
  description = "The cluster identifier"
  value       = aws_rds_cluster.aurora_cluster.cluster_identifier
}

output "database_name" {
  description = "The name of the database"
  value       = aws_rds_cluster.aurora_cluster.database_name
}

output "master_username" {
  description = "The master username"
  value       = aws_rds_cluster.aurora_cluster.master_username
}

output "connection_secret_arn" {
  description = "ARN of the secret containing the connection string"
  value       = aws_secretsmanager_secret.db_config.arn
}