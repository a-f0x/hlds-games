package main

import (
	"encoding/json"
	"hlds-games/internal/common"
	"hlds-games/internal/eventcollector"
	"log"
	"strconv"
)

func main() {
	common.FakeEnv()
	amqpPort, err := strconv.ParseInt(*common.GetEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}

	client := common.NewAmqpClient(
		*common.GetEnv("RABBITMQ_HOST"),
		amqpPort, *common.GetEnv("RABBITMQ_USER"),
		*common.GetEnv("RABBITMQ_PASSWORD"),
		2,
	)

	ec := eventcollector.NewEventCollector(
		client,
	)
	heartBeatChannel, actionChannel := ec.Collect()
	for {
		select {

		case heartBeat := <-heartBeatChannel:
			message, _ := json.Marshal(heartBeat)
			log.Printf("heartBeat, %s", string(message))

		case action := <-actionChannel:
			message, _ := json.Marshal(action)
			log.Printf("action, %s", string(message))
		}
	}
}
