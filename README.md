# VulnPulse — Event-Driven Vulnerability Hub

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)
![Docker](https://img.shields.io/badge/docker-ready-2496ED?logo=docker)

**VulnPulse** is a multi-tenant, event-driven vulnerability management platform that ingests vulnerability data (CVEs), matches them to tenant assets asynchronously, generates alerts, and notifies external systems via signed webhooks.

## 🚀 Features

- **Multi-Tenant Architecture** - Complete data isolation with tenant-scoped access control
- **Event-Driven Design** - Asynchronous job processing with RabbitMQ
- **REST API** - Comprehensive API for managing assets, vulnerabilities, and alerts
- **JWT Authentication** - Secure token-based auth with role-based access control (RBAC)
- **Webhook Notifications** - HMAC-signed webhooks with retry logic
- **Vulnerability Matching** - Automatic impact analysis when vulnerabilities are ingested
- **Alert Management** - Track and manage security alerts across your infrastructure
- **Production-Ready** - Docker Compose setup, structured logging, graceful shutdown

## 🏗️ Architecture

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   API       │─────▶│  RabbitMQ    │─────▶│   Worker    │
│  Service    │      │   (Jobs)     │      │   Service   │
└─────────────┘      └──────────────┘      └─────────────┘
       │                                           │
       ├──────────────────┬────────────────────────┤
       │                  │                        │
  ┌────▼────┐       ┌────▼────┐            ┌─────▼─────┐
  │Postgres │       │  Redis  │            │ Webhooks  │
  └─────────┘       └─────────┘            └───────────┘
```

**Components:**

- **API Service**: REST endpoints for managing vulnerabilities, assets, alerts, webhooks
- **Worker Service**: Consumes jobs from RabbitMQ, performs vulnerability-to-asset matching
- **Webhook Dispatcher**: Delivers signed webhook payloads with exponential backoff retry
- **PostgreSQL**: System of record for all data
- **Redis**: Caching and deduplication
- **RabbitMQ**: Message queue for async job processing

## 📦 Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Message Queue**: RabbitMQ 3
- **HTTP Router**: Gorilla Mux
- **Authentication**: JWT with bcrypt
- **Migrations**: Goose
- **Containerization**: Docker & Docker Compose

## 🚦 Quick Start

### Prerequisites

- Docker & Docker Compose
- `jq` (for demo script)
- `psql` (for seed script)

### 1. Start Services

```bash
docker-compose up -d
```

This starts:

- PostgreSQL (`:5432`)
- Redis (`:6379`)
- RabbitMQ (`:5672`, management UI at `:15672`)
- API Service (`:8080`)
- Worker Service (×2 replicas)

### 2. Seed Demo Data

```bash
./scripts/seed.sh
```

This creates:

- Demo tenant: "Demo Corp"
- Admin user: `admin@democorp.com` / `admin123`
- Sample assets: nodejs, express, lodash, apache-httpd, nginx

### 3. Run Demo

```bash
./scripts/demo.sh
```

The demo will:

1. Login and get JWT token
2. List existing assets
3. Create a new asset (log4j 2.14.1)
4. Ingest Log4Shell vulnerability (CVE-2021-44228)
5. Wait for worker to match vulnerability to assets
6. Display generated alerts
7. Update alert status to "acknowledged"

## 📚 API Documentation

### Authentication

**Login**

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@democorp.com",
  "password": "admin123"
}

Response:
{
  "token": "eyJhbGc...",
  "expires_at": "2024-02-11T13:42:00Z",
  "user": {
    "id": "...",
    "email": "admin@democorp.com",
    "role": "admin",
    "tenant_id": "..."
  }
}
```

### Assets

**Create Asset**

```bash
POST /api/v1/assets
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "log4j",
  "type": "package",
  "version": "2.14.1",
  "metadata": {}
}
```

**List Assets**

```bash
GET /api/v1/assets
Authorization: Bearer <token>
```

**Update Asset**

```bash
PUT /api/v1/assets/{id}
Authorization: Bearer <token>
```

**Delete Asset**

```bash
DELETE /api/v1/assets/{id}
Authorization: Bearer <token>
```

### Vulnerabilities

**Ingest Vulnerability**

```bash
POST /api/v1/vulnerabilities
Authorization: Bearer <token>
Content-Type: application/json

{
  "cve_id": "CVE-2021-44228",
  "title": "Log4Shell RCE",
  "description": "...",
  "severity": "critical",
  "affected_products": [
    {
      "name": "log4j",
      "versions": ["2.14.1", "2.15.0"]
    }
  ]
}
```

**List Vulnerabilities**

```bash
GET /api/v1/vulnerabilities?severity=critical&limit=50&offset=0
Authorization: Bearer <token>
```

### Alerts

**List Alerts**

```bash
GET /api/v1/alerts?status=open&limit=50
Authorization: Bearer <token>
```

**Update Alert Status**

```bash
PATCH /api/v1/alerts/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "acknowledged"
}

Valid statuses: open, acknowledged, resolved, false_positive
```

### Health Check

```bash
GET /health
```

## 🔐 Authentication & Authorization

### Roles

- **Admin**: Full access, can manage tenants and users
- **Analyst**: Can manage assets, vulnerabilities, alerts, webhooks (own tenant)
- **Viewer**: Read-only access (own tenant)

### JWT Token

Tokens are valid for 24 hours by default (configurable via `JWT_DURATION_HOURS`).

Include in requests:

```
Authorization: Bearer <token>
```

## 🔧 Configuration

Environment variables:

| Variable             | Default          | Description                  |
| -------------------- | ---------------- | ---------------------------- |
| `PORT`               | `8080`           | API server port              |
| `DATABASE_URL`       | `postgres://...` | PostgreSQL connection string |
| `REDIS_URL`          | `localhost:6379` | Redis address                |
| `RABBITMQ_URL`       | `amqp://...`     | RabbitMQ connection string   |
| `JWT_SECRET`         | _required_       | Secret for signing JWTs      |
| `JWT_DURATION_HOURS` | `24`             | Token validity duration      |

## 🧪 Development

### Build Locally

```bash
# Build API
go build -o bin/api ./cmd/api

# Build Worker
go build -o bin/worker ./cmd/worker

# Run migrations
go run cmd/migrate/main.go up
```

### Run Tests

```bash
go test ./... -v -race -cover
```

### Access Management UIs

- **RabbitMQ Management**: http://localhost:15672 (vulnpulse/devpassword)
- **API**: http://localhost:8080

## 📈 Production Considerations

- Change default passwords in `docker-compose.yml`
- Set a strong `JWT_SECRET` environment variable
- Use connection pooling (configured in `pkg/database/postgres.go`)
- Enable TLS for all services
- Add Prometheus metrics (TODO)
- Set up log aggregation (already JSON-structured)
- Configure webhook retry limits
- Add rate limiting per tenant

## 🛠️ Project Structure

```
vulnpulse/
├── cmd/
│   ├── api/            # API service entry point
│   └── worker/         # Worker service entry point
├── internal/
│   ├── api/            # HTTP handlers, middleware, routes
│   ├── worker/         # Job consumer
│   ├── service/        # Business logic (matching)
│   ├── repository/     # Database access layer
│   └── domain/         # Core models
├── pkg/
│   ├── auth/           # JWT, RBAC
│   ├ config/          # Configuration
│   ├── database/       # DB connection, migrations
│   ├── logger/         # Structured logging
│   ├── queue/          # RabbitMQ client
│   └── cache/          # Redis client
├── migrations/         # SQL migrations
└── scripts/            # Seed, demo scripts
```

## 📝 License

MIT

## 🤝 Contributing

Contributions welcome! Please open an issue or PR.

## 📧 Support

For questions or issues, please open a GitHub issue.

---

**Built with ❤️ using Go**
# vuln-pulse
