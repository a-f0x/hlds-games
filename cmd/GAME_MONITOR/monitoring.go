package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hlds-games/internal/api"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/config"
	"hlds-games/internal/managment"
	"log"
)

func main() {
	//status, err := api.GetServerInfo("127.0.0.1", 8090)(context.TODO())
	//if err != nil {
	//	log.Fatalf(fmt.Sprintf("fail to get server status. %s"), err.Error())
	//}
	//log.Printf("status = %v", status)

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
	heartBeatChannel, actionChannel, err := managment.Collect(context.TODO(), client)
	if err != nil {
		log.Fatalf(fmt.Sprintf("%s", err.Error()))
	}
	for {
		select {
		case heartBeat := <-heartBeatChannel:
			message, _ := json.Marshal(heartBeat)
			log.Printf("heartBeat, %s", string(message))
			status, clientError := api.GetServerInfo("127.0.0.1", 8090)(context.TODO())
			if clientError != nil {
				log.Fatalf(fmt.Sprintf("fail to get server status. %s", clientError.Error()))
			}
			log.Printf("status: %v", status)
			return
		case action := <-actionChannel:
			message, _ := json.Marshal(action)
			log.Printf("action, %s", string(message))
		}
	}
}
