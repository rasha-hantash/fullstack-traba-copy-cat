# backend.tf
terraform {
  backend "s3" {
    bucket         = "traba-terraform-states"  
    key            = "bootstrap/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-locks" 
  }
}