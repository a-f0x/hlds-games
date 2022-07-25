package main

import (
	"hlds-games/internal/game"
	"hlds-games/pkg/messages"
	"log"
	"os"
	"strconv"
)

func main() {
	getEnv := func(key string) *string {
		_, ok := os.LookupEnv(key)
		if !ok {
			log.Fatalf("Env %s not set\n", key)
		} else {
			value := os.Getenv(key)
			return &value
		}
		return nil
	}
	amqpPort, err := strconv.ParseInt(*getEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}
	hldsServerPort, err := strconv.ParseInt(*getEnv("PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid game PORT %s", err.Error())
	}

	ga := game.NewLauncher(
		hldsServerPort,
		*getEnv("MAP"),
		*getEnv("GAME_TYPE"),
		*getEnv("RCON_PASSWORD"),
		make(chan *messages.Message[messages.HeartBeatMessagePayload]),
		make(chan *messages.Message[messages.ActionMessagePayload]),
	)

	amqpClient := game.NewAmqGameClient(
		*getEnv("RABBITMQ_HOST"),
		amqpPort, *getEnv("RABBITMQ_USER"),
		*getEnv("RABBITMQ_PASSWORD"),
	)

	err = amqpClient.Connect()
	if err != nil {
		log.Fatalf("error connect to ampq %s", err)
	}
	ga.RunGame()
	for {
		select {
		case heartBeat := <-ga.HeartBeat:
			err := amqpClient.SendHeartBeat(*heartBeat)
			if err != nil {
				log.Printf("Error send heart beat notification. %s", err)
			}
		case action := <-ga.Action:
			err := amqpClient.SendGameEvent(*action)
			if err != nil {
				log.Printf("Error send action notification. %s", err)
			}
		}
	}
}
