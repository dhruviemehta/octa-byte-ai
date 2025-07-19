output "endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.main.endpoint
}

output "port" {
  description = "RDS instance port"
  value       = aws_db_instance.main.port
}

output "secret_arn" {
  description = "ARN of the secret containing DB credentials"
  value       = aws_secretsmanager_secret.db_password.arn
}
