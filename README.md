![Go Version](https://img.shields.io/badge/go-1.24-00ADD8?logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-16-336791?logo=postgresql&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)

# WMS SaaS

Backend service for inventory management with proper concurrency control. Handles simultaneous stock updates without overselling using version-based optimistic locking.

Go + PostgreSQL + Docker.

## Background

Standard inventory systems have a race condition: when multiple requests try to reserve the last item, they all read the same stock count, decrement it, and write back. You end up shipping more than you have.

Example:

- Request A reads stock: 1 remaining
- Request B reads stock: 1 remaining (same snapshot)
- Both subtract 1 and update
- Final count: 0, but you just sold 2 items

This happens constantly in flash sales or high-traffic checkouts.

## Solution

Stock records include a `version` field that increments on every update. Write operations only succeed if the version matches what was read:

```sql
UPDATE stocks
SET quantity = quantity - $1, version = version + 1
WHERE product_id = $2 AND version = $3
RETURNING version
```

If the version changed between read and write, the WHERE clause fails (zero rows affected). The API returns 409 Conflict. Client retries with fresh data.

No row locks. No blocking reads. Just compare-and-swap at the database level.

Verified with 50 concurrent requests hitting one product. Results: ~800 req/sec throughput, zero negative quantities, accurate final count.

## Architecture

Hexagonal (ports & adapters) to keep domain logic separate from infrastructure:

```
internal/
├── core/
│   ├── domain/       # Entities (Product, Stock)
│   ├── services/     # Business logic & Transaction orchestration
│   └── ports/        # Interfaces (Repository/Service definitions)
└── adapters/
    ├── repository/   # PostgreSQL implementation (pgx/v5)
    ├── handler/      # HTTP Transport (Fiber v2)
    └── config/       # Env configuration
```

Every stock change gets logged to `stock_mutations` table for audit trail.

## Testing Concurrency

Includes a CLI tool to verify behavior under load:

```bash
# Manually add stock to test product
docker exec -it wms-postgres psql -U wmsadmin -d wms_saas
UPDATE stocks SET quantity = 100 WHERE product_id = '<uuid>';

# Run 50 concurrent reservation requests
go run cmd/stress_test/main.go -id <uuid> -n 50 -c 50
```

Typical results (local machine):

- Throughput: ~800 requests/second
- Successful reservations: Limited by available stock
- Failed with 409: Remaining requests (optimistic lock detection)
- Database integrity: No negative values, no lost updates

The 409 responses prove the locking works. That's the database rejecting stale writes.

## API

| Endpoint                    | Method | Description                             |
| --------------------------- | ------ | --------------------------------------- |
| `/api/v1/products`          | POST   | Create product (initializes stock at 0) |
| `/api/v1/products`          | GET    | List products                           |
| `/api/v1/stock/reserve`     | POST   | Decrease stock                          |
| `/api/v1/stock/release`     | POST   | Increase stock                          |
| `/api/v1/stock/:id`         | GET    | Current level                           |
| `/api/v1/stock/:id/history` | GET    | Mutation log                            |

Response format:

```json
{
  "success": true,
  "message": "Stock reserved",
  "data": {...}
}
```

HTTP codes: 200 (success), 400 (validation error), 404 (not found), 409 (version conflict), 500 (server error)

## Setup

```bash
# Start services
docker-compose up -d

# Verify
curl http://localhost:8080/health

# Create product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"sku":"ITEM-001","name":"Test Item","description":"Testing"}'

# Reserve stock (use ID from response)
curl -X POST http://localhost:8080/api/v1/stock/reserve \
  -H "Content-Type: application/json" \
  -d '{"product_id":"<uuid>","quantity":1,"order_id":"ORDER-001"}'
```

Database: `docker exec -it wms-postgres psql -U wmsadmin -d wms_saas`

## Schema

**products** - SKU, name, description  
**stocks** - Quantity + version counter  
**stock_mutations** - Audit log (immutable)

Indexes on SKU lookups, product_id joins, timestamp sorting.

Database constraint: `CHECK (quantity >= 0)` prevents negative stock at write time.

## Config

Environment variables (`.env`):

- `DB_USER`, `DB_PASSWORD`, `DB_NAME` - PostgreSQL connection
- `API_PORT` - Server port (default 8080)
- `JWT_SECRET` - Token signing (currently unused)

## Technical Details

**Why optimistic over pessimistic locking**  
`SELECT FOR UPDATE` blocks all concurrent reads. Optimistic locking only fails conflicting writes, so reads never wait. Better for read-heavy workloads.

**Why pgx instead of database/sql**  
30-40% faster for PostgreSQL. Better type safety. Native support for RETURNING clauses and batch operations.

**Transaction propagation**  
Context values carry transaction state. Repository methods automatically join active transactions without explicit passing.

## Planned Improvements

- JWT authentication for multi-tenant access control
- Swagger/OpenAPI specification
- Redis caching layer for frequently accessed inventory
- Prometheus metrics (conflict rate, latency percentiles)

---

_Go 1.24 · PostgreSQL 16 · Fiber v2_
