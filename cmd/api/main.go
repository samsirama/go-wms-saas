package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/samsirama/go-wms-saas/internal/adapters/config"
	"github.com/samsirama/go-wms-saas/internal/adapters/handler"
	"github.com/samsirama/go-wms-saas/internal/adapters/repository"
	"github.com/samsirama/go-wms-saas/internal/core/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	cfg := config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := cfg.Database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Database connected successfully")

	productRepo := repository.NewProductRepo(pool)
	stockRepo := repository.NewStockRepo(pool)
	txManager := repository.NewPostgresDB(pool)

	productSvc := services.NewProductService(productRepo, stockRepo, txManager)
	stockSvc := services.NewStockService(stockRepo, txManager)

	productHandler := handler.NewProductHandler(productSvc)
	stockHandler := handler.NewStockHandler(stockSvc)

	app := fiber.New(fiber.Config{
		AppName:               "WMS SaaS API",
		ServerHeader:          "Fiber",
		StrictRouting:         true,
		CaseSensitive:         true,
		DisableStartupMessage: false,
		BodyLimit:             4 * 1024 * 1024,
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		ErrorHandler:          customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Security.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    "WMS SaaS",
		})
	})

	api := app.Group("/api/v1")

	api.Post("/products", productHandler.Create)
	api.Get("/products", productHandler.List)
	api.Get("/products/:id", productHandler.GetByID)

	api.Post("/stock/reserve", stockHandler.ReserveStock)
	api.Post("/stock/release", stockHandler.ReleaseStock)
	api.Get("/stock/:id", stockHandler.GetStockLevel)
	api.Get("/stock/:id/history", stockHandler.GetMutationHistory)

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s (env: %s)", cfg.Server.Port, cfg.App.Env)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Shutdown complete")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": err.Error(),
		"success": false,
	})
}
