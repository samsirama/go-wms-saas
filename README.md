![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-16-336791?logo=postgresql&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)

# WMS SaaS

Warehouse management backend built to handle high-concurrency inventory operations. Solves the classic race-condition problem when hundreds of users try to purchase the last available item simultaneously.

**Stack:** Go (Fiber), PostgreSQL, Redis, Docker

## Problem Statement

Traditional inventory systems fail under heavy concurrent load:

- Simple read-modify-write patterns cause overselling
- Pessimistic locks create bottlenecks
- Lost updates lead to negative stock levels

This service uses **optimistic locking** to ensure data consistency without sacrificing performance.

## Architecture

Hexagonal architecture (ports & adapters) for clean separation of concerns:

```
internal/
├── core/
│   ├── domain/      # Business entities (Product, Stock)
│   ├── services/    # Business logic with optimistic locking
│   └── ports/       # Repository interfaces
└── adapters/
    ├── repository/  # PostgreSQL implementation (pgx)
    ├── handler/     # HTTP layer (Fiber)
    └── config/      # Environment config
```

## How It Works

Every stock record has a `version` field. Updates use compare-and-set:

```sql
UPDATE stocks
SET quantity = quantity - $1, version = version + 1
WHERE product_id = $2 AND version = $3
```

If the version changed between read and write, the query affects 0 rows → transaction fails → caller retries. No overselling.

All mutations are logged to `stock_mutations` table for audit trail.

## Running Locally

```bash
cp .env.example .env
docker-compose up -d
curl http://localhost:8080/health
```

The API runs on `:8080`, PostgreSQL on `:5432`, Redis on `:6379`.

Hot-reload is enabled via Air during development.

## Database Schema

**products** - Product catalog  
**stocks** - Current inventory + version counter  
**stock_mutations** - Immutable audit log of all changes

Indexes on frequently queried columns (SKU, product_id, timestamps).

## Configuration

Set via environment variables:

- `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database credentials
- `REDIS_PASSWORD` - Cache auth
- `JWT_SECRET` - Token signing key

See `.env.example` for full list.

## Development Roadmap

- [x] **Core Domain:** Business logic & entities
- [x] **Database:** Schema migration with Optimistic Locking
- [x] **Infrastructure:** Docker Compose & Air setup
- [ ] **Repository Layer:** Native `pgx` implementation
- [ ] **API:** REST Handlers (Fiber)
- [ ] **Security:** JWT Auth & Rate Limiting
- [ ] **Cache:** Redis integration

## Design Decisions

I went with optimistic locking over pessimistic (`SELECT FOR UPDATE`) because it scales better under load. Reads don't block, conflicts are detected at write time, and the database doesn't hold locks.

Using `pgx` directly instead of `database/sql` for the 30-40% performance gain and better type safety when working with PostgreSQL-specific features.

Hexagonal architecture keeps the core business logic isolated from frameworks. Makes testing easier and means I can swap out Fiber or PostgreSQL later without rewriting the service layer.

## Security

Request body capped at 4MB to prevent memory exhaustion. Read/write timeouts protect against slowloris attacks. All inputs validated. Transactions ensure atomicity.

---

_Go 1.22 · PostgreSQL 16 · Fiber v2_
