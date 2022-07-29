package main

import (
	"context"
	"fmt"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/config"
	"hlds-games/internal/management"
	"hlds-games/internal/management/telegram"
	"log"
)

func main() {
	monitoring()
}
func monitoring() {
	common.FakeEnvRabbit("127.0.0.1")

	repository, err := telegram.NewFileChatRepository("./data")
	if err != nil {
		log.Fatal(err)
	}
	t := telegram.NewTelegram(config.GetTelegramBotConfig(), repository)
	botEvents := t.Start()
	rabbitConfig := config.GetRabbitConfig()
	client := rabbit.NewAmqpConsumer(
		rabbitConfig.RabbitHost,
		rabbitConfig.RabbitPort,
		rabbitConfig.RabbitUser,
		rabbitConfig.RabbitPassword,
		2,
	)
	heartBeatChannel, actionChannel, err := management.Collect(context.TODO(), client)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	gm := management.NewGameManager("192.168.88.61")
	for {
		select {
		case heartBeat := <-heartBeatChannel:
			gm.RegisterGame(heartBeat)
		case action := <-actionChannel:
			t.NotifyAll(fmt.Sprintf("action: %v", action.Payload))
		case botEvent := <-botEvents:
			switch action := botEvent.BotAction; action {
			case telegram.ListServers:
				games := gm.ListGames()
				t.Notify(fmt.Sprintf("%v", games), botEvent.ChatId)
			case telegram.RconCommand:
				t.Notify("temporary not implemented...", botEvent.ChatId)
			}
		}
	}
}
