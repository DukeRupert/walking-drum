// handlers/user_handler.go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// Request/Response structs

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	Name     *string `json:"name,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type UserResponse struct {
	ID               uuid.UUID              `json:"id"`
	Email            string                 `json:"email"`
	Name             string                 `json:"name"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	StripeCustomerID *string                `json:"stripe_customer_id,omitempty"`
	IsActive         bool                   `json:"is_active"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// Handlers

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "UserHandler").
		Str("method", "CreateUser").
		Logger()

	logger.Debug().Msg("Processing user creation request")

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Mask the email in logs for privacy
	maskedEmail := h.maskEmail(req.Email)
	logger.Debug().
		Str("email", maskedEmail).
		Str("name", req.Name).
		Msg("Received user creation request")

	// Basic validation
	if req.Email == "" || req.Password == "" || req.Name == "" {
		logger.Error().
			Str("email", maskedEmail).
			Bool("passwordProvided", req.Password != "").
			Bool("nameProvided", req.Name != "").
			Msg("Validation failed: missing required fields")
		http.Error(w, "Email, password, and name are required", http.StatusBadRequest)
		return
	}

	// Hash the password
	logger.Debug().Str("email", maskedEmail).Msg("Hashing password")
	passwordHash, err := h.hashPassword(req.Password)
	if err != nil {
		logger.Error().
			Err(err).
			Str("email", maskedEmail).
			Msg("Failed to hash password")
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
		IsActive:     true,
	}

	logger.Debug().
		Str("email", maskedEmail).
		Str("name", req.Name).
		Msg("Creating user in database")

	err = h.userRepo.Create(r.Context(), user)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			logger.Error().
				Err(err).
				Str("email", maskedEmail).
				Msg("User with this email already exists")
			http.Error(w, "User with this email already exists", http.StatusConflict)
			return
		}
		logger.Error().
			Err(err).
			Str("email", maskedEmail).
			Msg("Failed to create user in database")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("email", maskedEmail).
		Str("userId", user.ID.String()).
		Msg("User created successfully, generating response")

	response, err := h.modelToResponse(user)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userId", user.ID.String()).
			Str("email", maskedEmail).
			Msg("Failed to generate user response")
		http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("userId", user.ID.String()).
		Str("email", maskedEmail).
		Msg("User created successfully")
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "UserHandler").
		Str("method", "GetUser").
		Logger()

	vars := mux.Vars(r)
	userIDStr := vars["id"]
	
	logger.Debug().Str("userID", userIDStr).Msg("Processing user request")
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error().Err(err).Str("userID", userIDStr).Msg("Invalid user ID format")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	logger.Debug().Str("userID", userID.String()).Msg("Fetching user from repository")
	
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			logger.Info().
				Err(err).
				Str("userID", userID.String()).
				Msg("User not found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("userID", userID.String()).
			Msg("Failed to get user from repository")
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("userID", userID.String()).
		Str("email", user.Email). // Assuming user has an Email field
		Msg("User found, generating response")

	response, err := h.modelToResponse(user)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userID", userID.String()).
			Msg("Failed to generate user response")
		http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	
	logger.Info().
		Str("userID", userID.String()).
		Msg("User returned successfully")
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the existing user
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}

	if req.Password != nil {
		passwordHash, err := h.hashPassword(*req.Password)
		if err != nil {
			http.Error(w, "Failed to process password", http.StatusInternalServerError)
			return
		}
		user.PasswordHash = passwordHash
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Perform the update
	err = h.userRepo.Update(r.Context(), user)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			http.Error(w, "Email is already in use", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(user)
	if err != nil {
		http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "UserHandler").
		Str("method", "DeleteUser").
		Logger()

	vars := mux.Vars(r)
	idStr := vars["id"]
	
	logger.Debug().Str("userId", idStr).Msg("Processing user deletion request")

	userID, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error().Err(err).Str("userId", idStr).Msg("Invalid user ID format")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("userId", userID.String()).
		Msg("Deleting user from database")

	err = h.userRepo.Delete(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			logger.Error().
				Err(err).
				Str("userId", userID.String()).
				Msg("User not found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("userId", userID.String()).
			Msg("Failed to delete user from database")
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("userId", userID.String()).
		Msg("User deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "UserHandler").
		Str("method", "ListUsers").
		Logger()

	logger.Debug().Msg("Processing list users request")

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default
	offset := 0 // Default

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else if err != nil {
			logger.Debug().Err(err).Str("limit", limitStr).Msg("Invalid limit parameter")
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		} else if err != nil {
			logger.Debug().Err(err).Str("offset", offsetStr).Msg("Invalid offset parameter")
		}
	}

	logger.Debug().
		Int("limit", limit).
		Int("offset", offset).
		Msg("Retrieving users from database")

	users, err := h.userRepo.List(r.Context(), limit, offset)
	if err != nil {
		logger.Error().
			Err(err).
			Int("limit", limit).
			Int("offset", offset).
			Msg("Failed to list users from database")
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}
	
	logger.Debug().
		Int("count", len(users)).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Successfully retrieved users, generating response")

	// Convert to response objects
	var responses []UserResponse
	for _, user := range users {
		response, err := h.modelToResponse(user)
		if err != nil {
			logger.Error().
				Err(err).
				Str("userId", user.ID.String()).
				Msg("Failed to generate user response")
			http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Int("count", len(responses)).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Users returned successfully")
}

// Helper functions

func (h *UserHandler) modelToResponse(user *models.User) (UserResponse, error) {
    response := UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: user.CreatedAt.Format(http.TimeFormat),
        UpdatedAt: user.UpdatedAt.Format(http.TimeFormat),
        IsActive:  user.IsActive,
    }

    if user.StripeCustomerID != nil {
        response.StripeCustomerID = user.StripeCustomerID
    }

    if user.Metadata != nil {
        response.Metadata = *user.Metadata
    }

    return response, nil
}

func (h *UserHandler) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Helper function to mask emails for privacy in logs
func (h *UserHandler) maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		// Not a valid email, return safely masked value
		return "invalid-email-format"
	}
	
	username := parts[0]
	domain := parts[1]
	
	// Mask username part: show first and last character if username is 3+ chars
	if len(username) > 2 {
		masked := string(username[0]) + strings.Repeat("*", len(username)-2) + string(username[len(username)-1])
		return masked + "@" + domain
	} else if len(username) > 0 {
		// For very short usernames, just show one char and asterisks
		return string(username[0]) + "*@" + domain
	}
	
	// Fallback for empty username
	return "***@" + domain
}
