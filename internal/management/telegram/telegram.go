package telegram

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"hlds-games/internal/config"
	"hlds-games/internal/management"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	listServersCommand = tgbotapi.BotCommand{
		Command:     "/list",
		Description: "List servers info",
	}
	playerEventsOnCommand = tgbotapi.BotCommand{
		Command:     "/player_events_on",
		Description: "Send player events",
	}
	playerEventsOffCommand = tgbotapi.BotCommand{
		Command:     "/player_events_off",
		Description: "Do not send player events",
	}
)

type Telegram struct {
	config   *config.TelegramConfig
	bot      *tgbotapi.BotAPI
	selfName string
	chats    ChatRepository
	BotEvent chan BotEvent
}

func NewTelegram(config *config.TelegramConfig, chatRepository ChatRepository) *Telegram {
	t := &Telegram{
		config:   config,
		chats:    chatRepository,
		BotEvent: make(chan BotEvent),
	}
	if config.Proxy.Enabled {
		err := os.Setenv("HTTP_PROXY", config.Proxy.Url)
		if err != nil {
			log.Printf("can`t set HTTP_PROXY env. %s", err.Error())
		}
	}

	return t
}
func (t *Telegram) Start() <-chan BotEvent {
	go func() {
		bot := t.tryConnect()
		t.bot = bot
		t.selfName = bot.Self.UserName
		log.Printf("TelegramBot %s connected\n", t.selfName)
		cfg := tgbotapi.NewSetMyCommands(listServersCommand, playerEventsOnCommand, playerEventsOffCommand)
		_, _ = bot.Request(cfg)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 30
		updates := bot.GetUpdatesChan(u)
		for update := range updates {
			log.Printf("update %v", update)
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}
			t.onUpdateReceived(update)
		}
	}()
	return t.BotEvent
}

func (t *Telegram) NotifyAll(message string) {
	for _, group := range t.chats.GetAll() {
		if group.PlayerEventsEnabled {
			t.Notify(message, group.Id)
		}
	}
}
func (t *Telegram) SendGameList(games []management.Game, chatId int64) {
	chat := t.getChat(chatId)
	if len(games) == 0 {
		t.sendText("No servers online", chatId)
		return
	}
	if chat.AllowExecuteRcon {
		t.sendMessage(buildMessagesWithConsole(games, chatId))

	}
}

func buildMessagesWithConsole(games []management.Game, chatId int64) tgbotapi.MessageConfig {
	//btns := make([]tgbotapi.InlineKeyboardButton, len(games))
	//for i, game := range games {
	//	button := GameButton{Rcon, game.Key()}
	//	jsonButton, _ := json.Marshal(button)
	//	btns[i] = tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("rcon %s", game.Name), string(jsonButton))
	//}
	//row := tgbotapi.NewInlineKeyboardRow(btns...  )
	//numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	//msg := tgbotapi.NewMessage(chatId, management.BuildGamesText(games))
	//msg.ReplyMarkup = numericKeyboard
	//return msg

	game := games[0]
	button := GameButton{
		Rcon,
		game.Key(),
	}
	jsonButton, _ := json.Marshal(button)
	row := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("rcon %s", game.Name), string(jsonButton)),
	)
	numericKeyboard :=
		tgbotapi.NewInlineKeyboardMarkup(row)
	msg := tgbotapi.NewMessage(chatId, management.BuildGamesText(games))
	msg.ReplyMarkup = numericKeyboard
	return msg
}

func (t *Telegram) Notify(message string, chatId int64) {
	t.sendText(message, chatId)
}

func (t *Telegram) tryConnect() *tgbotapi.BotAPI {
	log.Println("Trying to connect telegram bot api...")
	b, err := tgbotapi.NewBotAPIWithClient(t.config.Bot.Token, tgbotapi.APIEndpoint, &http.Client{})
	if err != nil {
		ticker := time.NewTicker(time.Duration(t.config.Bot.ReconnectTimeout) * time.Second)
		for {
			<-ticker.C
			b, err := tgbotapi.NewBotAPIWithClient(t.config.Bot.Token, tgbotapi.APIEndpoint, &http.Client{})
			if err != nil {
				log.Printf("Error connect to telegram bot api: %s\nTry reconnect after %d sec...", err, t.config.Bot.ReconnectTimeout)
				continue
			}
			ticker.Stop()
			return b
		}
	}
	return b
}

