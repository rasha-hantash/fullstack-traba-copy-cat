environment = "production"
vpc_cidr    = "10.0.0.0/16" # Production VPC CIDR

# Instance sizes - larger for production
frontend_instance_count = 2
frontend_instance_type  = "t3.medium"
backend_instance_count  = 2
backend_instance_type   = "t3.medium"
database_instance_type  = "db.t3.medium"

# Auto Scaling
min_capacity     = 2
max_capacity     = 4
desired_capacity = 2

# Database
database_backup_retention_days = 30
database_multi_az              = true # Enable multi-AZ for production redundancy