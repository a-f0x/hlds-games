package common

import (
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

const (
	gameEventsExchange        = "hlds-games"
	HeartBeatQueue            = "heart-beat"
	GameEventsQueue           = "game-action"
	contentType               = "application/json"
	reconnectSec              = 2
	stateDisconnected  uint32 = 0
	stateConnected     uint32 = 1
	stateConnecting    uint32 = 2
)

type AmqpClient struct {
	host                 string
	port                 int64
	user                 string
	password             string
	gameEventAmqpChannel *amqp.Channel
	connection           *amqp.Connection
	//state                uint32
	mu              sync.Mutex
	isConnectedChan chan bool
}

func NewAmqpClient(host string, port int64, user string, password string) *AmqpClient {
	client := &AmqpClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		//state:           stateDisconnected,
		isConnectedChan: make(chan bool),
	}
	return client
}
func (ac *AmqpClient) Connect() <-chan bool {
	go func() {
		err := ac.handleConnection()
		if err != nil {
			log.Fatalf("Error connect to amqp: %s\n", err.Error())
		}
	}()
	return ac.isConnectedChan
}

func (ac *AmqpClient) handleConnection() error {
	for {
		time.Sleep(time.Duration(reconnectSec) * time.Second)
		if ac.isConnected() {
			continue
		}
		err := ac.initConnection()
		if err != nil {
			return err
		}
	}
}

func (ac *AmqpClient) initConnection() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", ac.user, ac.password, ac.host, ac.port)
	log.Printf("Trying to connect amqp: %s:%d\n", ac.host, ac.port)
	conn, openConnectionError := amqp.Dial(url)
	if openConnectionError != nil {
		log.Printf("Error to connect amqp: %s:%d\n%s\nTry reconnect after %d sec.\n", ac.host, ac.port, openConnectionError, reconnectSec)
		ac.isConnectedChan <- false
		return nil
	}

	ch, channelError := createChannel(conn)
	if channelError != nil {
		ac.isConnectedChan <- false
		conn.Close()
		return channelError
	}

	ac.connection = conn
	ac.gameEventAmqpChannel = ch
	ac.isConnectedChan <- true
	//ac.setConnectionStateConnected()
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

func (ac *AmqpClient) MarshalAndSend(message any, queue string, expirationMs string) error {
	bytes, _ := json.Marshal(message)
	return ac.send(bytes, queue, expirationMs)
}

func (ac *AmqpClient) Stream(queue string) <-chan []byte {
	messages, err := ac.gameEventAmqpChannel.Consume(
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
	resultChan := make(chan []byte)
	go func() {
		for delivery := range messages {
			resultChan <- delivery.Body
		}
	}()

	return resultChan
}

func (ac *AmqpClient) send(message []byte, queue string, expirationMs string) error {
	if ac.isConnected() {
		err := ac.gameEventAmqpChannel.Publish(
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
			_ = ac.gameEventAmqpChannel.Close()
			//ac.setConnectionStateDisconnected()
		}
		return nil
	}
	return errors.New("connection not established")
}

//func (ac *AmqpClient) setConnectionStateConnected() {
//	ac.mu.Lock()
//	ac.state = stateConnected
//	ac.isConnectedChan <- true
//	ac.mu.Unlock()
//}

//func (ac *AmqpClient) setConnectionStateDisconnected() {
//	ac.mu.Lock()
//	if ac.connection != nil {
//		ac.connection.Close()
//	}
//
//	//ac.state = stateDisconnected
//	ac.isConnectedChan <- false
//	ac.mu.Unlock()
//}

//func (ac *AmqpClient) setConnectionStateConnecting() bool {
//	ac.mu.Lock()
//	defer ac.mu.Unlock()
//	if ac.state == stateDisconnected {
//		ac.state = stateConnecting
//		return true
//	}
//	return false
//}

func (ac *AmqpClient) isConnected() bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.connection == nil {
		return false
	}
	return !ac.connection.IsClosed()

}

//func (ac *AmqpClient) getConnectionState() uint32 {
//	ac.mu.Lock()
//	s := ac.state
//	ac.mu.Unlock()
//	return s
//}
