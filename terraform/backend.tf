# terraform {
#   required_version = ">= 1.5.0"
#   required_providers {
#     aws = {
#       source  = "hashicorp/aws"
#       version = "~> 5.0"
#     }
#   }
#   backend "s3" {
#     bucket         = "octa-byte-terraform-state-staging"
#     key            = "environments/staging/terraform.tfstate"
#     region         = "ap-south-1" # Changed from us-west-2 to ap-south-1
#     dynamodb_table = "terraform-state-lock"
#     encrypt        = true
#   }
# }
