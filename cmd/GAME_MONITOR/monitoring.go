package main

import (
	"encoding/json"
	"hlds-games/internal/common"
	"hlds-games/internal/eventcollector"
	"log"
	"os"
	"strconv"
)

func fakeEnv() {
	os.Setenv("RABBITMQ_HOST", "192.168.88.44")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guestUsr")
	os.Setenv("RABBITMQ_PASSWORD", "guestPwd")

}
func main() {
	fakeEnv()
	amqpPort, err := strconv.ParseInt(*common.GetEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}

	client := common.NewAmqpClient(
		*common.GetEnv("RABBITMQ_HOST"),
		amqpPort, *common.GetEnv("RABBITMQ_USER"),
		*common.GetEnv("RABBITMQ_PASSWORD"),
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
