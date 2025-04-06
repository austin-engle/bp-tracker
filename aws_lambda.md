# Migrating bp-tracker to AWS Lambda + RDS (Serverless)

This document outlines the necessary changes to run the `bp-tracker` application serverlessly on AWS Lambda, triggered by API Gateway, and using AWS RDS for persistent data storage. This approach is optimized for infrequent usage (e.g., 1-2 times per day) to minimize costs.

## Architecture Overview

1.  **API Gateway (HTTP API):** Acts as the HTTP frontend, receiving requests and triggering the Lambda function. Use the cheaper HTTP API type.
2.  **AWS Lambda Function:** Contains the Go application logic, packaged as a `.zip` file or container image. Executes only when requests arrive.
3.  **AWS RDS:** Provides the persistent relational database (e.g., PostgreSQL or MySQL). A small instance type (`t4g.micro` or `t3.micro`) is recommended.
4.  **AWS RDS Proxy (Recommended):** Manages database connections efficiently between Lambda and RDS, mitigating connection overhead and improving performance.
5.  **AWS Secrets Manager:** Securely stores the RDS database password.

## Required Changes

### 1. Go Application Code (`internal/`, `cmd/`)

*   **Remove HTTP Server:** Eliminate the `http.ListenAndServe` logic. The application will no longer run as a persistent server.
*   **Implement Lambda Handler:**
    *   Use the `aws-lambda-go` library.
    *   Create a main handler function (e.g., `HandleRequest`) that accepts an API Gateway event payload.
    *   Adapt existing web framework code (if any) to work within the handler, possibly using helper libraries like `aws-lambda-go-api-proxy` to translate events.
*   **Database Logic:**
    *   Replace the SQLite driver (`mattn/go-sqlite3`) with an appropriate RDS driver (e.g., `jackc/pgx` for PostgreSQL or `go-sql-driver/mysql` for MySQL). Update `go.mod`.
    *   Modify database connection logic to:
        *   Read connection details (host, port, user, dbname, password) from environment variables.
        *   Connect to the RDS endpoint (ideally the RDS Proxy endpoint).
        *   Handle connection setup/teardown within the Lambda invocation OR rely on RDS Proxy's pooling.
*   **Request/Response:** Adjust code to read request details from the Lambda event object and return responses in the format required by API Gateway.

### 2. Configuration

*   **Environment Variables:** Ensure *all* configuration (DB connection strings, ports, etc.) is read from environment variables provided to the Lambda function.
*   **Secrets Management:** Store the database password in AWS Secrets Manager. Grant the Lambda function's IAM role permission to read it.

### 3. `Dockerfile`

*   **Purpose:** Primarily for building the Lambda deployment package (if using container image deployment) and for local testing via `docker-compose`.
*   **Modifications:**
    *   Ensure `bp.db` is NOT copied into the image.
    *   If deploying Lambda via container image, the `CMD`/`ENTRYPOINT` needs to be compatible with the Lambda Runtime Interface Emulator (RIE). Often, you just build the binary and Lambda invokes it directly. Consult AWS Lambda container image documentation.
    *   Ensure multi-stage builds are clean and produce a minimal final image containing only the compiled Go binary and necessary assets (like static web files if any).

### 4. `.dockerignore`

*   Add or ensure `bp.db` is present to exclude it from the Docker build context.
*   Also exclude `.git`, `.env`, `aws_lambda.md`, etc.

### 5. `docker-compose.yml` (Local Development)

*   **Replace `bp.db`:** Remove any references or volumes related to the local `bp.db` file.
*   **Add Database Service:** Define a new service using a standard database image (e.g., `postgres:alpine` or `mysql:latest`).
    *   Configure its environment variables (e.g., `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`).
    *   Use a Docker volume to persist data locally between `docker-compose up/down` cycles.
*   **Update `app` Service:**
    *   Modify the `environment` section to provide `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` pointing to the database service defined above (e.g., `DB_HOST: db`).
    *   Ensure `depends_on` specifies the database service.

## Implementation Details for AI Assistant

This section provides specific guidance on the code modifications:

*   **Main Function (`cmd/.../main.go`):**
    *   Remove any calls to `http.ListenAndServe` or similar web server startup functions.
    *   The `main` function should now typically contain only `lambda.Start(YourHandlerFunction)`.
    *   `YourHandlerFunction` needs to match a signature compatible with `aws-lambda-go`, often `func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)`.
*   **Dependency Changes (`go.mod`):**
    *   Remove `github.com/mattn/go-sqlite3`.
    *   Add `github.com/aws/aws-lambda-go`.
    *   Add the appropriate RDS database driver (e.g., `github.com/jackc/pgx/v4` for PostgreSQL or `github.com/go-sql-driver/mysql` for MySQL).
    *   Run `go mod tidy`.
*   **Database Connection:**
    *   Locate the code responsible for opening the SQLite database connection (`sql.Open("sqlite3", ...)`).
    *   Replace it with code that opens a connection to PostgreSQL or MySQL using the new driver.
    *   Construct the Database Source Name (DSN) string dynamically using environment variables fetched via `os.Getenv("VAR_NAME")`. Example DSNs:
        *   PostgreSQL: `"postgresql://user:password@host:port/dbname"`
        *   MySQL: `"user:password@tcp(host:port)/dbname"`
    *   Ensure environment variables like `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` are read.
