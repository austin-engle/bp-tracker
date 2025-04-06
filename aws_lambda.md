# Migrating bp-tracker to AWS Lambda + RDS (Serverless)

This document outlines the steps taken and the final configuration to run the `bp-tracker` application serverlessly on AWS Lambda, triggered by API Gateway, and using AWS RDS PostgreSQL for persistent data storage.

## Architecture Overview

1.  **API Gateway (HTTP API):** Acts as the HTTP frontend, receiving requests and triggering the Lambda function. Uses the cheaper HTTP API type with **payload format version 1.0** for compatibility with the Lambda Go proxy adapter.
2.  **AWS Lambda Function:** Contains the Go application logic (using Gin framework adapted via `aws-lambda-go-api-proxy/gin`), packaged as a container image based on the official `public.ecr.aws/lambda/go:1` base image. Executes only when requests arrive.
3.  **AWS RDS (PostgreSQL):** Provides the persistent relational database.
4.  **AWS RDS Proxy (Optional - Currently Bypassed):** While recommended for managing database connections, the current configuration connects Lambda directly to the RDS instance endpoint for troubleshooting simplicity. Reverting to use the proxy requires updating the `DB_HOST` environment variable in Terraform.
5.  **AWS Secrets Manager:** Securely stores the RDS database password, which is fetched directly by the Lambda function code using the AWS SDK.
6.  **AWS ECR:** Stores the Docker container image for the Lambda function.
7.  **Terraform:** Used to provision and manage all the AWS infrastructure components.

## Required Changes (Summary)

*   **Go Application Code (`internal/`, `cmd/`):
    *   Removed `http.ListenAndServe`.
    *   Implemented Lambda handler using `aws-lambda-go` and `aws-lambda-go-api-proxy/gin`.
    *   Switched database driver to `jackc/pgx/v5`.
    *   Modified database connection logic (`internal/database/db.go`) to read config from environment variables and fetch the password from Secrets Manager using the AWS SDK.
    *   Added `/migrate` route and handler (`internal/handlers/handlers.go`) to apply `schema.sql`.
*   **Configuration:** All configuration (DB host, port, user, name, SSL mode, secret ARN) is read from Lambda environment variables.
*   **`Dockerfile`:** Implemented multi-stage build using `golang:1.22-alpine` and final stage using `public.ecr.aws/lambda/go:1`, copying only the compiled `bootstrap` binary, static web assets, and `schema.sql`.
*   **Terraform (`terraform/main.tf`):** Defines VPC, security groups, ECR repository, RDS instance, Secrets Manager secret, RDS Proxy (optional), Lambda function, IAM roles/policies, and API Gateway.

## Lambda Configuration Details

*   **Runtime:** Container Image (`public.ecr.aws/lambda/go:1` base)
*   **Handler:** N/A (defined by container CMD)
*   **Environment Variables (Essential):
    *   `DB_HOST`: Endpoint of the RDS instance (e.g., `bp-tracker-db.xxxxxxxx.us-west-2.rds.amazonaws.com`) or RDS Proxy endpoint. *Ensure this does NOT include the port if connecting directly to RDS.*
    *   `DB_PORT`: `5432`
    *   `DB_USER`: Database username (e.g., `dbadmin`)
    *   `DB_NAME`: Database name (e.g., `bptrackerdb`)
    *   `DB_SSLMODE`: `require` (TLS is required for RDS/Proxy)
    *   `SECRET_ARN`: ARN of the secret in Secrets Manager containing the DB password.
*   **VPC:** Configured to run within the private subnets of the VPC.
*   **Security Group:** Associated with a security group (`lambda-sg`) allowing egress (needed for RDS/Proxy and Secrets Manager access).
*   **IAM Role:** Assigned an execution role (`lambda-exec-role`) with:
    *   `AWSLambdaBasicExecutionRole` (for CloudWatch Logs).
    *   `AWSLambdaVPCAccessExecutionRole` (for VPC access).
    *   Custom policy allowing `secretsmanager:GetSecretValue` on the specific `SECRET_ARN`.

## API Gateway Configuration Details

