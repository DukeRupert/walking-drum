# Coffee Subscription Service

A coffee subscription management system built with Go.

## Getting Started

1. Clone the repository
2. Copy `.env.example` to `.env` and update the configuration
3. Run `docker-compose up` to start the development environment

## Project Structure

The project follows a clean architecture approach with the following main components:

- `cmd/api`: Application entry point
- `internal/api`: HTTP server setup
- `internal/config`: Configuration management
- `internal/domain`: Domain models and DTOs
- `internal/handlers`: HTTP request handlers
- `internal/middleware`: HTTP middleware
- `internal/repositories`: Data access layer
- `internal/services`: Business logic
- `internal/stripe`: Stripe integration
- `migrations`: Database migration files
- `pkg`: Reusable packages
- `scripts`: Utility scripts

## Development

To run the application locally:

```bash
go run cmd/api/main.go
```

To build the application:

```bash
go build -o coffee-subscription ./cmd/api
```

To run database migrations:

```bash
./scripts/db/migrate.sh
```
