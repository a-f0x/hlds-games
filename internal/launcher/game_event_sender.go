package launcher

import (
	"hlds-games/internal/common"
	"hlds-games/internal/messages"
)

type GameEventSender interface {
	SendHeartBeat(message messages.Message[messages.HeartBeatMessagePayload]) error

	SendGameEvent(message messages.Message[messages.ActionMessagePayload]) error
}

const (
	actionExpirationTimeMs    = "15000" //15 sec
	heartBeatExpirationTimeMs = "2000"  //2 sec
)

type AmqpGameEventSender struct {
	amqpClient *common.AmqpClient
}

func NewAmqpGameEventSender(client *common.AmqpClient) *AmqpGameEventSender {
	return &AmqpGameEventSender{
		amqpClient: client,
	}
}

func (agc *AmqpGameEventSender) SendHeartBeat(message messages.Message[messages.HeartBeatMessagePayload]) error {
	return agc.amqpClient.MarshalAndSend(message, common.HeartBeatQueue, heartBeatExpirationTimeMs)
}

func (agc *AmqpGameEventSender) SendGameEvent(message messages.Message[messages.ActionMessagePayload]) error {
	return agc.amqpClient.MarshalAndSend(message, common.GameEventsQueue, actionExpirationTimeMs)
}
