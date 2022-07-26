package rabbit

import (
	"errors"
	"fmt"
	"log"
)

type AmqpConsumer struct {
	client *AmqpClientV2
	//isConnectedChan   <-chan bool
	streams map[string]*chan []byte
}

func NewAmqpConsumer(host string, port int64, user string, password string, reconnectionTimeSec int32) *AmqpConsumer {
	c := AmqpConsumer{
		client:  newAmqpClientV2(host, port, user, password, reconnectionTimeSec),
		streams: make(map[string]*chan []byte),
	}
	connectionInfo := c.client.connect()
	go c.watch(connectionInfo)
	return &c
}

func (ac *AmqpConsumer) watch(connectionInfo <-chan bool) {
	for isConnectedState := range connectionInfo {
		log.Printf("consumer state %v", isConnectedState)
		if isConnectedState {
			for queue, stream := range ac.streams {
				messages, err := ac.client.channel.Consume(
					queue,
					"",
					true,
					false,
					false,
					false,
					nil,
				)
				if err != nil {
					log.Fatalf("Queue %s consumer error %s", queue, err)
				}
				stream := stream
				go func() {
					for delivery := range messages {
						*stream <- delivery.Body
					}
				}()
			}
		} else {

		}
	}
}

func (ac *AmqpConsumer) Subscribe(queue string) (<-chan []byte, error) {
	stream := ac.streams[queue]
	if stream != nil {
		return nil, errors.New(fmt.Sprintf("already subscribed to queue %s", queue))
	}
	s := make(chan []byte)
	ac.streams[queue] = &s
	return s, nil
}
