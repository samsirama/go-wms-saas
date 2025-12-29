# API Documentation

Base URL: `http://localhost:8080/api/v1`

## Products

### Create Product

```http
POST /products
Content-Type: application/json

{
  "sku": "LAPTOP-001",
  "name": "MacBook Pro 16\"",
  "description": "Apple MacBook Pro"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "uuid",
    "sku": "LAPTOP-001",
    "name": "MacBook Pro 16\"",
    "description": "Apple MacBook Pro",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

### List Products

```http
GET /products?limit=20&offset=0
```

### Get Product by ID

```http
GET /products/:id
```

## Stock Operations

### Reserve Stock

```http
POST /stock/reserve
Content-Type: application/json

{
  "product_id": "uuid",
  "quantity": 5,
  "order_id": "ORDER-123"
}
```

**Success (200):**

```json
{
  "success": true,
  "message": "Stock reserved successfully",
  "data": {
    "product_id": "uuid",
    "quantity": 5,
    "order_id": "ORDER-123"
  }
}
```

**Conflict (409) - Optimistic Lock:**

```json
{
  "success": false,
  "message": "Stock was modified by another transaction, please retry",
  "data": null
}
```

**Bad Request (400) - Insufficient Stock:**

```json
{
  "success": false,
  "message": "Insufficient stock available",
  "data": null
}
```

### Release Stock

```http
POST /stock/release
Content-Type: application/json

{
  "product_id": "uuid",
  "quantity": 5,
  "order_id": "ORDER-123"
}
```

### Get Stock Level

```http
GET /stock/:id
```

**Response:**

```json
{
  "success": true,
  "message": "Stock level retrieved successfully",
  "data": {
    "id": "uuid",
    "product_id": "uuid",
    "quantity": 95,
    "version": 5,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

### Get Mutation History

```http
GET /stock/:id/history?limit=20&offset=0
```

## Error Codes

- `200` OK
- `201` Created
- `400` Bad Request (validation, insufficient stock)
- `404` Not Found
- `409` Conflict (optimistic lock failure - retry)
- `500` Internal Server Error

## Testing the API

```bash
# Start the server
docker-compose up -d

# Health check
curl http://localhost:8080/health

# Create a product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"sku":"TEST-001","name":"Test Product","description":"A test"}'

# Reserve stock
curl -X POST http://localhost:8080/api/v1/stock/reserve \
  -H "Content-Type: application/json" \
  -d '{"product_id":"<uuid>","quantity":1,"order_id":"ORD-001"}'
```
