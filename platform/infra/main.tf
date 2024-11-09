# environments/staging/main.tf
module "networking" {
  source = "../../modules/networking"

  environment         = "staging"
  vpc_cidr           = "10.0.0.0/16"
  public_subnet_count = 2
  private_subnet_count = 2
}

module "security" {
  source = "../../modules/security"

  environment            = "staging"
  vpc_id                = module.networking.vpc_id
  frontend_container_port = 80
  backend_container_port = 3000
  bastion_allowed_cidrs = ["YOUR_IP/32"] # Replace with your IP
}

# environments/staging/main.tf

# ... after networking and security modules ...

module "database" {
  source = "../../modules/database"

  environment         = "staging"
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_ids = [module.security.aurora_sg_id]
  
  instance_class     = "db.t4g.medium"
  engine_version     = "15.4"
  database_name      = "traba"
  
  # Additional configurations for staging
  backup_retention_period = 7
  deletion_protection    = false
}

module "bastion" {
  source = "../../modules/bastion"

  environment    = "staging"
  vpc_id        = module.networking.vpc_id
  subnet_id     = module.networking.public_subnet_ids[0]
  allowed_cidrs = ["YOUR_IP/32"]  # Replace with your IP
  instance_type = "t3.micro"
  key_name      = "bastion-key-pair"
}

# Add a security group rule to allow bastion access to Aurora
resource "aws_security_group_rule" "aurora_from_bastion" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = module.bastion.bastion_security_group_id
  security_group_id        = module.security.aurora_sg_id
  description             = "Allow PostgreSQL access from bastion host"
}

# environments/staging/main.tf

module "frontend" {
  source = "../../modules/frontend"

  environment        = "staging"
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_ids = [module.security.frontend_sg_id]
  certificate_arn   = module.dns.certificate_arn
  container_image   = "${var.ecr_repository_url}:${var.frontend_image_tag}"
  cluster_id        = aws_ecs_cluster.main.id
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn     = aws_iam_role.ecs_task_role.arn
  
  container_port     = 80
  health_check_path = "/"
}

module "backend" {
  source = "../../modules/backend"

  environment        = "staging"
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_ids = [module.security.backend_sg_id]
  certificate_arn   = module.dns.certificate_arn
  container_image   = "${var.ecr_repository_url}:${var.backend_image_tag}"
  cluster_id        = aws_ecs_cluster.main.id
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn     = aws_iam_role.ecs_task_role.arn
  
  container_port     = 3000
  health_check_path = "/health"
  
  depends_on = [
    module.database
  ]
}


module "certificates" {
  source = "./modules/dns"

  environment         = local.environment
  domain_name        = local.domain_name
  frontend_domain    = local.frontend_domain
  backend_domain     = local.backend_domain
  frontend_lb_dns_name = aws_lb.frontend[0].dns_name
  frontend_lb_zone_id  = aws_lb.frontend[0].zone_id
  backend_lb_dns_name  = aws_lb.backend[0].dns_name
  backend_lb_zone_id   = aws_lb.backend[0].zone_id
}