
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "octa-byte-ai"
      ManagedBy   = "Terraform"
    }
  }
}

module "infrastructure" {
  source = "../../"

  environment        = var.environment
  aws_region         = var.aws_region
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  db_instance_class  = var.db_instance_class
  app_name           = var.app_name
}
