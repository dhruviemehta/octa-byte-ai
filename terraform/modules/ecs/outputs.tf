output "cluster_name" {
  description = "Name of ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "service_name" {
  description = "Name of ECS service"
  value       = aws_ecs_service.app.name
}
