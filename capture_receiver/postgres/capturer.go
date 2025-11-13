package postgres

import "github.com/nrf24l01/go-web-utils/pg_kit"

type Capturer struct {
	pg_kit.BaseModel
	Name 	 string `json:"name" pg_kit:"unique;index"`
	ApiKey   string `json:"api_key" pg_kit:"unique;index"`
	Enabled  bool   `json:"enabled" pg_kit:"default:true;index"`
}