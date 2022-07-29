package main

import (
	"context"
	"encoding/json"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/config"
	"hlds-games/internal/management"
	"log"
	"time"
)

func main() {
	monitoring()
}
func monitoring() {
	common.FakeEnvRabbit("127.0.0.1")
	rabbitConfig := config.GetRabbitConfig()
	client := rabbit.NewAmqpConsumer(
		rabbitConfig.RabbitHost,
		rabbitConfig.RabbitPort,
		rabbitConfig.RabbitUser,
		rabbitConfig.RabbitPassword,
		2,
	)
	heartBeatChannel, actionChannel, err := management.Collect(context.TODO(), client)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	gm := management.NewGameManager("192.168.88.61")
	count := 0
	for {
		select {
		case heartBeat := <-heartBeatChannel:
			count++
			message, _ := json.Marshal(heartBeat)
			log.Printf("heartBeat %d, %s", count, string(message))
			gm.RegisterGame(heartBeat)
			log.Printf("Registered games: %v", gm.ListGames())
			if count >= 10 {
				time.Sleep(time.Duration(10) * time.Second)
				log.Printf("Registered games: %v", gm.ListGames())
				return
			}

		case action := <-actionChannel:
			message, _ := json.Marshal(action)
			log.Printf("action, %s", string(message))
		}
	}

}
