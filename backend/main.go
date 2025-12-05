package main

import (
	"github.com/go-playground/validator/v10"

	echokitMw "github.com/nrf24l01/go-web-utils/echokit/middleware"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	pgKit "github.com/nrf24l01/go-web-utils/pg_kit"
	"github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/backend/core"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/postgres"
	"github.com/nrf24l01/sniffly/backend/routes"

	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)
func main() {
	// Try to load .env file in non-production environment
	if os.Getenv("PRODUCTION_ENV") != "true" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("failed to load .env: %v", err)
		}
	}
	
	// Configuration initialization
	config, err := core.BuildConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to build config: %v", err)
	}

	// Data sources initialization
	db, err := pgKit.RegisterPostgres(config.PGConfig, &postgres.User{})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	rdb := redis.NewRedisClient(config.RedisConfig)

	// Create echo object
	e := echo.New()

	// Register custom validator
	v := validator.New()
	e.Validator = &echokitMw.CustomValidator{Validator: v}

	// Logs
	if os.Getenv("NO_LOGS") != "true" {
		e.Use(echoMw.Logger())
	}

	// Echo Configs
    e.Use(echoMw.Recover())
	log.Printf("Setting allowed origin to: %s", config.WebAppConfig.AllowOrigin)
	e.Use(echoMw.CORSWithConfig(echoMw.CORSConfig{
		AllowOrigins: []string{config.WebAppConfig.AllowOrigin},
		AllowMethods: []string{echo.GET, echo.POST, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// Health check endpoint
	e.GET("/ping", func(c echo.Context) error {
	return c.JSON(200, echokitSchemas.Message{Status: "Sniffly backend is ok"})
	})

	// Register routes
	handler := &handlers.Handler{DB: db, Config: config, RDB: rdb}
	routes.RegisterRoutes(e, handler)
	
	// Start server
	e.Logger.Fatal(e.Start(config.WebAppConfig.AppHost))
}