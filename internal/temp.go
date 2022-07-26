package main

import (
	"fmt"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/rcon"
	"hlds-games/internal/stats"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	testRabbit()
}

func testRabbit() {
	common.FakeEnvRabbit("127.0.0.1")
	go consume()
	produce()

}
func produce() {
	type message struct {
		Num   int       `json:"num"`
		Queue string    `json:"queue"`
		Time  time.Time `json:"time"`
	}
	amqpPort, err := strconv.ParseInt(*common.GetEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}
	producer := rabbit.NewAmqpProducer(
		*common.GetEnv("RABBITMQ_HOST"),
		amqpPort, *common.GetEnv("RABBITMQ_USER"),
		*common.GetEnv("RABBITMQ_PASSWORD"),
		1,
	)

	for i := 1; i <= 100; i++ {
		time.Sleep(time.Duration(2) * time.Second)
		producer.MarshallAndSend(message{
			Num:   i,
			Queue: "GameEventsQueue",
			Time:  time.Now(),
		}, rabbit.GameEventsQueue, "60000")

		producer.MarshallAndSend(message{
			Num:   i,
			Queue: "HeartBeatQueue",
			Time:  time.Now(),
		}, rabbit.HeartBeatQueue, "60000")

	}
}

func consume() {
	amqpPort, err := strconv.ParseInt(*common.GetEnv("RABBITMQ_PORT"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid RABBITMQ_PORT %s", err.Error())
	}
	consumer := rabbit.NewAmqpConsumer(
		*common.GetEnv("RABBITMQ_HOST"),
		amqpPort, *common.GetEnv("RABBITMQ_USER"),
		*common.GetEnv("RABBITMQ_PASSWORD"),
		1,
	)
	//time.Sleep(time.Duration(5) * time.Second)
	subscribeGameEventsQueue, err := consumer.Subscribe(rabbit.GameEventsQueue)
	if err != nil {
		log.Fatalf(err.Error())
	}

	go func() {
		for {
			bytes := <-subscribeGameEventsQueue
			log.Printf("incoming <-%s", string(bytes))
		}
	}()

	//time.Sleep(time.Duration(5) * time.Second)
	subscribeHeartBeatQueue, err2 := consumer.Subscribe(rabbit.HeartBeatQueue)
	if err2 != nil {
		log.Fatalf(err2.Error())
	}

	go func() {
		for {
			bytes := <-subscribeHeartBeatQueue
			log.Printf("incoming <-%s", string(bytes))
		}
	}()

}

func readStats() {
	path := "./data/csstats.dat"
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return
	}

	b := make([]byte, stat.Size())
	file.Read(b)
	sd := stats.NewStatsReader(b)
	result := sd.ReadStats()
	fmt.Printf("stats: %v", result)

	rcon := rcon.NewRcon("127.0.0.1", 27017, "asjop2340239857uG")

	status, err := rcon.GetServerStatus()
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Printf("status: %s\n", status)

}
