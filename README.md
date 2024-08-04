# Gophermart

Gophermart is an educational project for the course "Advanced GO Developer" from Ya.Praktikum.
The project is a simple online store with a RESTful API that allows users to register, log in, and manage their bonuses.

## Requirements

- Docker
- Docker Compose
- Make (for using the Makefile)

## Setup and Launch

1. Clone the repository:

 ```sh
 git clone https://github.com/a-bondar/gophermart.git
 cd gophermart
 ```

2. Set the necessary environment variables in a `.env` file.

3. Build Docker images and start containers:

 ```sh
 make up
 ```

## Environment Variables

Here are the environment variables you can set in your `.env` file and what they mean:

- `RUN_PORT`: The port on which the service will run (default is `8080`).
- `RUN_ADDRESS`: The address and port the service will bind to, set to `:${RUN_PORT}` by default.
- `ACCRUAL_SYSTEM_PORT`: The port on which the accrual service will run (default is `8081`).
- `ACCRUAL_SYSTEM_ADDRESS`: The address of the accrual service which will be used by http requests inside main app, set to `http://accrual:${ACCRUAL_SYSTEM_PORT}` by default.
- `ACCRUAL_RUN_ADDRESS`: The address and port the accrual service will bind to, set to `:${ACCRUAL_SYSTEM_PORT}` by default.
- `DB_PORT`: The port on which the PostgreSQL database will run (default is `5432`).
- `DB_USER`: The username for accessing the PostgreSQL database.
- `DB_PASSWORD`: The password for the PostgreSQL database user.
- `DB_NAME`: The name of the PostgreSQL database.
- `DATABASE_URI`: The connection URI for the PostgreSQL database.
- `JWT_SECRET`: The secret key for JWT token generation.
- `JWT_EXP`: The expiration time for JWT tokens (default is `1h`).

## Basic Commands

### Build Docker Images and Start Containers

To build Docker images and start the containers in the background, run:

 ```sh
 make up
 ```

### Stop and Remove Containers

To stop and remove all containers, run:

 ```sh
 make down
 ```

## Additional Commands

### View Logs

To view logs of all containers in real-time, run:

 ```sh
 make logs
 ```

### Clean Up Unused Docker Data

To clean up unused Docker data, run:

 ```sh
 make clean
 ```

### Start Development Mode with File Watching

To start containers in development mode with file watching, run:

 ```sh
 make develop
 ```

### Run Linter

To run the linter using `golangci-lint`, run:

 ```sh
 make lint
 ```

### Create a new DB migration

To create a new migration, run:

 ```sh
 make db_migration_new
 ```

### Apply DB migrations

To apply all migrations, run:

 ```sh
 make db_migrate_up
 ```

### Rollback DB migrations

To roll back the last migration, run:

 ```sh
 make db_migrate_down
 ```