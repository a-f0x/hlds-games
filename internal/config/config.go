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

func GetRabbitConfig() *RabbitConfig {
	return &RabbitConfig{
		RabbitHost:     *common.GetEnv("RABBITMQ_HOST"),
		RabbitPort:     common.GetEnvInt64Value("RABBITMQ_PORT"),
		RabbitUser:     *common.GetEnv("RABBITMQ_USER"),
		RabbitPassword: *common.GetEnv("RABBITMQ_PASSWORD"),
	}
}
func GetHldsGameConfig() *HldsGameConfig {
	return &HldsGameConfig{
		GameType:        *common.GetEnv("GAME_TYPE"),
		RconPassword:    *common.GetEnv("RCON_PASSWORD"),
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
