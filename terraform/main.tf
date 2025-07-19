# VPC Module
module "vpc" {
  source = "./modules/vpc"

  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}

# Security Module
module "security" {
  source = "./modules/security"

  environment = var.environment
  vpc_id      = module.vpc.vpc_id
  app_port    = var.app_port
}

# RDS Module
module "rds" {
  source = "./modules/rds"

  environment          = var.environment
  db_instance_class    = var.db_instance_class
  private_subnet_ids   = module.vpc.private_subnet_ids
  db_security_group_id = module.security.db_security_group_id
}

# ALB Module
module "alb" {
  source = "./modules/alb"

  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  public_subnet_ids     = module.vpc.public_subnet_ids
  alb_security_group_id = module.security.alb_security_group_id
}

# ECS Module
module "ecs" {
  source = "./modules/ecs"

  environment           = var.environment
  app_name              = var.app_name
  app_port              = var.app_port
  private_subnet_ids    = module.vpc.private_subnet_ids
  app_security_group_id = module.security.app_security_group_id
  target_group_arn      = module.alb.target_group_arn
  db_endpoint           = module.rds.endpoint
  db_secret_arn         = module.rds.secret_arn # Pass secret ARN
  ecr_repository_url    = module.ecr.repository_url


}

module "ecr" {
  source = "./modules/ecr"

  environment = var.environment
  app_name    = var.app_name

  tags = {
    Module = "ecr"
  }
}
