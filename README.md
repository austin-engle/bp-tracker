# Blood Pressure Tracker

A simple web application to track blood pressure readings.

This application is designed to be deployed as a serverless function on AWS Lambda using API Gateway, ECR, RDS PostgreSQL, and Secrets Manager.

## Features

*   Enter systolic, diastolic, and pulse readings.
*   Calculates average readings.
*   Classifies blood pressure based on American Heart Association guidelines (2017).
*   Provides recommendations based on classification.
*   Displays statistics: last reading, 7-day average, 30-day average, all-time average.
*   Export all readings to CSV.
*   (AWS) Schema migration endpoint (`/migrate`).

## Technologies Used

*   **Backend:** Go (Golang)
*   **Web Framework:** Gin
*   **Database:** PostgreSQL (AWS RDS)
*   **Deployment:** AWS Lambda, API Gateway, ECR, Secrets Manager
*   **Infrastructure:** Terraform
*   **Containerization:** Docker

## Setup & Deployment (AWS Lambda)

This application is intended for deployment on AWS Lambda.

1.  **Prerequisites:**
    *   AWS Account
    *   Terraform installed
    *   Docker installed
    *   AWS CLI configured with appropriate credentials
    *   `awscurl` installed (`pip install awscurl`) for migration

2.  **Infrastructure Setup (Terraform):**
    *   Navigate to the `terraform/` directory.
    *   Create a `terraform.tfvars` file or set variables via environment/CLI. Essential variables include `aws_region`, `project_name`, `db_username`.
    *   Run `terraform init`.
    *   Run `terraform apply` to provision the VPC, RDS instance, ECR repository, Lambda function, API Gateway, and other resources. Note the API Gateway invoke URL from the output.

3.  **Build & Push Docker Image:**
    *   Navigate back to the project root.
    *   Build the image: `docker build -t bp-tracker-lambda . --platform linux/amd64`
    *   Tag the image: `docker tag bp-tracker-lambda:latest <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<tag>` (replace placeholders, `<tag>` is often `latest`)
    *   Log in to ECR: `aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <account-id>.dkr.ecr.<region>.amazonaws.com`
    *   Push the image: `docker push <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<tag>`

4.  **Update Lambda Function:**
    *   Update the Lambda function to use the new image:
        ```bash
        aws lambda update-function-code --function-name bp-tracker-app \
          --image-uri <account-id>.dkr.ecr.<region>.amazonaws.com/bp-tracker:<new_tag> \
          --region <region>
        ```

5.  **Database Migration:**
    *   Apply the database schema using the `/migrate` endpoint:
        ```bash
        # Replace <invoke_url> and <region>
        awscurl --service execute-api -X POST "<invoke_url>/migrate" --region <region>
        ```
    *   Ensure your AWS credentials have `execute-api:Invoke` permission for this route.

6.  **Access Application:** Open the API Gateway Invoke URL in your browser.

*For detailed steps and troubleshooting, see `aws_lambda.md`.*

## Local Development

While the primary target is AWS Lambda, local development using Docker Compose is possible for testing UI changes or basic functionality, but it uses a separate PostgreSQL container, **not** the RDS database.

*See `LOCAL_DEVELOPMENT.md` for instructions.*

## Project Structure

```
.
├── Dockerfile # Builds the Lambda container image
├── Dockerfile.local # (Example) Dockerfile for local dev if needed
├── Makefile # (Optional) For build/test commands
├── README.md # This file
├── aws_lambda.md # AWS deployment details
├── cmd/
│ └── server/
│ └── main.go # Main application entrypoint (Lambda handler)
├── docker-compose.yml # For local development environment
├── go.mod # Go module dependencies
├── go.sum
├── internal/ # Internal application code
│ ├── database/
│ │ ├── db.go # Database connection logic (RDS/Secrets Manager)
│ │ └── schema.sql # Database schema
│ ├── handlers/
│ │ └── handlers.go # HTTP request handlers (Gin)
│ ├── models/
│ │ └── models.go # Data structures
│ └── utils/
│ └── utils.go # Helper functions (e.g., BP classification)
├── scripts/ # Utility scripts (if any)
├── terraform/ # Terraform infrastructure code
│ ├── main.tf
│ ├── variables.tf
│ └── outputs.tf
└── web/ # Frontend files
├── static/
│ ├── css/
│ └── js/
└── templates/
└── index.html # Main HTML template
```


---

**2. Updated `quick_start.md`**

Please replace the content of `bp-tracker/quick_start.md` with this:

