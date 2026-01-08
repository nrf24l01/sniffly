package launch

import (
	"fmt"
	"log"
	"syscall"

	"golang.org/x/term"

	"github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/backend/core"
	"github.com/nrf24l01/sniffly/backend/postgres"
	"gorm.io/gorm"
)

func Dispatch(config *core.Config, db *gorm.DB, rdb *redis.RedisClient, args []string) {
	if len(args) == 0 {
		log.Printf("NO COMMAND LINE ARGUMENTS, RUN BACKEND")
		startBackend(config, db, rdb)
	}
	switch args[0] {
	case "create_user":
		if len(args) < 2 {
			log.Printf("Missing username for create_user")
			return
		}
		username := args[1]

		// Password enter
		fmt.Print("Enter password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		password := string(bytePassword)

		fmt.Print("Confirm password: ")
		bytePasswordConfirm, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		password_confirm := string(bytePasswordConfirm)
		if err != nil {
			log.Printf("Error reading password: %v", err)
			return
		}
		if password != password_confirm {
			log.Printf("Passwords do not match")
			return
		}

		// User create
		user := postgres.User{
			Username: username,
		}
		if err :=  user.SetPassword(password, config.Argon2idConfig); err != nil {
			log.Printf("Error setting password: %v", err)
			return
		}

		// User push to db
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Error creating user: %v", err)
			return
		}

		//Finish
		log.Printf("User %s created successfully", username)
		return
	case "run":
		startBackend(config, db, rdb)
	default:
		log.Printf("Unknown command: %s", args[0])
	}
}