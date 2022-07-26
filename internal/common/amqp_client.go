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
	gameEventsExchange = "hlds-games"
	HeartBeatQueue     = "heart-beat"
	GameEventsQueue    = "game-action"
	contentType        = "application/json"
	reconnectSec       = 2
)

type AmqpClient struct {
	host             string
	port             int64
	user             string
	password         string
	outAmqpChannel   *amqp.Channel
	inputAmqpChannel *amqp.Channel
	connection       *amqp.Connection
	mu               sync.Mutex
	streams          map[string]*chan []byte
}

func NewAmqpClient(host string, port int64, user string, password string) *AmqpClient {
	client := &AmqpClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,

		streams: make(map[string]*chan []byte),
	}
	client.connect()
	return client
}
func (ac *AmqpClient) connect() {
	isConnectedChan := make(chan bool)
	go ac.consume(isConnectedChan)
	go func() {
		err := ac.handleConnection(isConnectedChan)
		if err != nil {
			log.Fatalf("Error connect to amqp: %s\n", err.Error())
		}
	}()
}
func (ac *AmqpClient) MarshalAndSend(message any, queue string, expirationMs string) error {
	bytes, _ := json.Marshal(message)
	return ac.send(bytes, queue, expirationMs)
}

/* вот тут короче вообще дофига вопросов как сделать правильно.
Как отписываться ?
Как правильно делать если хотят подписаться из нескольких горутин ? Видимо придется хранить мапу слайсов каналов.
*/
func (ac *AmqpClient) Subscribe(queue string) (<-chan []byte, error) {
	stream := ac.streams[queue]
	if stream != nil {
		return nil, errors.New(fmt.Sprintf("already subscribed to queue %s", queue))
	}
	s := make(chan []byte)
	ac.streams[queue] = &s
	return s, nil
}

func (ac *AmqpClient) send(message []byte, queue string, expirationMs string) error {
	if ac.isConnected() {
		err := ac.outAmqpChannel.Publish(
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
			_ = ac.outAmqpChannel.Close()
		}
		return nil
	}
	return errors.New("connection not established")
}

func (ac *AmqpClient) handleConnection(isConnectedChan chan<- bool) error {
	for {
		time.Sleep(time.Duration(reconnectSec) * time.Second)
		if ac.isConnected() {
			continue
		}
		err := ac.initConnection(isConnectedChan)
		if err != nil {
			return err
		}
	}
}

/*
Работать не будет если это фанаут эксчейнж и нету роутинг кея.

----эмоции он----
В общем я порядком удивлен, что это все надо делать ручками и нет реконнектов в библиотеке работы с реббитом,
есть ишью https://github.com/rabbitmq/amqp091-go/issues/40
Вообще очень удивлен. Вместо того что бы думать о том как написать бизнес логику надо ебстись что бы написать самому
реконнекты, которые нужны практически во всех кейсах и любому разработчику. Почему нет этого нет в стандартной либе - загадка.
----эмоции офф----
*/
func (ac *AmqpClient) consume(isConnectedChan <-chan bool) {
	for {
		isConnected := <-isConnectedChan
		if isConnected {
			for queue, stream := range ac.streams {
				messages, err := ac.inputAmqpChannel.Consume(
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
		}
	}
}

func (ac *AmqpClient) initConnection(isConnectedChan chan<- bool) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", ac.user, ac.password, ac.host, ac.port)
	log.Printf("Trying to connect amqp: %s:%d\n", ac.host, ac.port)
	conn, openConnectionError := amqp.Dial(url)
	if openConnectionError != nil {
		log.Printf("Error to connect amqp: %s:%d\n%s\nTry reconnect after %d sec.\n", ac.host, ac.port, openConnectionError, reconnectSec)
		isConnectedChan <- false
		return nil
	}

	outChannel, outChannelError := createChannel(conn)
	if outChannelError != nil {
		isConnectedChan <- false
		conn.Close()
		return outChannelError
	}
	ac.outAmqpChannel = outChannel
	inputChannel, inputChannelError := createChannel(conn)
	if inputChannelError != nil {
		isConnectedChan <- false
		conn.Close()
		return inputChannelError
	}

	ac.inputAmqpChannel = inputChannel
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

func (ac *AmqpClient) isConnected() bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.connection == nil {
		return false
	}
	return !ac.connection.IsClosed()
}
