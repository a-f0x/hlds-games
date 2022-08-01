package common

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

func StringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func GetRequiredEnv(key string) string {
	_, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Env %s not set\n", key)
	}
	value := os.Getenv(key)
	return value
}
func GetEnv(key string) string {
	return os.Getenv(key)
}

func GetEnvInt64Value(key string) int64 {
	int64Val, err := strconv.ParseInt(GetRequiredEnv(key), 10, 64)
	if err != nil {
		log.Fatalf("Invalid env %s. %s", key, err.Error())
	}
	return int64Val
}
func GetEnvBoolValue(key string) bool {
	boolVal, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		log.Printf("Invalid env %s. %s", key, err.Error())
		return false
	}
	return boolVal
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

func FakeTelegramCfg(token string) {
	os.Setenv("TELEGRAM_BOT_TOKEN", token)
	os.Setenv("TELEGRAM_RECONNECT_TIMEOUT_SEC", "10")
	os.Setenv("TELEGRAM_ADMIN_PASSWORD", "123")
}
