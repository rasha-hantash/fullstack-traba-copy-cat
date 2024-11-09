# modules/bastion/variables.tf
variable "environment" {
  description = "Environment name (e.g., staging, prod)"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "subnet_id" {
  description = "ID of the public subnet for the bastion host"
  type        = string
}

variable "allowed_cidrs" {
  description = "List of CIDR blocks allowed to connect to bastion"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "instance_type" {
  description = "EC2 instance type for bastion host"
  type        = string
  default     = "t3.micro"
}

variable "key_name" {
  description = "Name of the SSH key pair"
  type        = string
  default     = "bastion-key-pair-2"
}

variable  "vpc_security_group_ids" {
  description = "List of security group IDs for the bastion host"
  type        = list(string)
}

# modules/bastion/main.tf
data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

# resource "aws_security_group" "bastion" {
#   name_prefix = "bastion-${var.environment}-"
#   description = "Security group for bastion host"
#   vpc_id      = var.vpc_id

#   ingress {
#     description = "SSH from allowed IPs"
#     from_port   = 22
#     to_port     = 22
#     protocol    = "tcp"
#     cidr_blocks = var.allowed_cidrs
#   }

#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   tags = {
#     Name        = "bastion-${var.environment}"
#     Environment = var.environment
#   }

#   lifecycle {
#     create_before_destroy = true
#   }
# }

resource "aws_iam_role" "bastion" {
  name_prefix = "bastion-${var.environment}-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Environment = var.environment
  }
}

resource "aws_iam_instance_profile" "bastion" {
  name_prefix = "bastion-${var.environment}-"
  role        = aws_iam_role.bastion.name
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = aws_iam_role.bastion.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_instance" "bastion" {
  ami           = data.aws_ami.amazon_linux_2.id
  instance_type = var.instance_type
  subnet_id     = var.subnet_id
  key_name      = var.key_name

  vpc_security_group_ids = var.vpc_security_group_ids  # Use the variable here
  iam_instance_profile        = aws_iam_instance_profile.bastion.name
  associate_public_ip_address = true

  root_block_device {
    volume_size = 8
    encrypted   = true
  }

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required" # IMDSv2
  }

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y postgresql15
              EOF

  tags = {
    Name        = "bastion-${var.environment}"
    Environment = var.environment
  }
}

# modules/bastion/outputs.tf
output "bastion_public_ip" {
  description = "Public IP of bastion host"
  value       = aws_instance.bastion.public_ip
}

output "bastion_public_dns" {
  description = "Public DNS of bastion host"
  value       = aws_instance.bastion.public_dns
}

output "instance_id" {
  description = "ID of the bastion instance"
  value       = aws_instance.bastion.id
}