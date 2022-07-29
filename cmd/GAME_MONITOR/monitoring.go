package main

import (
	"bytes"
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
		case actionEvent := <-actionChannel:

			switch action := actionEvent.Payload.ActionType; action {
			case "player_connected":
				message := fmt.Sprintf("%s: %s connected", actionEvent.ServerInfo.GameName, actionEvent.Payload.Player.NickName)
				t.NotifyAll(message)
			case "player_disconnected":
				message := fmt.Sprintf("%s: %s disconnected", actionEvent.ServerInfo.GameName, actionEvent.Payload.Player.NickName)
				t.NotifyAll(message)
			}
		case botEvent := <-botEvents:
			switch action := botEvent.BotAction; action {
			case telegram.ListServers:
				games := gm.ListGames()
				gl := len(games)
				var buffer bytes.Buffer
				for i, game := range games {
					buffer.WriteString(game.String())
					if i < gl-1 {
						buffer.WriteString("\n")
					}
				}
				t.Notify(buffer.String(), botEvent.ChatId)
			case telegram.RconCommand:
				t.Notify("temporary not implemented...", botEvent.ChatId)
			}
		}
	}
}
