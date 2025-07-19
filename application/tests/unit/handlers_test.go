package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"octa-byte-ai/internal/handlers"
	"octa-byte-ai/internal/models"
	"octa-byte-ai/pkg/logger"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHealthHandler(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Mock successful ping
	mock.ExpectPing().WillReturnError(nil)

	// Create handlers
	log := logger.New()
	h := handlers.New(db, log)

	// Create request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	h.Health(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("Handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	// Check response status
	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestCreateUserHandler(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create handlers
	log := logger.New()
	h := handlers.New(db, log)

	// Test data
	user := models.CreateUserRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Mock database query
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
		AddRow(1, user.Name, user.Email, "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Name, user.Email).
		WillReturnRows(rows)

	// Create request body
	body, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler with middleware context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "correlation_id", "test-correlation-id")
	req = req.WithContext(ctx)

	h.CreateUser(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Parse response
	var response models.User
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	// Check response data
	if response.Name != user.Name {
		t.Errorf("Expected name %v, got %v", user.Name, response.Name)
	}
	if response.Email != user.Email {
		t.Errorf("Expected email %v, got %v", user.Email, response.Email)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetUsersHandler(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create handlers
	log := logger.New()
	h := handlers.New(db, log)

	// Mock database query
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
		AddRow(1, "John Doe", "john@example.com", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z").
		AddRow(2, "Jane Smith", "jane@example.com", "2023-01-01T00:00:00Z", "2023-01-01T00:00:00Z")

	mock.ExpectQuery("SELECT id, name, email, created_at, updated_at FROM users").
		WillReturnRows(rows)

	// Create request
	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler with middleware context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "correlation_id", "test-correlation-id")
	req = req.WithContext(ctx)

	h.GetUsers(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var users []models.User
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Fatal(err)
	}

	// Check response data
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %v", len(users))
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
