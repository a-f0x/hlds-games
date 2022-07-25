package game

import (
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"hlds-games/pkg/messages"
	"log"
	"sync"
	"time"
)

const (
	gameEventsExchange               = "hlds-games"
	heartBeatQueue                   = "heart-beat"
	gameEventsQueue                  = "game-action"
	contentType                      = "application/json"
	actionExpirationTimeMs           = "15000" //15 sec
	heartBeatExpirationTimeMs        = "2000"  //2 sec
	reconnectSec                     = 3
	stateDisconnected         uint32 = 0
	stateConnected            uint32 = 1
	stateConnecting           uint32 = 2
)

type AmqpGameClient struct {
	host                 string
	port                 int64
	user                 string
	password             string
	gameEventAmqpChannel *amqp.Channel
	state                uint32
	mu                   sync.Mutex
}

func NewAmqGameClient(host string, port int64, user string, password string) *AmqpGameClient {
	return &AmqpGameClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		state:    stateDisconnected,
	}
}

func (agc *AmqpGameClient) Connect() error {
	return agc.tryConnect()
}

func (agc *AmqpGameClient) tryConnect() error {
	if !agc.setConnectionStateConnecting() {
		return nil
	}
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", agc.user, agc.password, agc.host, agc.port)
	log.Printf("Trying to connect amqp: %s:%d\n", agc.host, agc.port)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("Error to connect amqp: %s:%d\n%s\nTry reconnect after %d sec.\n", agc.host, agc.port, err, reconnectSec)
		time.Sleep(time.Duration(reconnectSec) * time.Second)
		agc.setConnectionStateDisconnected()
		return agc.tryConnect()
	}

	ch, err := createChannel(conn, gameEventsExchange)
	if err != nil {
		agc.setConnectionStateDisconnected()
		err := conn.Close()
		if err != nil {
			return err
		}
		return err
	}
	agc.gameEventAmqpChannel = ch
	agc.setConnectionStateConnected()
	log.Printf("Connection success amqp: %s:%d\n", agc.host, agc.port)
	return nil
}

func createChannel(connection *amqp.Connection, name string) (*amqp.Channel, error) {
	ch, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	err = ch.ExchangeDeclare(
		name,
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
		heartBeatQueue,
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
		gameEventsQueue,
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
		gameEventsQueue,
		gameEventsQueue,
		gameEventsExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		heartBeatQueue,
		heartBeatQueue,
		gameEventsExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (agc *AmqpGameClient) SendHeartBeat(message messages.Message[messages.HeartBeatMessagePayload]) error {
	return agc.marshalAndSend(message, heartBeatQueue, heartBeatExpirationTimeMs)
}

func (agc *AmqpGameClient) SendGameEvent(message messages.Message[messages.ActionMessagePayload]) error {
	return agc.marshalAndSend(message, gameEventsQueue, actionExpirationTimeMs)
}

func (agc *AmqpGameClient) marshalAndSend(message any, queue string, expirationMs string) error {
	bytes, _ := json.Marshal(message)
	return agc.send(bytes, queue, expirationMs)
}

func (agc *AmqpGameClient) send(message []byte, queue string, expirationMs string) error {
	if agc.isConnected() {
		err := agc.gameEventAmqpChannel.Publish(
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
			agc.setConnectionStateDisconnected()
			_ = agc.gameEventAmqpChannel.Close()
			go func() {
				_ = agc.tryConnect()
			}()
			return err
		}
		return nil
	}
	return errors.New("connection not established")
}

func (agc *AmqpGameClient) setConnectionStateConnected() {
	agc.setConnectionState(stateConnected)
}

func (agc *AmqpGameClient) setConnectionStateDisconnected() {
	agc.setConnectionState(stateDisconnected)
}

func (agc *AmqpGameClient) setConnectionStateConnecting() bool {
	agc.mu.Lock()
	defer agc.mu.Unlock()
	if agc.state == stateDisconnected {
		agc.state = stateConnecting
		return true
	}
	return false
}

func (agc *AmqpGameClient) isConnected() bool {
	agc.mu.Lock()
	b := agc.state == stateConnected
	agc.mu.Unlock()
	return b
}

func (agc *AmqpGameClient) setConnectionState(connectionState uint32) {
	agc.mu.Lock()
	agc.state = connectionState
	agc.mu.Unlock()
}

func (agc *AmqpGameClient) getConnectionState() uint32 {
	agc.mu.Lock()
	s := agc.state
	agc.mu.Unlock()
	return s
}
