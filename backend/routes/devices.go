package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/schemas"

	echokitMW "github.com/nrf24l01/go-web-utils/echokit/middleware"
)

func RegisterDeviceRoutes(e *echo.Echo, h *handlers.Handler) {
	group := e.Group("/devices")
	group.Use(echokitMW.JWTMiddleware(*h.Config.JWTConfig))

	group.GET("", h.GetDevicesHandler)
	group.GET("/block-rules", h.ListDeviceBlockRulesHandler)
	group.PATCH("/:id", h.UpdateDeviceLabelHandler, echokitMW.PathUuidV4Middleware("id"), echokitMW.BodyValidationMiddleware(func() interface{} {
		return &schemas.UpdateDeviceLabelRequest{}
	}))
	group.GET("/:id/block-rules", h.ListDeviceBlockRulesForDeviceHandler, echokitMW.PathUuidV4Middleware("id"))
	group.POST("/:id/block-rules", h.CreateDeviceBlockRuleHandler, echokitMW.PathUuidV4Middleware("id"), echokitMW.BodyValidationMiddleware(func() interface{} {
		return &schemas.CreateDeviceBlockRuleRequest{}
	}))
	group.PATCH("/:id/block-rules/:rule_id", h.UpdateDeviceBlockRuleHandler,
		echokitMW.PathUuidV4Middleware("id"),
		echokitMW.PathUuidV4Middleware("rule_id"),
		echokitMW.BodyValidationMiddleware(func() interface{} {
			return &schemas.UpdateDeviceBlockRuleRequest{}
		}),
	)
	group.DELETE("/:id/block-rules/:rule_id", h.DeleteDeviceBlockRuleHandler,
		echokitMW.PathUuidV4Middleware("id"),
		echokitMW.PathUuidV4Middleware("rule_id"),
	)
}
