package common

import (
	"log"
	"os"
)

func StringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func GetEnv(key string) *string {
	_, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Env %s not set\n", key)
	} else {
		value := os.Getenv(key)
		return &value
	}
	return nil
}
