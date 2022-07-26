package rabbit

import (
	"encoding/json"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"hlds-games/internal/common"
)

type AmqpProducer struct {
	client      *amqpClient
	isConnected *common.AtomicBool
}

func NewAmqpProducer(host string, port int64, user string, password string, reconnectionTimeSec int32) *AmqpProducer {
	p := AmqpProducer{
		client:      newAmqpClient(host, port, user, password, reconnectionTimeSec),
		isConnected: new(common.AtomicBool),
	}
	connectionInfo := p.client.connect()
	go p.watch(connectionInfo)
	return &p
}
func (ap *AmqpProducer) watch(connectionInfo <-chan bool) {
	for isConnectedState := range connectionInfo {
		//log.Printf("producer state %v", isConnectedState)
		ap.isConnected.Set(isConnectedState)
	}
}

func (ap *AmqpProducer) MarshallAndSend(message any, queue string, expirationMs string) error {
	bytes, _ := json.Marshal(message)
	return ap.send(bytes, queue, expirationMs)
}

func (ap *AmqpProducer) send(message []byte, queue string, expirationMs string) error {
	if ap.isConnected.Get() {
		err := ap.client.channel.Publish(
			gameEventsExchange,
			queue,
			false,
			false,
			amqp.Publishing{
				Expiration:  expirationMs,
				ContentType: contentType,
				Body:        message,
			})
		if err != nil {
			_ = ap.client.channel.Close()
		}
		return nil
	}
	return errors.New("connection not established")
}
