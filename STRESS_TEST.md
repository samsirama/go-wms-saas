# Stress Testing the WMS API

## About

This tool verifies that the optimistic locking mechanism works correctly under high concurrency. It launches 50 simultaneous requests to reserve stock from the same product.

## Expected Behavior

When 50 concurrent requests try to reserve 1 item each from a product with limited stock:

- Some requests succeed (200 OK)
- Some fail with 409 Conflict (optimistic lock) - **This is good!**
- The final stock level should be accurate (no overselling)

## Setup

1. Start the API server:

```bash
docker-compose up -d
```

2. Create a test product and note its ID:

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"sku":"STRESS-001","name":"Stress Test Product","description":"For testing"}'
```

3. Add some initial stock (e.g., 100 units):

```bash
# Get the product UUID from the response above
curl -X POST http://localhost:8080/api/v1/stock/reserve \
  -H "Content-Type: application/json" \
  -d '{"product_id":"YOUR-PRODUCT-UUID","quantity":-100,"order_id":"INIT"}'
```

Wait, that won't work since we validate quantity > 0. Instead, you'll need to manually INSERT stock:

```bash
docker exec -it wms-postgres psql -U wmsadmin -d wms_saas

UPDATE stocks SET quantity = 100 WHERE product_id = 'YOUR-PRODUCT-UUID';
```

4. Edit `cmd/stress_test/main.go` and replace `PRODUCT_ID` with your actual UUID

5. Run the stress test:

```bash
go run cmd/stress_test/main.go
```

## What to Look For

**Good results:**

- Mix of 200 (success) and 409 (conflict) responses
- Total successful reservations ≤ available stock
- No negative stock in database

**Bad results (bug in optimistic locking):**

- All requests succeed even when stock is insufficient
- Database shows negative quantity
- No 409 conflicts when stock runs out

## Verifying Final Stock

After the test, check the database:

```bash
docker exec -it wms-postgres psql -U wmsadmin -d wms_saas

SELECT s.quantity, s.version, p.name
FROM stocks s
JOIN products p ON s.product_id = p.id
WHERE p.sku = 'STRESS-001';
```

The quantity should equal: `initial_stock - successful_reservations`

## Cleanup

```bash
# Delete test product and stock
docker exec -it wms-postgres psql -U wmsadmin -d wms_saas

DELETE FROM products WHERE sku = 'STRESS-001';
```
