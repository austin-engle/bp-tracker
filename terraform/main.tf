terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source = "hashicorp/random"
      version = "~> 3.1"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.0"
    }
  }
}

locals {
  db_name = "bptrackerdb"
}

provider "aws" {
  region = var.aws_region
}

# Data source to get the public IP of the machine running Terraform
data "http" "my_ip" {
  url = "http://ifconfig.me/ip"
}

# Random string for DB password if not provided explicitly
resource "random_password" "db_password" {
  length           = 16
  special          = false
}

# --- Networking --- VPC, Subnets, Security Groups ---
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.project_name}-vpc"
  cidr = var.vpc_cidr_block

  azs             = slice(data.aws_availability_zones.available.names, 0, 2) # Use 2 AZs
  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs

  enable_nat_gateway   = true # Needed for Lambda in private subnet to reach AWS APIs (e.g., Secrets Manager)
  single_nat_gateway   = true # Cheaper for low usage
  enable_dns_hostnames = true

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }

  tags = var.tags
}

data "aws_availability_zones" "available" {}

# Security group for Lambda
resource "aws_security_group" "lambda" {
  name        = "${var.project_name}-lambda-sg"
  description = "Allow outbound traffic from Lambda, allow inbound from API GW is handled by Lambda permissions"
  vpc_id      = module.vpc.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1" # Allow all outbound
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.project_name}-lambda-sg"
  })
}

# Security group for RDS/Proxy
resource "aws_security_group" "db" {
  name        = "${var.project_name}-db-sg"
  description = "Allow traffic from Lambda and my IP to RDS/Proxy"
  vpc_id      = module.vpc.vpc_id

  # Allow inbound PostgreSQL traffic from Lambda SG
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.lambda.id]
    description     = "Allow PostgreSQL from Lambda"
  }

  # Allow inbound PostgreSQL traffic from my current IP (at terraform apply time)
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.my_ip.response_body)}/32"]
    description = "Allow PostgreSQL from my IP (Terraform apply time)"
  }

  # Allow outbound traffic (optional, but often needed for patches etc.)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.project_name}-db-sg"
  })
}

# --- ECR --- Repository for the Docker Image ---
resource "aws_ecr_repository" "app" {
  name                 = var.project_name
  image_tag_mutability = "MUTABLE" # Or IMMUTABLE for stricter versioning

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = var.tags
}

# --- Secrets Manager --- For RDS Password ---
resource "aws_secretsmanager_secret" "db_password" {
  name        = "${var.project_name}/db-password"
  description = "Master password for the RDS database"
  recovery_window_in_days = 0 # Set to 0 to force delete without waiting period (Use with caution!)

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = random_password.db_password.result
}

# --- RDS --- PostgreSQL Database Instance ---
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = module.vpc.private_subnets # Place RDS in private subnets

  tags = var.tags
}

resource "aws_db_instance" "main" {
  identifier           = "${var.project_name}-db"
  allocated_storage    = var.db_allocated_storage
  engine               = var.db_engine
  engine_version       = var.db_engine_version
  instance_class       = var.db_instance_class
  username             = var.db_username
  password             = random_password.db_password.result # Reference the random password
  # name                 = "${var.project_name}_db" # Initial DB name (optional, can be created by app)
  db_subnet_group_name = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.db.id]
  skip_final_snapshot  = true # Set to false for production
  publicly_accessible  = false # Keep DB private
  db_name              = local.db_name

  tags = var.tags
}

# --- RDS Proxy --- Recommended for Lambda ---
# IAM role for RDS Proxy to access secrets
resource "aws_iam_role" "rds_proxy_secret_role" {
  name = "${var.project_name}-rds-proxy-secret-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "rds.amazonaws.com"
        }
      }
    ]
  })
  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "rds_proxy_secret_attachment" {
  role       = aws_iam_role.rds_proxy_secret_role.name
  policy_arn = "arn:aws:iam::aws:policy/SecretsManagerReadWrite" # Might need refinement to least privilege
}

# resource "aws_db_proxy" "main" {
#   name                   = "${var.project_name}-proxy"
#   debug_logging          = false
#   engine_family          = "POSTGRESQL"
#   idle_client_timeout    = 1800
#   require_tls            = true # Enforce TLS
#   role_arn               = aws_iam_role.rds_proxy_secret_role.arn
#   vpc_security_group_ids = [aws_security_group.db.id]
#   vpc_subnet_ids         = module.vpc.private_subnets

