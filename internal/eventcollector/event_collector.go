package eventcollector

import (
	"context"
	"encoding/json"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/messages"
	"log"
)

type EventCollector struct {
	amqpClient *rabbit.AmqpConsumer
}

func NewEventCollector(
	client *rabbit.AmqpConsumer,
) *EventCollector {
	return &EventCollector{
		amqpClient: client,
	}
}

func (ec *EventCollector) Collect(ctx context.Context) (
	<-chan messages.Message[messages.HeartBeatMessagePayload],
	<-chan messages.Message[messages.ActionMessagePayload],
	error,
) {
	heartBeatChannel := make(chan messages.Message[messages.HeartBeatMessagePayload])
	actionChannel := make(chan messages.Message[messages.ActionMessagePayload])
	err := ec.collect(ctx, heartBeatChannel, actionChannel)
	if err != nil {
		return nil, nil, err
	}
	return heartBeatChannel, actionChannel, nil
}
func (ec *EventCollector) collect(ctx context.Context,
	heartBeatChannel chan messages.Message[messages.HeartBeatMessagePayload],
	actionChannel chan messages.Message[messages.ActionMessagePayload],
) error {
	heartBeatBytesChannel, err := ec.amqpClient.Subscribe(rabbit.HeartBeatQueue)
	if err != nil {
		return err
	}
	actionBeatBytesChannel, err2 := ec.amqpClient.Subscribe(rabbit.GameEventsQueue)
	if err2 != nil {
		return err2
	}
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
	return nil
}
