package eventcollector

import (
	"encoding/json"
	"hlds-games/internal/common"
	"hlds-games/internal/messages"
	"log"
)

type EventCollector struct {
	amqpClient *common.AmqpClient
}

func NewEventCollector(
	client *common.AmqpClient,
) *EventCollector {
	return &EventCollector{
		amqpClient: client,
	}
}

func (ec *EventCollector) Collect() (
	<-chan messages.Message[messages.HeartBeatMessagePayload],
	<-chan messages.Message[messages.ActionMessagePayload],
) {
	heartBeatChannel := make(chan messages.Message[messages.HeartBeatMessagePayload])
	actionChannel := make(chan messages.Message[messages.ActionMessagePayload])
	ch := ec.amqpClient.Connect()
	go func() {
		for {
			isConnected := <-ch
			log.Printf("EventCollector: is connected %v", isConnected)
			if isConnected {
				ec.collect(heartBeatChannel, actionChannel)
				continue
			}
		}
	}()
	return heartBeatChannel, actionChannel
}
func (ec *EventCollector) collect(
	heartBeatChannel chan messages.Message[messages.HeartBeatMessagePayload],
	actionChannel chan messages.Message[messages.ActionMessagePayload],
) {
	heartBeatBytesChannel := ec.amqpClient.Stream(common.HeartBeatQueue)
	actionBeatBytesChannel := ec.amqpClient.Stream(common.GameEventsQueue)
	go func() {
		for {
			select {
			case hb := <-heartBeatBytesChannel:
				m := new(messages.Message[messages.HeartBeatMessagePayload])
				err := json.Unmarshal(hb, m)
				if err != nil {
					log.Fatalf("HeartBeatMessagePayload unmarshal error: %s. Source: %s ", err, hb)
				}
				heartBeatChannel <- *m
			case ga := <-actionBeatBytesChannel:
				m := new(messages.Message[messages.ActionMessagePayload])
				err := json.Unmarshal(ga, m)
				if err != nil {
					log.Fatalf("ActionMessagePayload unmarshal error: %s. Source: %s ", err, ga)
				}
				actionChannel <- *m
			}
		}
	}()

}
