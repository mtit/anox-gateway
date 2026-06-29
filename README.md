# anox-gateway

Anox gateway subscribes to `anox-server` service-registry updates over WebSocket and proxies requests by service name.

## Environment

- `ANOX_URL`: anox-server address, default `127.0.0.1:8848`
- `HTTP_HOST`: listen host, default `0.0.0.0`
- `HTTP_PORT`: listen port, default `8080`
- `JWT_SECRET`: HS256 JWT signing secret. If empty, requests are forwarded as anonymous user `0`.

## Routing

Requests are routed as:

- `/api/user/login` -> lowest-load instance of `user-service` at `/login`
- `/api/order/create` -> lowest-load instance of `order-service` at `/create`

The second URL segment is mapped to a registered service name by appending `-service`.

Instance priority is sorted by lower `cpu_percent`, then higher `memory_avail_mb`.

## Forwarded Headers

Gateway forwards common proxy headers to services:

- `X-User-ID`: `user_id` claim from a valid HS256 JWT, or `0` when verification fails.
- `X-Real-IP`
- `X-Forwarded-For`
- `X-Forwarded-Host`
- `X-Forwarded-Proto`
- `X-Anox-Gateway`
- `X-Anox-Service`
- `X-Anox-Instance`

## Run

```bash
go run .
```

## Docker

```bash
docker compose up --build -d
```
