package main

import (
	"context"
	"fmt"
	"hlds-games/internal/api"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/config"
	"hlds-games/internal/management/telegram"
	"hlds-games/internal/rcon"
	"hlds-games/internal/stats"
	"log"
	"os"
	"time"
)

func main() {
	testTelegram()

}

func testTelegram() {
	common.FakeTelegramCfg("fake_token")
	repository, err := telegram.NewFileChatRepository("./data")
	if err != nil {
		log.Fatal(err)
	}
	ch := telegram.NewTelegram(config.GetTelegramBotConfig(), repository).Start()
	for {
		e := <-ch
		fmt.Printf("message: %v", e)
	}

}

func testChatRepo() {
	repository, err := telegram.NewFileChatRepository("./data")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("chats: %v", repository.GetAll())
	repository.SaveChat(
		&telegram.Chat{Name: "chatnameUpdated", Id: 1},
	)
	log.Printf("chats: %v", repository.GetAll())

	repository.SaveChat(
		&telegram.Chat{Name: "chatname11Updated", Id: 11},
	)
	log.Printf("chats: %v", repository.GetAll())

	repository.RemoveChat(1)
	log.Printf("chats: %v", repository.GetAll())
}
func testGrpcRconCommand() {
	common.FakeEnvGameCfg()
	hldsGameConfig := config.GetHldsGameConfig()
	rc := rcon.NewRcon(hldsGameConfig.Host, hldsGameConfig.HldsGamePort, hldsGameConfig.RconPassword)
	grpcApiConfig := config.GetGrpcApiConfig()
	apiServer := api.NewHLDSApiServer(hldsGameConfig.GameType, grpcApiConfig, rc)
	go apiServer.RunServer()
	api.GetServerInfo("127.0.0.1", grpcApiConfig.GrpcApiPort)

}
func testRabbit() {
	common.FakeEnvRabbit("127.0.0.1")
	rbtCfg := config.GetRabbitConfig()
	go consume(rbtCfg)
	produce(rbtCfg)

}
func produce(rabbitConfig *config.RabbitConfig) {
	type message struct {
		Num   int       `json:"num"`
		Queue string    `json:"queue"`
		Time  time.Time `json:"time"`
	}
	producer := rabbit.NewAmqpProducer(
		rabbitConfig.RabbitHost,
		rabbitConfig.RabbitPort,
		rabbitConfig.RabbitUser,
		rabbitConfig.RabbitPassword,
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

func consume(rabbitConfig *config.RabbitConfig) {
	consumer := rabbit.NewAmqpConsumer(
		rabbitConfig.RabbitHost,
		rabbitConfig.RabbitPort,
		rabbitConfig.RabbitUser,
		rabbitConfig.RabbitPassword,
		1,
	)
	//time.Sleep(time.Duration(5) * time.Second)
	subscribeGameEventsQueue, err := consumer.Subscribe(context.TODO(), rabbit.GameEventsQueue)
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
	subscribeHeartBeatQueue, err2 := consumer.Subscribe(context.TODO(), rabbit.HeartBeatQueue)
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
