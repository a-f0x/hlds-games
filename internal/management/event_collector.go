package management

import (
	"context"
	"encoding/json"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/messages"
	"log"
)

//todo разобраться как использовать контекст для отписки и корректного закрытия канала и конекшена ребита

func Collect(ctx context.Context, amqpClient *rabbit.AmqpConsumer) (
	<-chan messages.Message[messages.HeartBeatMessagePayload],
	<-chan messages.Message[messages.ActionMessagePayload],
	error,
) {
	heartBeatChannel := make(chan messages.Message[messages.HeartBeatMessagePayload])
	actionChannel := make(chan messages.Message[messages.ActionMessagePayload])
	err := collect(ctx, amqpClient, heartBeatChannel, actionChannel)
	if err != nil {
		return nil, nil, err
	}
	return heartBeatChannel, actionChannel, nil
}
func collect(
	ctx context.Context,
	amqpClient *rabbit.AmqpConsumer,
	heartBeatChannel chan messages.Message[messages.HeartBeatMessagePayload],
	actionChannel chan messages.Message[messages.ActionMessagePayload],
) error {
	heartBeatBytesChannel, err := amqpClient.Subscribe(ctx, rabbit.HeartBeatQueue)
	if err != nil {
		return err
	}
	actionBeatBytesChannel, err2 := amqpClient.Subscribe(ctx, rabbit.GameEventsQueue)
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
