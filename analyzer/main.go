package analyzer

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nrf24l01/sniffly/analyzer/core"
)

func main() {
	if os.Getenv("PRODUCTION_ENV") != "true" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("failed to load .env: %v", err)
		}
	}

	cfg := core.BuildConfigFromEnv()
}