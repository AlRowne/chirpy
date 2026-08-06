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

## Testing

Run all tests with:

```bash
go test ./...
```

Format the Go source files with:

```bash
gofmt -w *.go internal/database/auth/*.go
```

## API reference

Unless stated otherwise, JSON endpoints use `Content-Type: application/json`.

### Health check

```http
GET /api/healthz
```

Returns `200 OK` with the text `OK` when the server is ready.

### Create a user

```http
POST /api/users
Content-Type: application/json
```

Request body:

```json
{
  "email": "alice@example.com",
  "password": "secret-password"
}
```

Returns `201 Created` with the new user's public information. The password hash is never returned.

### Update the current user

```http
PUT /api/users
Authorization: Bearer <access-token>
Content-Type: application/json
```

Request body:

```json
{
  "email": "new-address@example.com",
  "password": "new-password"
}
```

Returns `200 OK` with the updated public user information.

### Log in

```http
POST /api/login
Content-Type: application/json
```

Request body:

```json
{
  "email": "alice@example.com",
  "password": "secret-password"
}
```

Returns an access token and a refresh token:

```json
{
  "id": "00000000-0000-0000-0000-000000000000",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z",
  "email": "alice@example.com",
  "is_chirpy_red": false,
  "token": "<jwt-access-token>",
  "refresh_token": "<refresh-token>"
}
```

Access tokens expire after one hour. Refresh tokens expire after 60 days unless revoked.

### Refresh an access token

```http
POST /api/refresh
Authorization: Bearer <refresh-token>
```

Returns a new JWT access token.

### Revoke a refresh token

```http
POST /api/revoke
Authorization: Bearer <refresh-token>
```

Returns `204 No Content` when the refresh token has been revoked.

### Create a chirp

```http
POST /api/chirps
Authorization: Bearer <access-token>
Content-Type: application/json
```

Request body:

```json
{
  "body": "Hello from Chirpy!"
}
```

Chirp bodies may contain up to 140 characters. Returns `201 Created` with the created chirp.

### List chirps

```http
GET /api/chirps
```

Returns all chirps ordered by creation time.

To filter chirps by author, provide the author's UUID:

```http
GET /api/chirps?author_id=<user-uuid>
```

### Get one chirp

```http
GET /api/chirps/{chirpID}
```

Returns the chirp matching the UUID in the path.

### Delete a chirp

```http
DELETE /api/chirps/{chirpID}
Authorization: Bearer <access-token>
```

Only the user who created the chirp may delete it. A successful deletion returns `204 No Content`.

### Polka webhook

```http
POST /api/polka/webhooks
Authorization: ApiKey <polka-api-key>
Content-Type: application/json
```

The endpoint handles `user.upgraded` events:

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "00000000-0000-0000-0000-000000000000"
  }
}
```

A valid upgrade sets the user's `is_chirpy_red` flag to `true`. Other event types are acknowledged with `204 No Content` without changing the user.

### Static files

```http
GET /app/...
```

Files are served from the project root through the `/app/` prefix. Static file requests increment the file-server metric counter.

### Metrics

```http
GET /admin/metrics
```

Returns an HTML page showing the number of requests served through `/app/`.

### Development reset

```http
POST /admin/reset
```

Resets the static file hit counter and deletes all users when `PLATFORM=dev`. The user records, related chirps, and refresh tokens are removed through the database's cascading foreign-key relationships.

This endpoint is forbidden when `PLATFORM` is not `dev` and should not be exposed in production.

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
│   └── auth                       # Password, JWT, and token helpers
├── sql/schema                      # Database migrations
├── sql/queries                     # SQL queries used by sqlc
└── index.html                      # Static application page
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

No license has been specified for this project yet.
