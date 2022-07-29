package config

import "hlds-games/internal/common"

type RabbitConfig struct {
	RabbitHost     string
	RabbitPort     int64
	RabbitUser     string
	RabbitPassword string
}
type HldsGameConfig struct {
	GameType        string
	RconPassword    string
	HldsGamePort    int64
	Host            string
	LogReceiverPort int64
}
type GrpcApiConfig struct {
	GrpcApiPort int64
}
type TelegramProxyConfig struct {
	Enabled bool
	Url     string
}
type TelegramBotConfig struct {
	Token            string
	ReconnectTimeout int64
	AdminPassword    string
}
type TelegramConfig struct {
	Proxy *TelegramProxyConfig
	Bot   *TelegramBotConfig
}

func GetTelegramBotConfig() *TelegramConfig {
	return &TelegramConfig{
		Proxy: &TelegramProxyConfig{
			Enabled: common.GetEnvBoolValue("TELEGRAM_PROXY_ENABLED"),
			Url:     common.GetEnv("TELEGRAM_PROXY_URL"),
		},
		Bot: &TelegramBotConfig{
			Token:            common.GetRequiredEnv("TELEGRAM_BOT_TOKEN"),
			ReconnectTimeout: common.GetEnvInt64Value("TELEGRAM_RECONNECT_TIMEOUT_SEC"),
			AdminPassword:    common.GetRequiredEnv("TELEGRAM_ADMIN_PASSWORD"),
		},
	}

}
func GetRabbitConfig() *RabbitConfig {
	return &RabbitConfig{
		RabbitHost:     common.GetRequiredEnv("RABBITMQ_HOST"),
		RabbitPort:     common.GetEnvInt64Value("RABBITMQ_PORT"),
		RabbitUser:     common.GetRequiredEnv("RABBITMQ_USER"),
		RabbitPassword: common.GetRequiredEnv("RABBITMQ_PASSWORD"),
	}
}
func GetHldsGameConfig() *HldsGameConfig {
	return &HldsGameConfig{
		GameType:        common.GetRequiredEnv("GAME_TYPE"),
		RconPassword:    common.GetRequiredEnv("RCON_PASSWORD"),
		HldsGamePort:    common.GetEnvInt64Value("PORT"),
		Host:            "127.0.0.1",
		LogReceiverPort: 27999,
	}
}

func GetGrpcApiConfig() *GrpcApiConfig {
	return &GrpcApiConfig{
		GrpcApiPort: common.GetEnvInt64Value("GRPC_API_PORT"),
	}
}
