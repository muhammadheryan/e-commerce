package main

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	userapp "github.com/muhammadheryan/e-commerce/application/user"
	"github.com/muhammadheryan/e-commerce/cmd/config"
	redisclient "github.com/muhammadheryan/e-commerce/cmd/redis"
	_ "github.com/muhammadheryan/e-commerce/docs"
	redisRepo "github.com/muhammadheryan/e-commerce/repository/redis"
	userRepo "github.com/muhammadheryan/e-commerce/repository/user"
	"github.com/muhammadheryan/e-commerce/transport"
	"github.com/muhammadheryan/e-commerce/utils/logger"
	"go.uber.org/zap"
)

// @title E-COMMERCE API
// @version 1.0
// @description E-COMMERCE API Documentation
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize global logger
	if err := logger.Init(cfg.Environment); err != nil {
		// fallback to standard log if zap init fails
		panic(err)
	}
	defer logger.Close()

	logger.Info("Starting server", zap.String("env", cfg.Environment))

	// Connect to database
	db, err := sqlx.Connect("mysql", cfg.GetDSN())
	if err != nil {
		logger.Fatal("err connect db", zap.Error(err))
	}

	// Initialize Redis client
	if err := redisclient.New(cfg); err != nil {
		logger.Fatal("err connect redis", zap.Error(err))
	}
	defer func() {
		_ = redisclient.Close()
	}()

	// Set database connection pool settings
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Initialize repositories
	UserRepo := userRepo.NewUserRepository(db)
	RedisRepo := redisRepo.NewRepository()

	// Initialize application layers
	UserApp := userapp.NewUserApp(cfg, UserRepo, RedisRepo)

	httpTransport := transport.NewTransport(UserApp)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      httpTransport,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	logger.Info("HTTP server running", zap.String("port", cfg.Server.Port))
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal("failed server", zap.Error(err))
	}
}