func (t *Telegram) onAction(chatId int64, message string, action BotAction) {
	t.BotEvent <- BotEvent{
		ChatId:    chatId,
		BotAction: action,
		Message:   message,
	}
}

func (t *Telegram) sendText(message string, chatId int64) {
	t.sendMessage(tgbotapi.NewMessage(chatId, message))
}

func (t *Telegram) sendMessage(messageConfig tgbotapi.MessageConfig) {
	_, e := t.bot.Send(messageConfig)
	if e != nil {
		te, ok := e.(tgbotapi.Error)
		if ok {
			if te.Code == 403 || te.Code == 400 {
				t.removeChat(messageConfig.ChatID)
			}
		}
		log.Printf("Error send message : %s \n", e)
	}
}

func (t *Telegram) createChat(chatName string, chatId int64, muted bool, allowRcon bool) *Chat {
	c := &Chat{
		chatName,
		chatId,
		muted,
		allowRcon}
	err := t.chats.SaveChat(c)
	if err != nil {
		log.Fatalf("fail to add chat. %s", err.Error())
	}
	return c
}

func (t *Telegram) updateChat(chat *Chat) *Chat {
	err := t.chats.SaveChat(chat)
	if err != nil {
		log.Fatalf("fail to add chat. %s", err.Error())
	}
	return chat
}

func (t *Telegram) removeChat(chatId int64) {
	err := t.chats.RemoveChat(chatId)
	if err != nil {
		log.Fatalf("fail to remove chat. %s", err.Error())
	}
}

func (t *Telegram) getChat(chatId int64) *Chat {
	return t.chats.GetChat(chatId)
}

func (t *Telegram) onUpdateReceived(update tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	groupName := update.Message.Chat.Title
	userName := update.Message.From.UserName
	text := update.Message.Text
	if update.Message.Chat.Type != "group" && update.Message.Chat.Type != "supergroup" {
		t.onDirectMessageReceived(chatId, userName, text)
		return
	}
	t.onGroupMessageReceived(chatId, groupName, text)
}

func (t *Telegram) onDirectMessageReceived(chatId int64, userName string, text string) {
	chat := t.chats.GetChat(chatId)
	if chat == nil {
		chat = t.updateChat(&Chat{
			Name:                userName,
			Id:                  chatId,
			PlayerEventsEnabled: false,
			AllowExecuteRcon:    false,
		})
	}
	args := strings.Split(text, " ")
	if t.onBotCommandReceived(strings.Join(args[:1], ""), args[1:], chatId) {
		return
	}
	if chat.AllowExecuteRcon {
		t.onAction(chatId, text, RconCommand)
		return
	}
	if t.config.Bot.AdminPassword != text {
		t.sendText("Please enter the password", chatId)
		return
	}
	chat.AllowExecuteRcon = true
	t.updateChat(chat)
	t.sendText("You are added as the administrator. Write me a command and I will execute it on the server.\n"+
		"See list of commands http://cs1-6cfg.blogspot.com/p/cs-16-client-and-console-commands.html", chatId)
}

func (t *Telegram) onGroupMessageReceived(chatId int64, groupName string, text string) {
	chat := t.chats.GetChat(chatId)
	if chat == nil {
		chat = t.updateChat(&Chat{
			Name:                groupName,
			Id:                  chatId,
			PlayerEventsEnabled: false,
			AllowExecuteRcon:    false,
		})
	}
	if !strings.Contains(text, t.selfName) {
		return
	}
	args := strings.Split(text, "@"+t.selfName)
	t.onBotCommandReceived(strings.Join(args[:1], ""), strings.Fields(strings.Join(args[1:], " ")), chatId)
}

func (t *Telegram) onBotCommandReceived(command string, args []string, chatId int64) bool {
	switch command {
	case playerEventsOffCommand.Command:
		t.allowSendPlayerEvents(chatId, false)
		return true
	case playerEventsOnCommand.Command:
		t.allowSendPlayerEvents(chatId, true)
		return true
	case listServersCommand.Command:
		t.onAction(chatId, "", ListServers)
		return true
	}
	return false
}

func (t *Telegram) allowSendPlayerEvents(chatId int64, allow bool) {
	chat := t.getChat(chatId)
	if chat == nil {
		return
	}
	chat.PlayerEventsEnabled = allow
	t.updateChat(chat)
}
