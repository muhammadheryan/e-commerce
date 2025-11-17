package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	orderapp "github.com/muhammadheryan/e-commerce/application/order"
	productapp "github.com/muhammadheryan/e-commerce/application/product"
	userapp "github.com/muhammadheryan/e-commerce/application/user"
	warehouseapp "github.com/muhammadheryan/e-commerce/application/warehouse"
	"github.com/muhammadheryan/e-commerce/cmd/config"
	redisclient "github.com/muhammadheryan/e-commerce/cmd/redis"
	_ "github.com/muhammadheryan/e-commerce/docs"
	orderRepo "github.com/muhammadheryan/e-commerce/repository/order"
	productRepo "github.com/muhammadheryan/e-commerce/repository/product"
	redisRepo "github.com/muhammadheryan/e-commerce/repository/redis"
	txRepo "github.com/muhammadheryan/e-commerce/repository/tx"
	userRepo "github.com/muhammadheryan/e-commerce/repository/user"
	warehouse "github.com/muhammadheryan/e-commerce/repository/warehouse"
	"github.com/muhammadheryan/e-commerce/thirdparty/rabbitmq"
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
// @description Enter the token with the `Bearer` prefix, e.g: "Bearer <your_token>"

// @securityDefinitions.apikey InternalAPIKey
// @in header
// @name Authorization
// @description Enter the internal API key with the `Bearer` prefix, e.g: "Bearer <your_internal_api_key>"
func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize global logger
	if err := logger.Init(cfg.Environment); err != nil {
		// fallback to standard log if zap init fails
		panic(err)
	}
	defer logger.Close()

	logger.Info(cfg.ProjectName)
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
	RedisRepo := redisRepo.NewRedisRepository()
	ProductRepo := productRepo.NewProductRepository(db)
	OrderRepo := orderRepo.NewOrderRepository(db)
	txRepo := txRepo.NewTxRepository(db)
	warehouseRepo := warehouse.NewWarehouseRepository(db)

	// Initialize RabbitMQ publisher
	publisher, err := rabbitmq.NewPublisher(
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
	)
	if err != nil {
		logger.Fatal("failed to connect rabbitmq publisher", zap.Error(err))
	}
	defer func() {
		_ = publisher.Close()
	}()

	// Initialize RabbitMQ consumer
	consumer, err := rabbitmq.NewConsumer(
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		"http://localhost:"+cfg.Server.Port,
		cfg.InternalAPIKey,
	)
	if err != nil {
		logger.Fatal("failed to connect rabbitmq consumer", zap.Error(err))
	}
	defer func() {
		_ = consumer.Close()
	}()

	// Start consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		logger.Fatal("failed to start rabbitmq consumer", zap.Error(err))
	}

	// Initialize application layers
	UserApp := userapp.NewUserApp(cfg, UserRepo, RedisRepo)
	ProductApp := productapp.NewProductApp(ProductRepo)
	OrderApp := orderapp.NewOrderApp(cfg, txRepo, OrderRepo, warehouseRepo, publisher)
	WarehouseApp := warehouseapp.NewWarehouseApp(txRepo, warehouseRepo)

	httpTransport := transport.NewTransport(UserApp, ProductApp, OrderApp, WarehouseApp, cfg.InternalAPIKey)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      httpTransport,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown handling
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("Shutting down server...")
		cancel()
		if err := server.Close(); err != nil {
			logger.Error("Server close error", zap.Error(err))
		}
	}()

	logger.Info("HTTP server running", zap.String("port", cfg.Server.Port))
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal("failed server", zap.Error(err))
	}
}