*   **Web Framework Adapters (If Applicable):**
    *   If using Gin, Echo, etc., find where the router/engine is created and routes are defined.
    *   Instead of calling `router.Run()`, use an adapter library (like `github.com/awslabs/aws-lambda-go-api-proxy/gin` or `/echo`) to create the Lambda handler function. The adapter will translate API Gateway events into standard `http.Request` objects for the framework.
    *   Example (Gin):
        ```go
        import (
            "github.com/aws/aws-lambda-go/lambda"
            ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
            "github.com/gin-gonic/gin"
        )

        var ginLambda *ginadapter.GinLambda

        func init() {
            // Setup Gin router (define routes, middleware)
            router := gin.Default()
            // ... register routes ...
            ginLambda = ginadapter.New(router)
        }

        func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
            // If no REST server is needed, attach one to the GinLambda
            return ginLambda.ProxyWithContext(ctx, req)
        }

        func main() {
            lambda.Start(Handler)
        }
        ```
*   **`Dockerfile`:** Modify the final stage to just copy the compiled Go binary. The `CMD` should typically just be the path to the binary (e.g., `CMD ["/main"]`). Lambda's runtime environment handles invoking it. Remove any `EXPOSE` instructions related to the old web server port.
*   **`docker-compose.yml`:** Ensure the `app` service's `environment` variables match those needed for the RDS connection string (e.g., `DB_HOST: db`, `DB_USER: localuser`, etc.) and that a `db` service (using `postgres` or `mysql` image) is defined with corresponding credentials.

## AWS Setup Summary

1.  Create RDS instance and configure security groups.
2.  (Recommended) Set up RDS Proxy pointing to the RDS instance.
3.  Create Lambda function (Go runtime or container image), configure environment variables (using Secrets Manager for password), VPC access, and IAM role.
4.  Create API Gateway (HTTP API), define routes, and integrate them with the Lambda function.
5.  Deploy the API Gateway.

## Deployment Workflow

This section outlines the typical steps to deploy changes to the AWS environment after the initial Terraform setup.

1.  **Code Changes:** Make necessary changes to the Go application code (`internal/`, `cmd/`).
2.  **Build Docker Image:** Build the application's Docker image locally using the updated `Dockerfile`.
    ```bash
    # Example: Replace <account-id>, <region>, <tag> with your values
    docker build -t <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<tag> .
    ```
    *   Common tags (`<tag>`) include `latest`, a commit hash (`git rev-parse --short HEAD`), or a version number.
3.  **Push Docker Image to ECR:** Authenticate Docker with ECR and push the newly built image.
    ```bash
    # Authenticate
    aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <account-id>.dkr.ecr.<region>.amazonaws.com

    # Push
    docker push <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<tag>
    ```
4.  **Apply Schema Migrations (if needed):**
    *   If you have made changes to `internal/database/schema.sql` (or if using a migration tool with new migration files), run the migration script against the **deployed RDS database**.
    *   You will need database credentials (host, port, user, password, dbname) for the deployed RDS instance. The host will be the **RDS Instance Endpoint** (or **RDS Proxy Endpoint** if connecting through the proxy, which is recommended). The password should be retrieved from **AWS Secrets Manager**.
    *   **Set Environment Variables:** Before running the script, set the required `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` environment variables in your terminal session where you will run the script. Use `DB_SSLMODE=require` when connecting to RDS/RDS Proxy.
        ```bash
        # Example (replace with actual values, retrieving password securely)
        export DB_HOST="<rds-proxy-or-instance-endpoint>"
        export DB_PORT="5432"
        export DB_USER="dbadmin" # Or your RDS username
        export DB_PASSWORD="$(aws secretsmanager get-secret-value --secret-id bp-tracker/db-password --query SecretString --output text)" # Example: retrieve from Secrets Manager
        export DB_NAME="<rds-db-name>" # e.g., postgres or bp_tracker_db
        export DB_SSLMODE="require"
        ```
    *   **Run the Migration Script:** Execute the Go script from the project root.
        ```bash
        go run scripts/migrate_schema.go
        ```
    *   **Note:** Running this script locally requires network connectivity from your machine to the RDS instance/proxy within the VPC (e.g., via VPN, bastion host, or temporarily adjusting security groups - use caution).
    *   **Alternative:** Integrate this script execution into a CI/CD pipeline step that runs within your AWS environment (e.g., using AWS CodeBuild) for more secure and automated database access.
5.  **Update Lambda Function:** Update the Lambda function to use the new Docker image tag pushed to ECR.
    *   **Terraform:** Update the `ecr_image_tag` variable in your Terraform configuration (e.g., in `terraform.tfvars` or via `-var="ecr_image_tag=<new_tag>"`) and run `terraform apply`.
    *   **AWS CLI:** Use the `aws lambda update-function-code` command:
        ```bash
        aws lambda update-function-code --function-name bp-tracker-app --image-uri <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<new_tag>
        ```
    *   **AWS Console:** Manually update the image URI in the Lambda function's configuration.

This setup provides a serverless execution model suitable for infrequent requests, minimizing costs while retaining data persistence through RDS.
