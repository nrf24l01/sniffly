package core

import random "github.com/nrf24l01/go-web-utils/misc/random"

func GenerateApiKey(randomGenerator *random.RandomGenerator) string {
	return randomGenerator.RandomString(32)
}