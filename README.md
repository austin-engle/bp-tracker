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
