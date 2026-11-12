package handler

import (
	"errors"
	"net/http"

	"github.com/ZinollinIlyas/tarkov_oauth/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth ports.AuthService
}

func NewAuthHandler(auth ports.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(r gin.IRouter) {
	auth := r.Group("/auth")
	auth.POST("/sign-up", h.signUp)
	auth.POST("/sign-in", h.signIn)

	users := r.Group("/users", h.RequireAuth())
	users.GET("/me", h.me)
	users.POST("/lookup", h.lookup)
}

type signUpRequest struct {
	Nickname    string `json:"nickname"     binding:"required,min=3,max=100"`
	DisplayName string `json:"display_name" binding:"max=100"`
	Password    string `json:"password"     binding:"required,min=8,max=72"`
}

type signInRequest struct {
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresAt   int64  `json:"expires_at"`
}

func (h *AuthHandler) signUp(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.auth.SignUp(c.Request.Context(), ports.SignUpInput{
		Nickname:    req.Nickname,
		DisplayName: req.DisplayName,
		Password:    req.Password,
	})
	if err != nil {
		if errors.Is(err, ports.ErrNicknameTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "nickname already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) signIn(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.auth.SignIn(c.Request.Context(), ports.SignInInput{
		Nickname: req.Nickname,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid nickname or password"})
		case errors.Is(err, ports.ErrUserDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": "user is disabled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   token.ExpiresAt.Unix(),
	})
}
