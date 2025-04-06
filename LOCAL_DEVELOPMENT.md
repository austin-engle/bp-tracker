# Local Development Setup

This project uses Docker Compose to simplify local development and provide an environment that closely mimics the target AWS Lambda setup.

## Prerequisites

*   Docker Desktop (or Docker Engine + Docker Compose)
*   Go (if you need to modify Go code outside Docker)

## Running Locally

1.  **Build and Start Containers:**
    Open a terminal in the project root (`bp-tracker`) and run:
    ```bash
    docker-compose up --build
    ```
    *   The `--build` flag ensures the Docker image is rebuilt if you've made code changes.
    *   This command will:
        *   Build the Go application into a binary inside a Docker container.
        *   Start a PostgreSQL container named `db`.
        *   Start the application container named `app`.
        *   The `app` container uses environment variables defined in `docker-compose.yml` to connect to the `db` container.

2.  **Accessing the Application:**
    *   The `docker-compose.yml` file configures the `app` service to potentially expose the Lambda Runtime Interface Emulator (RIE) on port `9000` of your host machine.
    *   You can interact with this endpoint using tools like `curl` or Postman to simulate API Gateway events. For example:
        ```bash
        # Example: Sending a simple GET request (adjust payload/path as needed)
        curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{
          "httpMethod": "GET",
          "path": "/",
          "headers": {},
          "body": null
        }'
        ```
    *   Refer to the AWS Lambda RIE documentation for details on how to structure invocation events.

3.  **Accessing the Database:**
    *   The PostgreSQL database container (`db`) exposes port `5432` on your host machine.
    *   You can connect to this database using any standard PostgreSQL client (e.g., `psql`, DBeaver, TablePlus).
    *   Connection details (from `docker-compose.yml`):
        *   **Host:** `localhost`
        *   **Port:** `5432`
        *   **User:** `postgres`
        *   **Password:** `password`
        *   **Database:** `bp_tracker_local`

4.  **Applying Schema:**
    *   The first time you run the setup, or if you clear the database volume, you need to apply the database schema.
    *   Connect to the local PostgreSQL database (see step 3) and execute the contents of `internal/database/schema.sql`.
    *   Example using `psql` (requires `psql` installed locally):
        ```bash
        psql -h localhost -p 5432 -U postgres -d bp_tracker_local -f internal/database/schema.sql
        # Enter 'password' when prompted
        ```

5.  **Stopping Containers:**
    *   Press `Ctrl+C` in the terminal where `docker-compose up` is running.
    *   To remove the containers (but keep the persistent database volume), run:
        ```bash
        docker-compose down
        ```
    *   To remove containers AND the database volume (deleting all local data):
        ```bash
        docker-compose down -v
        ```

## Code Changes

*   Modify Go code in the `internal/` and `cmd/` directories as needed.
*   Run `docker-compose up --build` again to see your changes reflected.
