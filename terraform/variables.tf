variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-west-2" # Or your preferred region
}

variable "project_name" {
  description = "Base name for resources"
  type        = string
  default     = "bp-tracker"
}

variable "vpc_cidr_block" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "List of CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"] # Example for 2 AZs
}

variable "private_subnet_cidrs" {
  description = "List of CIDR blocks for private subnets (for RDS and Lambda)"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24"] # Example for 2 AZs
}

variable "db_instance_class" {
  description = "Instance class for the RDS database"
  type        = string
  default     = "db.t3.micro" # Choose based on needs and cost
}

variable "db_allocated_storage" {
  description = "Allocated storage for RDS in GB"
  type        = number
  default     = 20
}

variable "db_engine" {
  description = "Database engine (e.g., postgres)"
  type        = string
  default     = "postgres"
}

variable "db_engine_version" {
  description = "Database engine version"
  type        = string
  default     = "15" # Match the version used locally if possible
}

variable "db_username" {
  description = "Master username for the RDS database"
  type        = string
  default     = "dbadmin"
}

variable "lambda_memory_size" {
  description = "Memory allocated to the Lambda function in MB"
  type        = number
  default     = 256 # Adjust based on application needs
}

variable "lambda_timeout" {
  description = "Timeout for the Lambda function in seconds"
  type        = number
  default     = 30 # API Gateway has a 29s timeout
}

variable "ecr_image_tag" {
  description = "Tag of the Docker image in ECR to deploy (e.g., 'latest' or a commit hash)"
  type        = string
  default     = "v0.0.4"
}

variable "tags" {
  description = "Common tags to apply to resources"
  type        = map(string)
  default = {
    Project = "bp-tracker"
    ManagedBy = "Terraform"
  }
}