*   **Type:** HTTP API
*   **Integrations:** Uses AWS Lambda proxy integration (`AWS_PROXY`).
*   **Payload Format Version:** **`1.0`** (Version 2.0 caused routing issues with the `aws-lambda-go-api-proxy/gin` adapter).
*   **Routes:**
    *   `$default`: Catch-all route pointing to the Lambda integration (handles `/`, `/submit`, `/export/csv`, etc.).
    *   `POST /migrate`: Specific route pointing to the *same* Lambda integration but secured using **IAM Authorization**. This ensures only authenticated AWS principals can trigger the migration.

## Deployment Workflow

This section outlines the typical steps to deploy changes to the AWS environment after the initial Terraform setup.

1.  **Code Changes:** Make necessary changes to the Go application code (`internal/`, `cmd/`) or `Dockerfile`.
2.  **Build Docker Image:** Build the application's Docker image locally, ensuring the correct platform.
    ```bash
    docker build -t bp-tracker-lambda . --platform linux/amd64
    ```
3.  **Tag Docker Image:** Tag the image for ECR.
    ```bash
    # Replace <account-id>, <region>, <tag> (e.g., latest, commit hash)
    docker tag bp-tracker-lambda:latest <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<tag>
    ```
4.  **Push Docker Image to ECR:** Authenticate Docker with ECR and push the newly built image.
    ```bash
    # Authenticate (run once per session or when credentials expire)
    aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <account-id>.dkr.ecr.<region>.amazonaws.com

    # Push
    docker push <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<tag>
    ```
5.  **Update Lambda Function Code:** Update the Lambda function to use the new Docker image tag pushed to ECR.
    ```bash
    # Replace <account-id>, <region>, <new_tag>
    aws lambda update-function-code --function-name bp-tracker-app \
      --image-uri <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<new_tag> \
      --region <region>
    ```
    *(Alternatively, update the image tag variable in Terraform and run `terraform apply`)*

6.  **Apply Schema Migrations (if schema.sql changed or DB is new):**
    *   The application includes a `POST /migrate` endpoint designed to apply the schema defined in `internal/database/schema.sql`.
    *   This endpoint is secured with **IAM authorization** via API Gateway.
    *   To invoke it, use a tool capable of making AWS Signature Version 4 signed requests, such as `awscurl`.
    *   Ensure your local AWS credentials (used by `awscurl`) have the necessary `execute-api:Invoke` permission for the `/migrate` route on the deployed API Gateway stage.
    *   Run the following command (replace placeholders):
        ```bash
        # Replace <invoke_url> with your API Gateway stage URL (e.g., https://xxxx.execute-api.us-west-2.amazonaws.com)
        # Replace <region> with your AWS region
        awscurl --service execute-api -X POST "<invoke_url>/migrate" --region <region>
        ```
    *   A successful run should output `{"message":"Schema migration applied successfully!"}`. Check CloudWatch logs for the corresponding Lambda invocation for detailed success or error messages from the `MigrateHandler`.

## Troubleshooting Notes

*   **Routing Issues (`/migrate`):** If specific routes (like `/migrate`) aren't working, ensure the API Gateway integration's **Payload Format Version is set to `1.0`**. Version 2.0 caused issues with path parameters needed by the `ginadapter`.
*   **DNS Errors (`no such host`):** If connecting *directly* to the RDS instance (not proxy), the `DB_HOST` environment variable must contain *only* the hostname, not the port. The RDS instance `.endpoint` attribute in Terraform sometimes includes the port. Use `split(":", aws_db_instance.main.endpoint)[0]` in Terraform if needed, or manually verify/correct the `DB_HOST` value in the Lambda console.
*   **TLS Errors (`tls error: EOF`):** When connecting to RDS or RDS Proxy, ensure `DB_SSLMODE=require` is set in the Lambda environment variables. Also verify security groups allow traffic on port 5432 between the Lambda function and the target (RDS instance or Proxy).
*   **Migration Errors (`relation "readings" does not exist`):** Indicates the schema migration via `POST /migrate` did not run or failed. Check the CloudWatch logs for the specific `/migrate` invocation for errors (e.g., `MIGRATION_ERROR: Error executing schema...`). Ensure the `/migrate` request is actually reaching the `MigrateHandler` and not being misrouted to `HomeHandler`.
*   **CloudWatch Logs:** Always check the CloudWatch log group for your Lambda function (`/aws/lambda/bp-tracker-app`) for detailed error messages and request/response logs.
