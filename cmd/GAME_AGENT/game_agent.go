package main

import (
	"hlds-games/internal/common"
	"hlds-games/internal/launcher"
	"hlds-games/internal/messages"
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
		make(chan *messages.Message[messages.HeartBeatMessagePayload]),
		make(chan *messages.Message[messages.ActionMessagePayload]),
	)
	var gameEventSender launcher.GameEventSender
	gameEventSender = getGameEventSender()
	ga.RunGame()
	for {
		select {
		case heartBeat := <-ga.HeartBeat:
			err := gameEventSender.SendHeartBeat(*heartBeat)
			if err != nil {
				log.Printf("Error send heart beat notification. %s", err)
			}
		case action := <-ga.Action:
			err := gameEventSender.SendGameEvent(*action)
			if err != nil {
				log.Printf("Error send action notification. %s", err)
			}
		}
	}

}
func getGameEventSender() launcher.GameEventSender {
	amqpPort, err := strconv.ParseInt(*common.GetEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}
	return launcher.NewAmqpGameEventSender(
		common.NewAmqpClient(
			*common.GetEnv("RABBITMQ_HOST"),
			amqpPort, *common.GetEnv("RABBITMQ_USER"),
			*common.GetEnv("RABBITMQ_PASSWORD"),
		),
	)
}