#   auth {
#     auth_scheme = "SECRETS"
#     description = "Credentials for DB"
#     iam_auth    = "DISABLED"
#     secret_arn  = aws_secretsmanager_secret.db_password.arn
#   }

#   tags = var.tags

#   # Depends on the DB instance being available
#   depends_on = [aws_db_instance.main]
# }

# --- IAM --- Role and Policies for Lambda ---
resource "aws_iam_role" "lambda_exec_role" {
  name = "${var.project_name}-lambda-exec-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
  tags = var.tags
}

# Basic execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Policy for VPC access
resource "aws_iam_role_policy_attachment" "lambda_vpc_access" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# Policy to allow Lambda to read the DB password secret
resource "aws_iam_policy" "lambda_read_db_secret" {
  name        = "${var.project_name}-lambda-read-db-secret-policy"
  description = "Allow Lambda to read the DB password secret"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "secretsmanager:GetSecretValue"
        ],
        Effect   = "Allow",
        Resource = aws_secretsmanager_secret.db_password.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_read_db_secret" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = aws_iam_policy.lambda_read_db_secret.arn
}

# --- Lambda --- Application Function ---
resource "aws_lambda_function" "app" {
  function_name = "${var.project_name}-app"
  role          = aws_iam_role.lambda_exec_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.app.repository_url}:${var.ecr_image_tag}"
  timeout       = var.lambda_timeout
  memory_size   = var.lambda_memory_size

  # Environment variables for the application
  environment {
    variables = {
      # TEMPORARY CHANGE: Point directly to RDS instance endpoint for testing
      # DB_HOST     = aws_db_proxy.main.endpoint # Connect via Proxy
      DB_HOST     = split(":", aws_db_instance.main.endpoint)[0]
      DB_PORT     = "5432"
      DB_USER     = var.db_username
      # DB_PASSWORD is read from Secrets Manager by the app code
      DB_NAME     = local.db_name # Or the default DB name if not specified in rds resource
      DB_SSLMODE  = "require" # Still require TLS for direct connection
      SECRET_ARN  = aws_secretsmanager_secret.db_password.arn # Pass secret ARN if app reads password itself
      # Add other env vars like GIN_MODE=release if needed
    }
  }

  # VPC configuration to access RDS Proxy (or DB directly in this case)
  vpc_config {
    subnet_ids         = module.vpc.private_subnets
    security_group_ids = [aws_security_group.lambda.id]
  }

  # TEMPORARY CHANGE: Depends directly on the DB instance now, not the proxy
  depends_on = [aws_db_instance.main, aws_ecr_repository.app]
  # depends_on = [aws_db_proxy.main, aws_ecr_repository.app]

  tags = var.tags
}

# --- API Gateway --- HTTP API Frontend ---
resource "aws_apigatewayv2_api" "http_api" {
  name          = "${var.project_name}-http-api"
  protocol_type = "HTTP"
  description   = "HTTP API for ${var.project_name}"

  tags = var.tags
}

# Lambda integration
resource "aws_apigatewayv2_integration" "lambda" {
  api_id             = aws_apigatewayv2_api.http_api.id
  integration_type   = "AWS_PROXY" # Use Lambda proxy integration
  integration_method = "POST" # Always POST for AWS_PROXY
  integration_uri    = aws_lambda_function.app.invoke_arn
  payload_format_version = "1.0" # Changed to v1.0
}

# Default route to catch all requests (for application access)
resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "$default" # Catch-all route key
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# --- NEW: Migration Route with IAM Authorization ---
resource "aws_apigatewayv2_route" "migrate" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "POST /migrate" # Specific route for POST /migrate
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"

  # IMPORTANT: Secure this route with IAM authorization
  authorization_type = "AWS_IAM"
}

# Deployment stage
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default" # Default stage
  auto_deploy = true

  tags = var.tags
}

# Permission for API Gateway to invoke Lambda for default stage
# (API Gateway handles permissions automatically for integrated routes like this)
resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal     = "apigateway.amazonaws.com"

  # Restrict to the specific API Gateway ARN execution context
  source_arn = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}
