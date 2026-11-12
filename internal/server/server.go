package server

import (
	"context"
	"fmt"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/adapters/handler"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/adapters/repository"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/config"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/migrations"
	"github.com/ZinollinIlyas/tarkov_oauth/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
}

func New(cfg *config.Config) (*Server, error) {
	db, err := gorm.Open(postgres.Open(cfg.Postgres.DSN()), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	if err := migrations.Up(context.Background(), sqlDB); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.TTL)
	authHandler := handler.NewAuthHandler(authService)

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	authHandler.Register(api)

	return &Server{cfg: cfg, router: router}, nil
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.cfg.HTTP.Port)
}
