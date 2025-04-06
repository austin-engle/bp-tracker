output "api_gateway_endpoint" {
  description = "The invocation URL for the API Gateway endpoint"
  value       = aws_apigatewayv2_stage.default.invoke_url
}

output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.app.repository_url
}

output "rds_instance_endpoint" {
  description = "Endpoint address of the RDS instance"
  value       = aws_db_instance.main.endpoint
}

output "db_proxy_endpoint" {
  description = "Endpoint address of the DB Proxy"
  value       = aws_db_proxy.main.endpoint
}

output "lambda_function_name" {
  description = "Name of the deployed Lambda function"
  value       = aws_lambda_function.app.function_name
}

output "db_password_secret_arn" {
  description = "ARN of the Secrets Manager secret storing the DB password"
  value       = aws_secretsmanager_secret.db_password.arn
}
