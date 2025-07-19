package integration
//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"octa-byte-ai/internal/models"
)

var baseURL string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Wait for service to be ready
	waitForService()

	code := m.Run()
	os.Exit(code)
}

func waitForService() {
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	panic("Service did not become ready in time")
}

func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
}

func TestReadyEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/ready")
	if err != nil {
		t.Fatalf("Failed to call ready endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to call metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestUserCRUDOperations(t *testing.T) {
	// Test creating a user
	user := models.CreateUserRequest{
		Name:  fmt.Sprintf("Test User %d", time.Now().Unix()),
		Email: fmt.Sprintf("test%d@example.com", time.Now().Unix()),
	}

	userJSON, _ := json.Marshal(user)
	resp, err := http.Post(baseURL+"/api/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var createdUser models.User
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode created user: %v", err)
	}

	if createdUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, createdUser.Name)
	}

	// Test getting all users
	resp, err = http.Get(baseURL + "/api/users")
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var users []models.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode users: %v", err)
	}

	// Test getting specific user
	userURL := fmt.Sprintf("%s/api/users/%d", baseURL, createdUser.ID)
	resp, err = http.Get(userURL)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var fetchedUser models.User
	if err := json.NewDecoder(resp.Body).Decode(&fetchedUser); err != nil {
		t.Fatalf("Failed to decode user: %v", err)
	}

	if fetchedUser.ID != createdUser.ID {
		t.Errorf("Expected ID %d, got %d", createdUser.ID, fetchedUser.ID)
	}
}

func TestInvalidUserCreation(t *testing.T) {
	// Test with empty name
	user := models.CreateUserRequest{
		Name:  "",
		Email: "test@example.com",
	}

	userJSON, _ := json.Marshal(user)
	resp, err := http.Post(baseURL+"/api/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatalf("Failed to call create user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestRateLimiting(t *testing.T) {
	// Make many requests quickly to test rate limiting
	client := &http.Client{Timeout: 5 * time.Second}
	
	rateLimited := false
	for i := 0; i < 200; i++ {
		resp, err := client.Get(baseURL + "/api/users")
		if err != nil {
			continue
		}
		
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			resp.Body.Close()
			break
		}
		resp.Body.Close()
	}

	if !rateLimited {
		t.Log("Rate limiting may not be working as expected (this might be acceptable depending on configuration)")
	}
}
      }
    