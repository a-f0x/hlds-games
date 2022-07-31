package rabbit

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	gameEventsExchange = "hlds-games"
	HeartBeatQueue     = "heart-beat"
	GameEventsQueue    = "game-action"
	contentType        = "application/json"
)

type amqpClient struct {
	host         string
	port         int64
	user         string
	password     string
	reconnectSec int32
	connection   *amqp.Connection
	channel      *amqp.Channel
	mu           sync.Mutex
}

func newAmqpClient(host string, port int64, user string, password string, reconnectionTimeSec int32) *amqpClient {
	client := &amqpClient{
		host:         host,
		port:         port,
		user:         user,
		password:     password,
		reconnectSec: reconnectionTimeSec,
	}
	return client
}

func (ac *amqpClient) connect() <-chan bool {
	isConnectedChan := make(chan bool)
	go func() {
		err := ac.handleConnection(isConnectedChan)
		if err != nil {
			log.Fatalf("Error connect to amqp: %s\n", err.Error())
		}
	}()
	return isConnectedChan
}

func (ac *amqpClient) handleConnection(isConnectedChan chan<- bool) error {
	ticker := time.NewTicker(time.Duration(ac.reconnectSec) * time.Second)
	for {
		<-ticker.C
		if ac.isConnected() {
			continue
		}
		err := ac.initConnection(isConnectedChan)
		if err != nil {
			return err
		}
	}
}
func (ac *amqpClient) initConnection(isConnectedChan chan<- bool) error {
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

func (ac *amqpClient) isConnected() bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.connection == nil {
		return false
	}
	return !ac.connection.IsClosed()
}
