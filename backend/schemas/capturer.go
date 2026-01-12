package schemas

type Capturer struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	ApiKey  string `json:"api_key"`
	Enabled bool   `json:"enabled"`
}

type CapturerCreateRequest struct {
	Name    string `json:"name" validate:"required,min=3,max=100"`
	Enabled bool   `json:"enabled" validate:"required"`
}

type CapturerCreateResponse struct {
	UUID   string `json:"uuid"`
	ApiKey string `json:"api_key"`
}

type CapturerUpdateRequest struct {
	Name    *string `json:"name" validate:"omitempty,min=3,max=100"`
	Enabled *bool   `json:"enabled" validate:"omitempty"`
}