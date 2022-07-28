package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hlds-games/internal/api"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/eventcollector"
	"log"
	"strconv"
)

func main() {
	status, err := api.GetServerInfo("127.0.0.1", 8090)(context.TODO())
	if err != nil {
		log.Fatalf(fmt.Sprintf("fail to get server status. %s"), err.Error())
	}
	log.Printf("status = %v", status)

}
func monitoring() {
	common.FakeEnvRabbit("127.0.0.1")
	amqpPort, err := strconv.ParseInt(*common.GetEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}

	client := rabbit.NewAmqpConsumer(
		*common.GetEnv("RABBITMQ_HOST"),
		amqpPort, *common.GetEnv("RABBITMQ_USER"),
		*common.GetEnv("RABBITMQ_PASSWORD"),
		2,
	)

	ec := eventcollector.NewEventCollector(client)
	heartBeatChannel, actionChannel, error := ec.Collect(context.TODO())
	if error != nil {
		log.Fatalf(fmt.Sprintf("%s"), error.Error())
	}
	for {
		select {
		case heartBeat := <-heartBeatChannel:
			message, _ := json.Marshal(heartBeat)
			log.Printf("heartBeat, %s", string(message))
			api.GetServerInfo("127.0.0.1", 8090)(context.Background())
			return
		case action := <-actionChannel:
			message, _ := json.Marshal(action)
			log.Printf("action, %s", string(message))
		}
	}
}