```markdown
# Quick Start (AWS Lambda Deployment)

This guide provides the minimal steps to deploy and run the Blood Pressure Tracker application on AWS Lambda. For detailed explanations, prerequisites, and troubleshooting, refer to `README.md` and `aws_lambda.md`.

**Assumptions:**

*   You have an AWS account.
*   Terraform, Docker, AWS CLI, and `awscurl` are installed and configured.

**Steps:**

1.  **Clone Repository:**
    ```bash
    git clone <repository_url>
    cd bp-tracker
    ```

2.  **Provision Infrastructure (Terraform):**
    *   Navigate to `terraform/`.
    *   Configure variables (e.g., via `terraform.tfvars` or command line), ensuring `aws_region` is set.
    *   Initialize Terraform:
        ```bash
        terraform init
        ```
    *   Apply Terraform configuration:
        ```bash
        terraform apply
        ```
    *   Note the `api_gateway_invoke_url` output value.

3.  **Build & Push Application Image:**
    *   Navigate back to the project root (`cd ..`).
    *   Build the Docker image:
        ```bash
        docker build -t bp-tracker-lambda . --platform linux/amd64
        ```
    *   Tag the image (replace placeholders):
        ```bash
        # Example tag: latest
        TAG=latest
        ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
        REGION=<your_aws_region> # e.g., us-west-2
        REPO_URI="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/bp-tracker"

        docker tag bp-tracker-lambda:latest "${REPO_URI}:${TAG}"
        ```
    *   Authenticate Docker with ECR:
        ```bash
        aws ecr get-login-password --region ${REGION} | docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com
        ```
    *   Push the image:
        ```bash
        docker push "${REPO_URI}:${TAG}"
        ```

4.  **Update Lambda Function:**
    ```bash
    aws lambda update-function-code --function-name bp-tracker-app \
      --image-uri "${REPO_URI}:${TAG}" \
      --region ${REGION}
    ```
    *(Wait a few moments for the update to propagate)*

5.  **Run Database Migration:**
    *   Use the API Gateway URL noted from the Terraform output.
    *   Ensure your AWS credentials allow `execute-api:Invoke`.
    ```bash
    INVOKE_URL=<api_gateway_invoke_url_from_terraform>

    awscurl --service execute-api -X POST "${INVOKE_URL}/migrate" --region ${REGION}
    ```
    *   Expected output: `{"message":"Schema migration applied successfully!"}`

6.  **Access Application:**
    *   Open the `INVOKE_URL` in your web browser.

You should now see the Blood Pressure Tracker application interface running on AWS Lambda.
```

---

**3. Updated `LOCAL_DEVELOPMENT.md`**

Please replace the content of `bp-tracker/LOCAL_DEVELOPMENT.md` with this:

```markdown
# Local Development Setup (Docker Compose)

This document describes how to run the Blood Pressure Tracker application locally using Docker Compose.

**Important Considerations:**

*   The local environment uses a **PostgreSQL container** managed by Docker Compose. It does **not** connect to the AWS RDS database used in the deployed AWS Lambda environment.
*   The local setup simulates the application logic but **does not perfectly replicate the AWS Lambda execution environment** or API Gateway behavior. Features specific to Lambda (like context handling or integration triggers) might behave differently.
*   Secrets (like the database password) are handled via environment variables directly in `docker-compose.yml` for local simplicity, unlike the AWS environment which uses Secrets Manager.
*   Database migrations need to be applied manually to the local PostgreSQL container.

## Prerequisites

*   Docker installed
*   Docker Compose installed (often included with Docker Desktop)

## Setup and Running

1.  **Clone the Repository:** If you haven't already:
    ```bash
    git clone <repository_url>
    cd bp-tracker
    ```

2.  **Review `docker-compose.yml`:**
    *   This file defines two main services:
        *   `db`: Runs a PostgreSQL container. It sets up the user (`localuser`), password (`localpassword`), and database name (`bptrackerdb`). The password here is **only** for the local container. Data is persisted in a Docker volume (`postgres_data`).
        *   `app`: Builds and runs the Go application using `Dockerfile`. **Crucially**, it passes environment variables (`DB_HOST=db`, `DB_USER=localuser`, `DB_PASSWORD=localpassword`, `DB_NAME=bptrackerdb`, `DB_SSLMODE=disable`, `SECRET_ARN=local`) to the Go application so it connects to the `db` service container. `SECRET_ARN` is set to `local` as a placeholder since Secrets Manager isn't used here. Note `DB_SSLMODE` is `disable` locally.

3.  **Build and Start Services:**
    ```bash
    docker-compose up --build
    ```
    *   `--build` ensures the Go application image is rebuilt if code changes.
    *   This will download the PostgreSQL image (if not present), build the `app` image, and start both containers.

4.  **Apply Database Schema (First Time):**
    *   The first time you run `docker-compose up`, the `bptrackerdb` database inside the `db` container will be empty. You need to apply the schema.
    *   Open a **new terminal window**.
    *   Execute the schema file against the running `db` container using `psql`:
        ```bash
        # Command explained:
        # docker-compose exec db : Execute command inside the 'db' service container
        # psql : The PostgreSQL client
        # -U localuser : Connect as user 'localuser'
        # -d bptrackerdb : Connect to database 'bptrackerdb'
        # -f /docker-entrypoint-initdb.d/schema.sql : Execute the schema file (copied via volume mount)

        docker-compose exec db psql -U localuser -d bptrackerdb -f /docker-entrypoint-initdb.d/schema.sql
        ```
    *   You should see output like `CREATE TABLE`, indicating success. If you get errors, check the `docker-compose logs db` output in the original terminal.

5.  **Access Application:**
    *   Open your web browser and navigate to `http://localhost:8080` (or the port mapped in `docker-compose.yml`).

## Stopping the Environment

*   Press `Ctrl+C` in the terminal where `docker-compose up` is running.
*   To remove the containers (but keep the database volume): `docker-compose down`
*   To remove containers AND the database volume (lose all local data): `docker-compose down -v`

## Code Changes

*   If you change Go code (`internal/`, `cmd/`), restart the services: `docker-compose down && docker-compose up --build`.
*   If you change frontend code (`web/`), you might only need to restart the `app` service (`docker-compose restart app`) or potentially rebuild (`docker-compose up --build -d app`).

This local setup is useful for iterating on the Go application logic and frontend templates before deploying to the full AWS environment. Remember to test thoroughly on AWS as well.

