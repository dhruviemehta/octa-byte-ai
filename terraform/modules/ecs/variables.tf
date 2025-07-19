variable "environment" {
  description = "Environment name"
  type        = string
}

variable "app_name" {
  description = "Application name"
  type        = string
}

variable "app_port" {
  description = "Application port"
  type        = number
}

variable "private_subnet_ids" {
  description = "Private subnet IDs"
  type        = list(string)
}

variable "app_security_group_id" {
  description = "App security group ID"
  type        = string
}

variable "target_group_arn" {
  description = "Target group ARN"
  type        = string
}

variable "db_endpoint" {
  description = "Database endpoint"
  type        = string
}

variable "db_secret_arn" {
  description = "Database secret ARN from RDS module"
  type        = string
}

variable "ecr_repository_url" {
  description = "ECR repository URL from ECR module"
  type        = string
}
