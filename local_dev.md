# Local Development Guide

This guide helps you run your Go application on your local machine while connecting to a PostgreSQL database in Docker.

## Prerequisites

- Docker and Docker Compose installed on your machine
- Go installed on your local machine (matching the version in your project)
- TailwindCSS, Templ, and Make installed locally (if needed)

## Setup Steps

### 1. Install Required Local Tools

#### Install Templ

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

#### Install TailwindCSS

For macOS or Linux:

```bash
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/x64/;s/aarch64/arm64/')
chmod +x tailwindcss-*
sudo mv tailwindcss-* /usr/local/bin/tailwindcss
```

For Windows, download from the TailwindCSS GitHub releases page and add it to your PATH.

#### Install Make

- **macOS**: Comes with Xcode Command Line Tools (`xcode-select --install`)
- **Linux**: Use your package manager (e.g., `sudo apt install make` for Debian/Ubuntu)
- **Windows**: Install via Chocolatey (`choco install make`) or use WSL

### 2. Set Up the Database

1. Create a `.env` file by copying the template:

   ```bash
   cp .env.example .env
   ```

2. Start the PostgreSQL database:
   ```bash
   docker-compose up -d
   ```

### 3. Configure Your Go Application

Make sure your application reads environment variables for database connection:

```go
// Example database connection in your Go code
db, err := sql.Open("postgres", fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    os.Getenv("POSTGRES_HOSTNAME"),
    os.Getenv("POSTGRES_PORT"),
    os.Getenv("POSTGRES_USER"),
    os.Getenv("POSTGRES_PASSWORD"),
    os.Getenv("POSTGRES_DB"),
))
```

### 4. Run Your Go Application Locally

1. Load environment variables:

   ```bash
   export $(grep -v '^#' .env | xargs)
   ```

2. Run your application:

   ```bash
   go run main.go
   ```

   Or with Make:

   ```bash
   make run
   ```

### 5. Accessing Your Database

To connect to the PostgreSQL database directly:

```bash
docker-compose exec db psql -U postgres -d your_database_name
```

Or using a local PostgreSQL client:

```bash
psql -h localhost -p 5432 -U postgres -d your_database_name
```

### 6. Stopping the Database

```bash
docker-compose down
```

## Important Notes

1. **Database Connection**: When running your app locally, set `POSTGRES_HOSTNAME=localhost` in your .env file. Your app connects to the database via localhost, which Docker exposes from the container.

2. **Port Conflicts**: Ensure no other services on your machine are using port 5432 to avoid conflicts with the PostgreSQL container.

3. **Development Workflow**: When you make changes to your Go code, you'll need to restart your local application. Changes to database schema may require interacting with the PostgreSQL container.
