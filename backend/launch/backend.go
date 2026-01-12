package launch

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nrf24l01/sniffly/backend/core"

	echokitMw "github.com/nrf24l01/go-web-utils/echokit/middleware"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	random "github.com/nrf24l01/go-web-utils/misc/random"
	"github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/routes"

	"log"
	"os"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)

func startBackend(config *core.Config, db *gorm.DB, rdb *redis.RedisClient, randomGenerator *random.RandomGenerator) {
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
	handler := &handlers.Handler{DB: db, Config: config, RDB: rdb, RandomGenerator: randomGenerator}
	routes.RegisterRoutes(e, handler)

	// Start server
	e.Logger.Fatal(e.Start(config.WebAppConfig.AppHost))
}