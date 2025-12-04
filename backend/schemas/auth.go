package schemas

import (
	"net/http"

	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type JwtResponse struct {
	Token string `json:"token"`
}

var DefaultNoTokenResponse = echokitSchemas.ErrorResponse{
	Message: "No token provided",
	Code:    http.StatusUnauthorized,
}

var DefaultInvalidTokenResponse = echokitSchemas.ErrorResponse{
	Message: "Invalid token",
	Code:    http.StatusUnauthorized,
}
