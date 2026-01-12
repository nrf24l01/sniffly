package handlers

import (
	"gorm.io/gorm"

	"github.com/nrf24l01/sniffly/backend/core"

	random "github.com/nrf24l01/go-web-utils/misc/random"
	redisutil "github.com/nrf24l01/go-web-utils/redis"
)

type Handler struct {
	DB *gorm.DB
	Config *core.Config
	RDB *redisutil.RedisClient
	RandomGenerator *random.RandomGenerator
}