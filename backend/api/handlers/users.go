package handlers

import (
	"net/http"
	"time"

	"github.com/HarshChauhan626/AegisIDP/backend/middleware"
	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler handles user management routes (Admin only).
type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// List returns all users.
// GET /api/users
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.userRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}
	resp := make([]gin.H, len(users))
	for i, u := range users {
		resp[i] = userResponse(&u)
	}
	c.JSON(http.StatusOK, successResponse(resp))
}

type createUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin developer viewer"`
}

// Create creates a new user (Admin only).
// POST /api/users
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("HASH_ERROR", "Failed to hash password"))
		return
	}

	user := &models.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hash),
		Role:         req.Role,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusConflict, errorResponse("CONFLICT", "Email already exists"))
		return
	}

	c.JSON(http.StatusCreated, successResponse(userResponse(user)))
}

type updateUserRequest struct {
	Role   *string `json:"role" binding:"omitempty,oneof=admin developer viewer"`
	Active *bool   `json:"active"`
	Name   *string `json:"name"`
}

// Update patches a user's role, active status, or name.
// PATCH /api/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	user, err := h.userRepo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "User not found"))
		return
	}

	// Prevent self-demotion
	if c.Param("id") == middleware.GetUserID(c) {
		c.JSON(http.StatusForbidden, errorResponse("FORBIDDEN", "Cannot modify your own account via this endpoint"))
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Active != nil {
		user.Active = *req.Active
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(userResponse(user)))
}
