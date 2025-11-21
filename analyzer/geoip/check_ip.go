package geoip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	redis "github.com/go-redis/redis/v8"
	redisutil "github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/analyzer/core"
)

type connection struct {
	ASN         *string `json:"asn"`
	ISP         *string `json:"isp"`
	ORG 	    *string `json:"org"`
	Domain	    *string `json:"domain"`
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
	city, err := rdb.Client.Get(rdb.Ctx, cfg.GeoIPCacheKeyPrefix+ip+"-city").Result()
	company, err := rdb.Client.Get(rdb.Ctx, cfg.GeoIPCacheKeyPrefix+ip+"-company").Result()
	if err != nil {
		if err == redis.Nil {
			city, company, err := getCityCompanyFromIP(ip)
			if err != nil {
				return "", "", err
			}
			err = rdb.Client.Set(rdb.Ctx, cfg.GeoIPCacheKeyPrefix+ip+"-city", city, time.Duration(cfg.GeoIPCacheTTL)*time.Second).Err()
			if err != nil {
				return "",  "", err
			}
			err = rdb.Client.Set(rdb.Ctx, cfg.GeoIPCacheKeyPrefix+ip+"-company", company, time.Duration(cfg.GeoIPCacheTTL)*time.Second).Err()
			if err != nil {
				return "",  "", err
			}
			return city, company, nil
		} else {
			return "", "", err
		}
	}
	return city, company, nil
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
		if payload.Message != nil && *payload.Message == "" {
			return "", "", fmt.Errorf("ipwho.is lookup failed")
		}
		return "", "", fmt.Errorf("ipwho.is: %s", *payload.Message)
	}
	
	var city, company string

	if payload.Connection != nil {
		if payload.Connection.ORG != nil {
			company = *payload.Connection.ORG
		} else if payload.Connection.ISP != nil {
			company = *payload.Connection.ISP
		}
	}

	if payload.City == nil {
		city = ""
	}

	return city, company, nil
}