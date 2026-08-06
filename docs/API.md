# Chirpy API Reference

Base URL: `http://localhost:8080`

Unless stated otherwise, JSON endpoints use `Content-Type: application/json`.

## Health check

```http
GET /api/healthz
```

Returns `200 OK` with the text `OK` when the server is ready.

## Users

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

## Authentication

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

## Chirps

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

## Webhooks

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

## Static files and administration

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
