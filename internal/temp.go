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
	consume()
	produce()

}
func produce() {
	type message struct {
		Num  int       `json:"num"`
		Time time.Time `json:"time"`
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
		m := message{
			Num:  i,
			Time: time.Now(),
		}
		err := producer.MarshallAndSend(m, common.GameEventsQueue, "60000")
		if err != nil {
			//log.Printf("error when send %v, %s", m, err)
		}
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
	time.Sleep(time.Duration(5) * time.Second)
	subscribe, err := consumer.Subscribe(common.GameEventsQueue)
	if err != nil {
		log.Fatalf(err.Error())
	}
	go func() {
		for {
			bytes := <-subscribe
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
