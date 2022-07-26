package main

import (
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/launcher"
	"log"
	"strconv"
)

func main() {
	hldsServerPort, err := strconv.ParseInt(*common.GetEnv("PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid game PORT %s", err.Error())
	}
	ga := launcher.NewLauncher(
		hldsServerPort,
		*common.GetEnv("MAP"),
		*common.GetEnv("GAME_TYPE"),
		*common.GetEnv("RCON_PASSWORD"),
	)
	gameEventSender := getGameEventSender()
	heartBeatChannel, actionChannel := ga.RunGame()
	for {
		select {
		case heartBeat := <-heartBeatChannel:
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
	amqpPort, err := strconv.ParseInt(*common.GetEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}
	client := rabbit.NewAmqpProducer(
		*common.GetEnv("RABBITMQ_HOST"),
		amqpPort, *common.GetEnv("RABBITMQ_USER"),
		*common.GetEnv("RABBITMQ_PASSWORD"),
		2,
	)
	return launcher.NewAmqpGameEventSender(client)
}
