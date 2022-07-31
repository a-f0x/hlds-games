package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hlds-games/internal/api"
	"hlds-games/internal/common"
	"hlds-games/internal/common/rabbit"
	"hlds-games/internal/config"
	"hlds-games/internal/management"
	"hlds-games/internal/management/telegram"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	monitoring()
}
func monitoring() {
	common.FakeEnvRabbit("127.0.0.1")
	common.FakeTelegramCfg("5424757267:AAEfIfjXElS5Svf9bp1TVz4HpqRNWorA9BA")
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
				t.SendGameList(games, botEvent.ChatId)
			case telegram.RconCommand:
				address := botEvent.Rcon.ServerAddress
				command := botEvent.Rcon.Command
				chatId := botEvent.ChatId
				messageId := botEvent.Rcon.MessageId
				game := gm.GetGame(address)
				if game == nil {
					t.Reply("Server is offline", chatId, messageId)
					continue
				}
				result, rconError := api.ExecuteRconCommand(address)(context.TODO(), command)
				if rconError != nil {
					log.Printf("fail to send rcon command '%s' to address '%s', Error: %s ", command, address, rconError.Error())
					t.Reply(fmt.Sprintf("%s:\nError send command. Try again later", game.Name), chatId, messageId)
					continue
				}
				log.Warnf("%s: Rcon command '%s' executed by user '%s'. Result: %s", game.Name, command, botEvent.Rcon.UserName, result.Result)
				t.Reply(fmt.Sprintf("%s:\n%s", game.Name, result.Result), chatId, messageId)
			}
		}
	}
}
