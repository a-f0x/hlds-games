package main

import (
	"hlds-games/internal/api"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/config"
	"hlds-games/internal/launcher"
	"hlds-games/internal/rcon"
	"log"
)

func main() {
	hldsGameConfig := config.GetHldsGameConfig()
	rc := rcon.NewRcon(hldsGameConfig.Host, hldsGameConfig.HldsGamePort, hldsGameConfig.RconPassword)

	grpcApiConfig := config.GetGrpcApiConfig()
	apiServer := api.NewHLDSApiServer(grpcApiConfig, rc)
	go apiServer.RunServer()

	ga := launcher.NewLauncher(hldsGameConfig)
	heartBeatChannel, actionChannel := ga.RunGame(common.GetRequiredEnv("MAP"))
	gameEventSender := getGameEventSender()
	for {
		select {
		case heartBeat := <-heartBeatChannel:
			heartBeat.Payload.ApiHost = heartBeat.Payload.GameHost
			heartBeat.Payload.ApiPort = grpcApiConfig.GrpcApiPort
			err := gameEventSender.SendHeartBeat(heartBeat)
			if err != nil {
				log.Printf("Error send heart beat notification. %s", err)
			}
		case action := <-actionChannel:
			err := gameEventSender.SendGameEvent(action)
			if err != nil {
				log.Printf("Error send action notification. %s", err)
			}
		}
	}
}
func getGameEventSender() *launcher.AmqpGameEventSender {
	amqpConfig := config.GetRabbitConfig()
	client := rabbit.NewAmqpProducer(
		amqpConfig.RabbitHost,
		amqpConfig.RabbitPort,
		amqpConfig.RabbitUser,
		amqpConfig.RabbitPassword,
		2,
	)
	return launcher.NewAmqpGameEventSender(client)
}
