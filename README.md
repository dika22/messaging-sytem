# Multi-Tenant Messaging System

A production-ready Go application that implements a multi-tenant messaging system using RabbitMQ and PostgreSQL with dynamic consumer management, partitioned data storage, and configurable concurrency.

## Features

- **Auto-Spawn Tenant Consumer**: Automatically creates RabbitMQ queues and consumers for new tenants
- **Auto-Stop Tenant Consumer**: Gracefully stops and cleans up resources when tenants are deleted
- **Partitioned Message Storage**: Uses PostgreSQL table partitioning for efficient multi-tenant data isolation
- **Configurable Concurrency**: Dynamic worker pool management per tenant
- **Graceful Shutdown**: Proper cleanup of all resources during application shutdown
- **Cursor Pagination**: Efficient pagination for message retrieval
- **Swagger Documentation**: Complete API documentation with interactive UI
- **Integration Tests**: Comprehensive tests using Docker containers
- **Production Ready**: Includes Docker, monitoring hooks, and proper error handling


## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional, for convenience commands)

### Running Locally

1. Clone the repository:
```bash
git clone <repository-url>
cd messaging-system
```

2. Install dependencies:
```bash
make deps
```

3. Start PostgreSQL and RabbitMQ:
```bash
docker-compose up postgres rabbitmq -d
```

4. Run the application:
```bash
make serve-http
```

5. The application will be available at:
   - API: http://localhost:8080
   - Swagger UI: http://localhost:8080/swagger/index.html
   - RabbitMQ Management: http://localhost:15672 (guest/guest)


### Running with Docker Compose

1. Start the services:
```bash
make docker-up
```
. Start the services:
```bash
make docker-up
```

## API Usage

### 1. Create a Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "concurrency_config": 5
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corp",
  "concurrency_config": 5,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### 2. Publish Messages

```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "payload": {
      "type": "order",
      "order_id": "12345",
      "amount": 99.99
    }
  }'
```

### 3. Retrieve Messages with Pagination

```bash
# First page
curl "http://localhost:8080/api/v1/messages?tenant_id=550e8400-e29b-41d4-a716-446655440000&limit=10"

# Next page with cursor
curl "http://localhost:8080/api/v1/messages?tenant_id=550e8400-e29b-41d4-a716-446655440000&cursor=2024-01-01T12:30:00Z&limit=10"
```

### 4. Update Tenant Concurrency

```bash
curl -X PUT http://localhost:8080/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000/config/concurrency \
  -H "Content-Type: application/json" \
  -d '{"workers": 10}'
```

### 5. Delete Tenant

```bash
curl -X DELETE http://localhost:8080/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000
```

## Testing

### Unit Tests

```bash
make test
```

### Integration Tests

Integration tests use Docker containers for PostgreSQL and RabbitMQ:

```bash
make test-integration
```

The tests cover:
- Complete tenant lifecycle (create, update, delete)
- Message publishing and consumption
- Cursor pagination
- Error handling and edge cases
- Concurrency updates
- Consumer management

## Database Schema

### Tenants Table

```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    concurrency_config INTEGER DEFAULT 3,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Messages Table (Partitioned)

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY LIST (tenant_id);

-- Individual partitions are created automatically:
-- messages_tenant_{tenant_id} PARTITION OF messages FOR VALUES IN ('{tenant_id}')
```
