package common

import (
	"log"
	"os"
	"strconv"
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
func GetEnvInt64Value(value string) int64 {
	int64Val, err := strconv.ParseInt(*GetEnv(value), 10, 64)
	if err != nil {
		log.Fatalf("Invalid env %s. %s", value, err.Error())
	}
	return int64Val
}

func FakeEnvRabbit(host string) {
	os.Setenv("RABBITMQ_HOST", host)
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guestUsr")
	os.Setenv("RABBITMQ_PASSWORD", "guestPwd")
}
func FakeEnvGameCfg() {
	os.Setenv("GAME_TYPE", "FAKE_GAME")
	os.Setenv("RCON_PASSWORD", "asjop2340239857uG")
	os.Setenv("PORT", "27017")
	os.Setenv("GRPC_API_PORT", "2020")

}
