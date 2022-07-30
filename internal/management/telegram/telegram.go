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
		Description: "Show server list.",
	}
	playerEventsOnCommand = tgbotapi.BotCommand{
		Command:     "/player_events_on",
		Description: "Send player events.",
	}
	playerEventsOffCommand = tgbotapi.BotCommand{
		Command:     "/player_events_off",
		Description: "Do not send player events.",
	}
	authCommand = tgbotapi.BotCommand{
		Command:     "/auth",
		Description: "Authorization for execute RCON commands. Only direct message.",
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
		cfg := tgbotapi.NewSetMyCommands(listServersCommand, playerEventsOnCommand, playerEventsOffCommand, authCommand)
		_, _ = bot.Request(cfg)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 30
		updates := bot.GetUpdatesChan(u)
		for update := range updates {
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
func (t *Telegram) SendGameList(games []*management.Game, chatId int64) {
	chat := t.getChat(chatId)
	if len(games) == 0 {
		t.sendText("No servers online", chatId)
		return
	}
	if chat.AllowExecuteRcon {
		t.sendMessage(BuildMessagesWithRconConsole(games, chatId))
		return
	}
	t.sendText(management.BuildGamesText(games), chatId)
}

func (t *Telegram) Notify(message string, chatId int64) {
	t.sendText(message, chatId)
}
func (t *Telegram) Reply(message string, chatId int64, messageId int) {
	msg := tgbotapi.NewMessage(chatId, message)
	msg.ReplyToMessageID = messageId
	t.sendMessage(msg)
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

func (t *Telegram) onAction(chatId int64, action BotAction) {
	t.BotEvent <- BotEvent{
		ChatId:    chatId,
		BotAction: action,
	}
}

func (t *Telegram) sendText(message string, chatId int64) {
	t.sendMessage(tgbotapi.NewMessage(chatId, message))
}

func (t *Telegram) sendMessage(messageConfig tgbotapi.MessageConfig) {
	_, e := t.bot.Send(messageConfig)
	if e != nil {
		te, ok := e.(*tgbotapi.Error)
		if ok {
			if te.Code == 403 || te.Code == 400 {
				t.removeChat(messageConfig.ChatID)
			}
		}
		log.Printf("Error send message : %s \n", e)
	}
}

func (t *Telegram) createChat(chatName string, chatId int64, chatType ChatType) *Chat {
	c := &Chat{
		Name:                chatName,
		ChatType:            chatType,
		Id:                  chatId,
		PlayerEventsEnabled: false,
		AllowExecuteRcon:    false,
	}
	err := t.chats.SaveChat(c)
	if err != nil {
		log.Fatalf("fail to add chat. %s", err.Error())
	}
	log.Printf("chat %s created", c.String())
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
	chat, err := t.chats.RemoveChat(chatId)
	if err != nil {
		log.Fatalf("fail to remove chat. %s", err.Error())
	}
	log.Printf("chat %s removed", chat.String())

}

func (t *Telegram) getChat(chatId int64) *Chat {
	return t.chats.GetChat(chatId)
}

func (t *Telegram) onUpdateReceived(update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.Chat.Type != "group" && update.Message.Chat.Type != "supergroup" {
			t.onDirectMessageReceived(update)
			return
		}
		t.onGroupMessageReceived(update)
		return
	}
	if update.CallbackData() != "" {
		t.onCallback(update)
	}
}

func (t *Telegram) onDirectMessageReceived(update tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	userName := update.Message.From.UserName
	text := update.Message.Text
	chat := t.getChat(chatId)
	if chat == nil {
		chat = t.createChat(userName, chatId, DirectChat)
	}
	args := strings.Split(text, " ")
	if t.onBotCommandReceived(strings.Join(args[:1], ""), args[1:], chatId, true) {
		return
	}
	if chat.CurrentRconAddress != "" && !chat.AllowExecuteRcon {
		chat.CurrentRconAddress = ""
		t.updateChat(chat)
	}
	if !chat.AllowExecuteRcon && t.config.Bot.AdminPassword == text {
		chat.AllowExecuteRcon = true
		t.updateChat(chat)
		t.sendText("You are added as the administrator. Write me a command and I will execute it on the server.\n"+
			"See list of commands http://cs1-6cfg.blogspot.com/p/cs-16-client-and-console-commands.html", chatId)
		t.onAction(chatId, ListServers)
	}
	if chat.AllowExecuteRcon && chat.CurrentRconAddress != "" {
		t.BotEvent <- BotEvent{
			ChatId:    chatId,
			BotAction: RconCommand,
			Rcon: &ExecuteRcon{
				ServerAddress: chat.CurrentRconAddress,
				Command:       text,
				MessageId:     update.Message.MessageID,
			},
		}
	}
}

func (t *Telegram) onGroupMessageReceived(update tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	groupName := update.Message.Chat.Title
	text := update.Message.Text
	chat := t.getChat(chatId)
	if chat == nil {
		chat = t.createChat(groupName, chatId, GroupChat)
	}
	if !strings.Contains(text, t.selfName) {
		return
	}
	args := strings.Split(text, "@"+t.selfName)
	t.onBotCommandReceived(strings.Join(args[:1], ""), strings.Fields(strings.Join(args[1:], " ")), chatId, false)
}

func (t *Telegram) onBotCommandReceived(command string, args []string, chatId int64, direct bool) bool {
	switch command {
	case playerEventsOffCommand.Command:
		t.allowSendPlayerEvents(chatId, false)
		return true
	case playerEventsOnCommand.Command:
		t.allowSendPlayerEvents(chatId, true)
		return true
	case listServersCommand.Command:
		t.onAction(chatId, ListServers)
		return true
	case authCommand.Command:
		if !direct {
			return false
		}
		t.auth(chatId)
		return true
	}
	return false
}
func (t *Telegram) auth(chatId int64) {
	chat := t.getChat(chatId)
	if chat.AllowExecuteRcon {
		t.sendText("Already authorized", chatId)
		return
	}
	t.sendText("Please enter the password", chatId)

}

//Attention! Update.Message is nil!
func (t *Telegram) onCallback(update tgbotapi.Update) {
	cd := update.CallbackData()
	cq := update.CallbackQuery
	chatId := cq.Message.Chat.ID
	t.hideMarkup(chatId, cq.Message.MessageID)
	t.sendText("Enter rcon command", chatId)
	data := &CallbackData{}
	err := json.Unmarshal([]byte(cd), data)
	if err != nil {
		log.Printf("Error parse callback: %s. %s", cd, err.Error())
		t.sendText(fmt.Sprintf("Internal error. %s", err.Error()), chatId)
		return
	}
	switch data.Type {
	case Rcon:
		chat := t.getChat(chatId)
		chat.CurrentRconAddress = data.Data
		t.updateChat(chat)
	}
}

func (t *Telegram) allowSendPlayerEvents(chatId int64, allow bool) {
	chat := t.getChat(chatId)
	if chat == nil {
		return
	}
	chat.PlayerEventsEnabled = allow
	t.updateChat(chat)
}

func (t *Telegram) hideMarkup(chatId int64, messageId int) {
	emptyMarkUp := tgbotapi.NewEditMessageReplyMarkup(
		chatId,
		messageId,
		tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("", "hide"),
			),
		),
	)
	_, err := t.bot.Send(emptyMarkUp)
	if err != nil {
		log.Printf("Error send message : %s \n", err)
	}
}
