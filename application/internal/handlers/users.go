package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"octa-byte-ai/internal/models"
	"github.com/gorilla/mux"
)

func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	correlationID := r.Context().Value("correlation_id").(string)

	h.logger.Info("Getting users", "correlation_id", correlationID)

	query := `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		h.logger.Error("Failed to query users", "error", err, "correlation_id", correlationID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			h.logger.Error("Failed to scan user", "error", err, "correlation_id", correlationID)
			continue
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	correlationID := r.Context().Value("correlation_id").(string)

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err, "correlation_id", correlationID)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Name == "" || req.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Creating user", "name", req.Name, "email", req.Email, "correlation_id", correlationID)

	query := `
		INSERT INTO users (name, email, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, name, email, created_at, updated_at
	`

	var user models.User
	err := h.db.QueryRow(query, req.Name, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err, "correlation_id", correlationID)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	correlationID := r.Context().Value("correlation_id").(string)
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting user", "id", id, "correlation_id", correlationID)

	query := `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`

	var user models.User
	err = h.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get user", "error", err, "correlation_id", correlationID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
