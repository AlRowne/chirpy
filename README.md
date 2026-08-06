# Chirpy

Chirpy is a small social-media-style REST API written in Go. Users can create accounts, authenticate with JWT access tokens, manage refresh tokens, publish chirps, and delete their own chirps.

The project uses PostgreSQL for persistence and `sqlc`-generated database code.

## Features

- User registration and profile updates
- Password hashing with Argon2id
- JWT-based access-token authentication
- Refresh-token creation, renewal, and revocation
- Chirp creation, listing, filtering, and deletion
- Chirpy Red user upgrades through a Polka webhook
- Static file serving from `/app/`
- Request metrics for the static file server
- Development-only database reset endpoint

## Requirements

- Go `1.26.5` or a compatible newer Go version
- PostgreSQL
- A database with the migrations from `sql/schema` applied

## Configuration

Create a `.env` file in the project root. The `.env` file is ignored by Git and should not contain values that are committed to the repository.

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET=replace-with-a-long-random-secret
POLKA_KEY=replace-with-your-polka-api-key
```

### Environment variables

| Variable | Description |
| --- | --- |
| `DB_URL` | PostgreSQL connection string used by the application. |
| `PLATFORM` | Runtime platform. The reset endpoint is only available when this is `dev`. |
| `SECRET` | Secret used to sign and validate JWT access tokens. |
| `POLKA_KEY` | API key expected by the Polka webhook endpoint. |

## Database setup

The migration files are located in `sql/schema` and are written in Goose-compatible format. Apply them in filename order:

1. `001_users.sql`
2. `002_chirps.sql`
3. `003_user_passwords.sql`
4. `004_refresh_tokens.sql`
5. `005_is_chirpy_red.sql`

The database schema contains users, chirps, refresh tokens, password hashes, and the `is_chirpy_red` user flag.

The SQL queries used by the application are located in `sql/queries`. Generated Go code is stored in `internal/database`.

## Running the application

Install dependencies and start the server with:

```bash
go mod download
go run .
```

The server listens on:

```text
http://localhost:8080
```

The application loads environment variables from `.env` during startup and exits if the file cannot be loaded or the database cannot be opened.

## API reference

The complete API documentation is available in [`docs/API.md`](docs/API.md).

It contains request formats, authentication requirements, response behavior, and examples for all health, user, authentication, chirp, webhook, static-file, metrics, and administration endpoints.

## Project structure

```text
.
├── main.go                         # Server setup and route registration
├── config.go                       # API configuration
├── handler_health.go               # Health endpoint
├── handler_metrics.go              # Static-file metrics middleware and endpoint
├── handler_admin.go                # Development reset endpoint
├── handler_auth.go                 # Login, refresh, and revoke handlers
├── handler_users.go                # User registration and profile updates
├── handler_chirps.go               # Chirp handlers
├── handler_webhooks.go             # Polka webhook handler
├── json.go                         # Shared JSON response helpers
├── internal/database               # sqlc-generated PostgreSQL code
│   └── auth                        # Password, JWT, and token helpers
├── sql/schema                      # Database migrations
├── sql/queries                     # SQL queries used by sqlc
├── docs/API.md                     # Detailed API reference
├── index.html                      # Static application page
└── LICENSE                         # MIT License
```

## Regenerating database code

The repository contains the SQL configuration in `sqlc.yaml`. If `sqlc` is installed, regenerate the database package from the schema and queries with:

```bash
sqlc generate
```

Review generated changes before committing them.

## Authentication overview

1. A user registers with an email address and password.
2. The password is stored as an Argon2id hash.
3. Login returns a short-lived JWT access token and a long-lived refresh token.
4. Protected endpoints receive the JWT as `Authorization: Bearer <token>`.
5. The refresh endpoint exchanges a valid refresh token for a new JWT.
6. Refresh tokens can be revoked explicitly through `/api/revoke`.

## License

This project is licensed under the [MIT License](LICENSE).
