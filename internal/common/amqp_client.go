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
	state                uint32
	mu                   sync.Mutex
}

func NewAmqpClient(host string, port int64, user string, password string) *AmqpClient {
	client := &AmqpClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		state:    stateDisconnected,
	}
	client.connect()
	return client
}
func (ac *AmqpClient) connect() {
	go func() {
		err := ac.tryConnect()
		if err != nil {
			log.Fatalf("Error connect to amqp: %s\n", err.Error())
		}
	}()
}
func (ac *AmqpClient) tryConnect() error {
	if !ac.setConnectionStateConnecting() {
		return nil
	}
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", ac.user, ac.password, ac.host, ac.port)
	log.Printf("Trying to connect amqp: %s:%d\n", ac.host, ac.port)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("Error to connect amqp: %s:%d\n%s\nTry reconnect after %d sec.\n", ac.host, ac.port, err, reconnectSec)
		time.Sleep(time.Duration(reconnectSec) * time.Second)
		ac.setConnectionStateDisconnected()
		return ac.tryConnect()
	}

	ch, err := createChannel(conn)
	if err != nil {
		ac.setConnectionStateDisconnected()
		err := conn.Close()
		if err != nil {
			return err
		}
		return err
	}
	ac.gameEventAmqpChannel = ch
	ac.setConnectionStateConnected()
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
		false,
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
		false,
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
		false,
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
			ac.setConnectionStateDisconnected()
			_ = ac.gameEventAmqpChannel.Close()
			ac.connect()
			return err
		}
		return nil
	}
	return errors.New("connection not established")
}

func (ac *AmqpClient) setConnectionStateConnected() {
	ac.setConnectionState(stateConnected)
}

func (ac *AmqpClient) setConnectionStateDisconnected() {
	ac.setConnectionState(stateDisconnected)
}

func (ac *AmqpClient) setConnectionStateConnecting() bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.state == stateDisconnected {
		ac.state = stateConnecting
		return true
	}
	return false
}

func (ac *AmqpClient) isConnected() bool {
	ac.mu.Lock()
	b := ac.state == stateConnected
	ac.mu.Unlock()
	return b
}

func (ac *AmqpClient) setConnectionState(connectionState uint32) {
	ac.mu.Lock()
	ac.state = connectionState
	ac.mu.Unlock()
}

func (ac *AmqpClient) getConnectionState() uint32 {
	ac.mu.Lock()
	s := ac.state
	ac.mu.Unlock()
	return s
}
