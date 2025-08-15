package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	messageUsecase "multi-tenant-service/internal/message/usecase"
	"multi-tenant-service/internal/tenant/repository"
	"multi-tenant-service/internal/tenant/usecase"

	"multi-tenant-service/package/connection/database"
	rabbitmq "multi-tenant-service/package/rabbit-mq"
	"multi-tenant-service/package/structs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSuite struct {
	db             *database.DB
	rabbitmq       *rabbitmq.Client
	tenantManager  usecase.ITenantUsecase
	messageService *messageUsecase.IMessageUsecase
	msgRepo        rm.IMessageRepository
	tenantRepo     repository.ITenantRepository
	router         http.Handler
	pool           *dockertest.Pool
	pgResource     *dockertest.Resource
	rmqResource    *dockertest.Resource
}

func setupTestSuite(t *testing.T) *TestSuite {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	// Start PostgreSQL container
	pgResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_PASSWORD=testpass",
			"POSTGRES_DB=testdb",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err)

	// Start RabbitMQ container
	rmqResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "rabbitmq",
		Tag:        "3-management",
		Env: []string{
			"RABBITMQ_DEFAULT_USER=guest",
			"RABBITMQ_DEFAULT_PASS=guest",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err)

	// Set expiry for containers
	pgResource.Expire(120)
	rmqResource.Expire(120)

	var db *database.DB
	var rabbitmqClient *rabbitmq.Client

	// Wait for PostgreSQL
	pool.Retry(func() error {
		databaseURL := fmt.Sprintf("postgres://postgres:testpass@localhost:%s/testdb?sslmode=disable", pgResource.GetPort("5432/tcp"))
		var dbErr error
		db, dbErr = database.Connect(databaseURL)
		if dbErr != nil {
			return dbErr
		}
		return db.Ping()
	})

	// Wait for RabbitMQ
	pool.Retry(func() error {
		rabbitmqURL := fmt.Sprintf("amqp://guest:guest@localhost:%s/", rmqResource.GetPort("5672/tcp"))
		var rmqErr error
		rabbitmqClient, rmqErr = rabbitmq.NewClient(rabbitmqURL)
		return rmqErr
	})

	// Run migrations
	databaseURL := fmt.Sprintf("postgres://postgres:testpass@localhost:%s/testdb?sslmode=disable", pgResource.GetPort("5432/tcp"))
	fmt.Println("databaseURL", databaseURL)
	db, err = database.Connect(databaseURL)
	require.NoError(t, nil)


	// Initialize services
	tenantManager := usecase.NewTenantUsecase(db, rabbitmqClient)
	messageService := messageUsecase.NewMessageUsecase(msgRepo, rabbitmqClient)

	// Setup router
	// router := api.ServeAPI(tenantManager, messageService)

	return &TestSuite{
		db:             db,
		rabbitmq:       rabbitmqClient,
		tenantManager:  tenantManager,
		// messageService: messageService,
		// router:         router,
		pool:           pool,
		pgResource:     pgResource,
		rmqResource:    rmqResource,
	}
}

func (ts *TestSuite) teardown() {
	if ts.db != nil {
		ts.db.Close()
	}
	if ts.rabbitmq != nil {
		ts.rabbitmq.Close()
	}
	if ts.pool != nil {
		ts.pool.Purge(ts.pgResource)
		ts.pool.Purge(ts.rmqResource)
	}
}

func TestTenantLifecycle(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown()

	// Test tenant creation
	createReq := structs.CreateTenantRequest{
		Name:              "Test Tenant",
		ConcurrencyConfig: 5,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var tenant structs.Tenant
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tenant))
	assert.Equal(t, "Test Tenant", tenant.Name)
	assert.Equal(t, 5, tenant.ConcurrencyConfig)

	tenantID := tenant.ID.String()

	// Test tenant retrieval
	req = httptest.NewRequest("GET", "/api/v1/tenants/"+tenantID, nil)
	w = httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test concurrency update
	updateReq := structs.UpdateConcurrencyRequest{Workers: 10}
	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", "/api/v1/tenants/"+tenantID+"/config/concurrency", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test tenant deletion
	req = httptest.NewRequest("DELETE", "/api/v1/tenants/"+tenantID, nil)
	w = httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify tenant is deleted
	req = httptest.NewRequest("GET", "/api/v1/tenants/"+tenantID, nil)
	w = httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMessagePublishingAndConsumption(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown()

	ctx := context.Background()

	// Create tenant first
	tenant, err := ts.tenantManager.CreateTenant(ctx, structs.CreateTenantRequest{
		Name:              "Message Test Tenant",
		ConcurrencyConfig: 3,
	})
	require.NoError(t, err)

	// Publish message
	messageReq := structs.CreateMessageRequest{
		TenantID: tenant.ID,
		Payload: map[string]interface{}{"test": "data", "value": 123},
	}

	body, _ := json.Marshal(messageReq)
	req := httptest.NewRequest("POST", "/api/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)

	// Wait for message processing
	time.Sleep(2 * time.Second)

	// Get messages
	req = httptest.NewRequest("GET", "/api/v1/messages?tenant_id="+tenant.ID.String(), nil)
	w = httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response structs.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Len(t, response.Data, 1)
	assert.Equal(t, tenant.ID, response.Data[0].TenantID)
}

func TestCursorPagination(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown()

	ctx := context.Background()

	// Create tenant
	tenant, err := ts.tenantManager.CreateTenant(ctx, structs.CreateTenantRequest{
		Name:              "Pagination Test Tenant",
		ConcurrencyConfig: 3,
	})
	require.NoError(t, err)

	// Insert test messages directly into database
	for i := 0; i < 15; i++ {
		_, err := ts.db.ExecContext(ctx, 
			"INSERT INTO messages (tenant_id, payload) VALUES ($1, $2)",
			tenant.ID, fmt.Sprintf(`{"message": %d}`, i))
		require.NoError(t, err)
	}

	// Test pagination
	req := httptest.NewRequest("GET", "/api/v1/messages?tenant_id="+tenant.ID.String()+"&limit=5", nil)
	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response structs.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Len(t, response.Data, 5)
	assert.NotNil(t, response.NextCursor)

	// Test next page
	req = httptest.NewRequest("GET", "/api/v1/messages?tenant_id="+tenant.ID.String()+"&cursor="+*response.NextCursor+"&limit=5", nil)
	w = httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var nextResponse structs.MessageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &nextResponse))
	assert.Len(t, nextResponse.Data, 5)
}

func TestInvalidRequests(t *testing.T) {
	ts := setupTestSuite(t)
	defer ts.teardown()

	// Test invalid tenant creation
	req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewReader([]byte(`{"invalid": "json"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test get non-existent tenant
	req = httptest.NewRequest("GET", "/api/v1/tenants/"+uuid.New().String(), nil)
	w = httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test invalid tenant ID
	req = httptest.NewRequest("GET", "/api/v1/tenants/invalid-id", nil)
	w = httptest.NewRecorder()
	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}