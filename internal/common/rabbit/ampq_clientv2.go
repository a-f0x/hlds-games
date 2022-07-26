package rabbit

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

const (
	gameEventsExchange = "hlds-games"
	HeartBeatQueue     = "heart-beat"
	GameEventsQueue    = "game-action"
	contentType        = "application/json"
)

type AmqpClientV2 struct {
	host         string
	port         int64
	user         string
	password     string
	reconnectSec int32
	connection   *amqp.Connection
	channel      *amqp.Channel
	mu           sync.Mutex
}

func newAmqpClientV2(host string, port int64, user string, password string, reconnectionTimeSec int32) *AmqpClientV2 {
	client := &AmqpClientV2{
		host:         host,
		port:         port,
		user:         user,
		password:     password,
		reconnectSec: reconnectionTimeSec,
	}
	return client
}

func (ac *AmqpClientV2) connect() <-chan bool {
	isConnectedChan := make(chan bool)
	go func() {
		err := ac.handleConnection(isConnectedChan)
		if err != nil {
			log.Fatalf("Error connect to amqp: %s\n", err.Error())
		}
	}()
	return isConnectedChan
}

func (ac *AmqpClientV2) handleConnection(isConnectedChan chan<- bool) error {
	for {
		time.Sleep(time.Duration(ac.reconnectSec) * time.Second)
		if ac.isConnected() {
			continue
		}
		err := ac.initConnection(isConnectedChan)
		if err != nil {
			return err
		}
	}
}
func (ac *AmqpClientV2) initConnection(isConnectedChan chan<- bool) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", ac.user, ac.password, ac.host, ac.port)
	log.Printf("Trying to connect amqp: %s:%d\n", ac.host, ac.port)
	conn, openConnectionError := amqp.Dial(url)
	if openConnectionError != nil {
		log.Printf("Error to connect amqp: %s:%d\n%s\nTry reconnect after %d sec.\n", ac.host, ac.port, openConnectionError, ac.reconnectSec)
		isConnectedChan <- false
		return nil
	}

	channel, channelError := createChannel(conn)
	if channelError != nil {
		isConnectedChan <- false
		conn.Close()
		return channelError
	}
	ac.channel = channel
	ac.connection = conn

	isConnectedChan <- true
	log.Printf("Connection success amqp: %s:%d\n", ac.host, ac.port)
	return nil
}
func createChannel(connection *amqp.Connection) (*amqp.Channel, error) {
	ch, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	err = ch.ExchangeDeclare(
		gameEventsExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	_, err = ch.QueueDeclare(
		HeartBeatQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		GameEventsQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		GameEventsQueue,
		GameEventsQueue,
		gameEventsExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		HeartBeatQueue,
		HeartBeatQueue,
		gameEventsExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (ac *AmqpClientV2) isConnected() bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.connection == nil {
		return false
	}
	return !ac.connection.IsClosed()
}
