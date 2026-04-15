# url-shortener

A fast, lightweight URL shortener API built in Go with user authentication, custom codes, and rate limiting. Store shortened URLs in PostgreSQL and manage your links with full CRUD operations.

## Features

- **Quick URL Shortening** - Generate short codes for any URL with a single API call
- **User Authentication** - Register for an API key and maintain ownership of your shortened links
- **Custom Codes** - Create branded short codes for your URLs (authenticated users only)
- **Link Management** - List all your shortened URLs and delete codes you no longer need
- **Rate Limiting** - Built-in protection against abuse with configurable limits per user
- **Deduplication** - Anonymous users get the same short code for the same URL (idempotent)
- **Graceful Shutdown** - HTTP server with 15-second timeout for clean shutdown

## Quick Start

### 1. Prerequisites

- Go 1.26+
- PostgreSQL database
- Make (optional, but recommended)

### 2. Configure the Environment

Create a `.env` file in the project root with the following variables:

```bash
# Required
PORT=8080
DB_URL=postgres://user:password@localhost:5432/url_shortener?sslmode=disable

# Optional (defaults shown)
BASE_URL=http://localhost:8080
RATE_LIMIT_ANON_REQUESTS=5
RATE_LIMIT_ANON_WINDOW=24h
RATE_LIMIT_AUTH_REQUESTS=5
RATE_LIMIT_AUTH_WINDOW=1h
```

See [Configuration](#configuration) below for details on each variable.

### 3. Run Database Migrations

```bash
make db/migrate
```

This runs all migrations in the `migrations/` folder. See [Database](#database) for manual migration steps.

### 4. Start the Server

```bash
make run
```

The server will start and log its address and base URL. Ctrl+C to stop gracefully.

### 5. Try It Out

**Register a user and get an API key:**

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"My User"}'
```

Response includes your API key.

**Shorten a URL (anonymous):**

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.example.com/very/long/path"}'
```

Returns: `{"code":"abc123"}`

**Shorten with a custom code (requires authentication):**

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","custom_code":"mycode"}'
```

**List your links (requires authentication):**

```bash
curl http://localhost:8080/links \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Delete a short code (requires authentication):**

```bash
curl -X DELETE http://localhost:8080/mycode \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Returns HTTP 204 No Content.

**Resolve a short code:**

```bash
curl -L http://localhost:8080/abc123
```

Redirects (HTTP 302) to the original URL.

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | **Required** | HTTP server port |
| `DB_URL` | **Required** | PostgreSQL connection string |
| `BASE_URL` | `http://localhost:{PORT}` | Public base URL for short links |
| `RATE_LIMIT_ANON_REQUESTS` | `5` | Max requests per window for anonymous users |
| `RATE_LIMIT_ANON_WINDOW` | `24h` | Time window for anonymous rate limit |
| `RATE_LIMIT_AUTH_REQUESTS` | `5` | Max requests per window for authenticated users |
| `RATE_LIMIT_AUTH_WINDOW` | `1h` | Time window for authenticated rate limit |

### Database

PostgreSQL is required. Use any of these to set up your database:

```bash
# Create database manually
createdb url_shortener

# Then run migrations
make db/migrate

# Or migrate manually
psql $DB_URL -f migrations/001_create_links_table.up.sql

# Rollback migrations
make db/rollback
```

## Development

### Commands

See `make help` for all available commands. Here are the most useful:

```bash
make build          # Build the binary to bin/url-shortener
make run            # Build and run immediately
make test           # Run all tests
make test/cover     # Run tests and open coverage report
make test/race      # Detect data races
make test/e2e       # Run end-to-end tests
make clean          # Remove build artifacts
make compile        # Cross-compile for Linux, Windows, macOS
```

### Testing

Run specific tests:

```bash
go test ./internal/services -run TestLinkService_Create
go test ./internal/handlers -v
```
