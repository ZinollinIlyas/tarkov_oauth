package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ZinollinIlyas/tarkov_oauth/internal/core/model"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/core/ports"
)

const ctxUser = "auth_user"

// publicUser — то, что можно отдавать о пользователе другим сервисам и игрокам.
type publicUser struct {
	ID          uuid.UUID `json:"id"`
	Nickname    string    `json:"nickname"`
	DisplayName string    `json:"display_name"`
}

func toPublic(u *model.User) publicUser {
	return publicUser{ID: u.ID, Nickname: u.Nickname, DisplayName: u.DisplayName}
}

// RequireAuth проверяет "Authorization: Bearer <access_token>".
func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		user, err := h.auth.Authenticate(c.Request.Context(), raw)
		switch {
		case errors.Is(err, ports.ErrInvalidToken):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		case errors.Is(err, ports.ErrUserDisabled):
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is disabled"})
			return
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.Set(ctxUser, user)
		c.Next()
	}
}

// GET /users/me — владелец токена.
func (h *AuthHandler) me(c *gin.Context) {
	user := c.MustGet(ctxUser).(*model.User)
	c.JSON(http.StatusOK, toPublic(user))
}

type lookupRequest struct {
	IDs []uuid.UUID `json:"ids" binding:"required,max=200"`
}

// POST /users/lookup {"ids": [...]} — публичные данные пользователей по id.
// Неизвестные id просто отсутствуют в ответе.
func (h *AuthHandler) lookup(c *gin.Context) {
	var req lookupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := h.auth.UsersByIDs(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	out := make([]publicUser, 0, len(users))
	for i := range users {
		out = append(out, toPublic(&users[i]))
	}
	c.JSON(http.StatusOK, out)
}
