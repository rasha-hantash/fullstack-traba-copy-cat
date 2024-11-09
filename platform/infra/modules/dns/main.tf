# modules/certificates/main.tf

variable "environment" {
  description = "Environment name (e.g., staging, prod)"
  type        = string
}

variable "domain_name" {
  description = "Main domain name"
  type        = string
}

variable "frontend_domain" {
  description = "Frontend domain name"
  type        = string
}

variable "backend_domain" {
  description = "Backend domain name"
  type        = string
}

variable "frontend_lb_dns_name" {
  description = "Frontend load balancer DNS name"
  type        = string
}

variable "frontend_lb_zone_id" {
  description = "Frontend load balancer zone ID"
  type        = string
}

variable "backend_lb_dns_name" {
  description = "Backend load balancer DNS name"
  type        = string
}

variable "backend_lb_zone_id" {
  description = "Backend load balancer zone ID"
  type        = string
}

# ACM Certificate
resource "aws_acm_certificate" "main" {
  domain_name               = var.domain_name
  subject_alternative_names = ["*.${var.domain_name}"]
  validation_method         = "DNS"

  tags = {
    Environment = var.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Route53 Zone Data Source
data "aws_route53_zone" "main" {
  name         = var.domain_name
  private_zone = false
}

# Certificate Validation Records
resource "aws_route53_record" "acm_validation" {
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

# Certificate Validation
resource "aws_acm_certificate_validation" "main" {
  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [for record in aws_route53_record.acm_validation : record.fqdn]
}

# Frontend DNS Record
resource "aws_route53_record" "frontend" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = var.frontend_domain
  type    = "A"

  alias {
    name                   = var.frontend_lb_dns_name
    zone_id                = var.frontend_lb_zone_id
    evaluate_target_health = true
  }
}

# Backend DNS Record
resource "aws_route53_record" "backend" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = var.backend_domain
  type    = "A"

  alias {
    name                   = var.backend_lb_dns_name
    zone_id                = var.backend_lb_zone_id
    evaluate_target_health = true
  }
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "frontend" {
  name              = "/ecs/traba-${var.environment}-frontend"
  retention_in_days = 30

  tags = {
    Environment = var.environment
  }
}

resource "aws_cloudwatch_log_group" "backend" {
  name              = "/ecs/traba-${var.environment}-backend"
  retention_in_days = 30

  tags = {
    Environment = var.environment
  }
}

# Outputs
output "certificate_arn" {
  description = "ARN of the ACM certificate"
  value       = aws_acm_certificate.main.arn
}

output "domain_validation_options" {
  description = "Domain validation options for the certificate"
  value       = aws_acm_certificate.main.domain_validation_options 
}

output "frontend_log_group_name" {
  description = "Name of the frontend CloudWatch log group"
  value       =  aws_cloudwatch_log_group.frontend.name
}

output "backend_log_group_name" {
  description = "Name of the backend CloudWatch log group"
  value       =  aws_cloudwatch_log_group.backend.name 
}