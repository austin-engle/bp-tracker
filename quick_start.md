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
