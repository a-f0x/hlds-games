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

func FakeEnvRabbit(host string) {
	os.Setenv("RABBITMQ_HOST", host)
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guestUsr")
	os.Setenv("RABBITMQ_PASSWORD", "guestPwd")
}
