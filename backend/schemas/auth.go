package schemas

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type JwtResponse struct {
	Token string `json:"token"`
}