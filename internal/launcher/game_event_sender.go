package launcher

import (
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/messages"
)

const (
	actionExpirationTimeMs    = "15000" //15 sec
	heartBeatExpirationTimeMs = "2000"  //2 sec
)

type AmqpGameEventSender struct {
	amqpClient *rabbit.AmqpProducer
}

func NewAmqpGameEventSender(client *rabbit.AmqpProducer) *AmqpGameEventSender {
	return &AmqpGameEventSender{
		amqpClient: client,
	}
}

func (agc *AmqpGameEventSender) SendHeartBeat(message messages.Message[messages.HeartBeatMessagePayload]) error {
	return agc.amqpClient.MarshallAndSend(message, rabbit.HeartBeatQueue, heartBeatExpirationTimeMs)
}

func (agc *AmqpGameEventSender) SendGameEvent(message messages.Message[messages.ActionMessagePayload]) error {
	return agc.amqpClient.MarshallAndSend(message, rabbit.GameEventsQueue, actionExpirationTimeMs)
}
