package geoip

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	redisutil "github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/analyzer/core"
	redis "github.com/redis/go-redis/v9"
)

type connection struct {
	ASN         interface{} `json:"asn"`
	ISP         *string     `json:"isp"`
	ORG         *string     `json:"org"`
	Domain      *string     `json:"domain"`
}

type answerPayload struct {
	Success     *bool   `json:"success"`
	CountryCode *string `json:"country_code"`
	City        *string `json:"city"`
	Region 	    *string `json:"region"`
	Message     *string `json:"message"`
	Connection  *connection `json:"connection"`
}

func CityCompanyFromIP(ip string, rdb *redisutil.RedisClient, cfg *core.AppConfig) (string, string, error) {
	cityKey := cfg.GeoIPCacheKeyPrefix + ip + "-city"
	compKey := cfg.GeoIPCacheKeyPrefix + ip + "-company"

	city, cityErr := rdb.Client.Get(rdb.Ctx, cityKey).Result()
	company, compErr := rdb.Client.Get(rdb.Ctx, compKey).Result()

	if cityErr == nil && compErr == nil {
		return city, company, nil
	}

	if cityErr != nil && cityErr != redis.Nil {
		return "", "", cityErr
	}
	if compErr != nil && compErr != redis.Nil {
		return "", "", compErr
	}

	// At least one cache entry is missing -> fetch from upstream
	log.Printf("Cache miss for IP %s: cityErr=%v compErr=%v", ip, cityErr, compErr)
	newCity, newCompany, err := getCityCompanyFromIP(ip)
	if err != nil {
		return "", "", err
	}

	if err := rdb.Client.Set(rdb.Ctx, cityKey, newCity, time.Duration(cfg.GeoIPCacheTTL)*time.Second).Err(); err != nil {
		return "", "", err
	}
	if err := rdb.Client.Set(rdb.Ctx, compKey, newCompany, time.Duration(cfg.GeoIPCacheTTL)*time.Second).Err(); err != nil {
		return "", "", err
	}

	return newCity, newCompany, nil
}

func getCityCompanyFromIP(ip string) (string, string, error) {
	url := fmt.Sprintf("https://ipwho.is/%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("ipwho.is returned status %d", resp.StatusCode)
	}

	var payload answerPayload

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", err
	}

	if payload.Success != nil && !*payload.Success {
		if payload.Message != nil {
			if *payload.Message == "Reserved range" {
				return "Local Network", "Local Network", nil
			}
			if *payload.Message == "" {
				return "", "", fmt.Errorf("ipwho.is lookup failed")
			}
			return "", "", fmt.Errorf("ipwho.is: %s", *payload.Message)
		}
		return "", "", fmt.Errorf("ipwho.is lookup failed")
	}
	
	var city, company string

	if payload.Connection != nil {
		company = ""
		if payload.Connection.ORG != nil && *payload.Connection.ORG != "" {
			company = *payload.Connection.ORG
		}
		if payload.Connection.ISP != nil && *payload.Connection.ISP != "" {
			if company != "" {
				company += " (" + *payload.Connection.ISP + ")"
			} else {
				company = *payload.Connection.ISP
			}
		}
		if payload.Connection.Domain != nil && *payload.Connection.Domain != "" {
			if company != "" {
				company += " (" + *payload.Connection.Domain + ")"
			} else {
				company = *payload.Connection.Domain
			}
		}
	}

	if payload.City != nil {
		city = *payload.City
	} else {
		city = ""
	}

	return city, company, nil
}