package handlers

import (
	"net/http"
	"time"

	"github.com/HarshChauhan626/AegisIDP/backend/auth"
	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication routes.
type AuthHandler struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(userRepo repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret}
}

// loginRequest is the expected body for POST /api/auth/login.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Login authenticates a user and returns a JWT token pair.
//
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse("INVALID_CREDENTIALS", "Invalid email or password"))
		return
	}

	if !user.Active {
		c.JSON(http.StatusForbidden, errorResponse("ACCOUNT_DISABLED", "Account has been disabled"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse("INVALID_CREDENTIALS", "Invalid email or password"))
		return
	}

	pair, err := auth.GenerateTokenPair(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("TOKEN_ERROR", "Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"user":          userResponse(user),
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
	}))
}

// refreshRequest is the expected body for POST /api/auth/refresh.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh issues a new access token given a valid refresh token.
//
// POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse("TOKEN_INVALID", "Invalid or expired refresh token"))
		return
	}

	pair, err := auth.GenerateTokenPair(claims.UserID, claims.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("TOKEN_ERROR", "Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, successResponse(gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
	}))
}

// SeedAdminUser creates the default admin user on first run if no users exist.
func SeedAdminUser(userRepo repository.UserRepository) error {
	users, err := userRepo.List()
	if err != nil || len(users) > 0 {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		ID:           uuid.NewString(),
		Email:        "admin@platform.local",
		PasswordHash: string(hash),
		Name:         "Platform Admin",
		Role:         models.RoleAdmin,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return userRepo.Create(admin)
}

// userResponse strips sensitive fields from User.
func userResponse(u *models.User) gin.H {
	return gin.H{
		"id":         u.ID,
		"email":      u.Email,
		"name":       u.Name,
		"role":       u.Role,
		"active":     u.Active,
		"created_at": u.CreatedAt,
	}
}
