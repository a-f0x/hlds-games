package rabbit

import "github.com/rabbitmq/amqp091-go"

type stream struct {
	queue           string
	incomingChannel <-chan amqp091.Delivery
	outChan         chan []byte
}

func (s *stream) consume() {
	go func() {
		for delivery := range s.incomingChannel {
			s.outChan <- delivery.Body
		}
	}()
}
