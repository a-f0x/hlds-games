package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"hlds-games/internal/config"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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
		t.selfName = bot.Self.UserName
		log.Printf("TelegramBot %s connected\n", t.selfName)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 30
		updates := bot.GetUpdatesChan(u)
		for update := range updates {
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
		if !group.Muted {
			t.Notify(message, group.Id)
		}
	}
}
func (t *Telegram) Notify(message string, chatId int64) {
	t.sendMessage(message, chatId)
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
				log.Printf("Error connect to  telegram bot api: %s\nTry reconnect after %d sec...", err, t.config.Bot.ReconnectTimeout)
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

func (t *Telegram) sendMessage(message string, chatId int64) {

	msg := tgbotapi.NewMessage(chatId, message)
	_, e := t.bot.Send(msg)
	if e != nil {
		te, ok := e.(tgbotapi.Error)
		if ok {
			if te.Code == 403 || te.Code == 400 {
				t.removeChat(chatId)

			}
		}
		log.Println(fmt.Errorf("Error send message : %s \n", e))
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
			Name:             userName,
			Id:               chatId,
			Muted:            true,
			AllowExecuteRcon: true,
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
		t.sendMessage("Please enter the password", chatId)
		return
	}

	chat.AllowExecuteRcon = true
	t.updateChat(chat)
	t.sendMessage("You are added as the administrator. Write me a command and I will execute it on the server.\n"+
		"See list of commands http://cs1-6cfg.blogspot.com/p/cs-16-client-and-console-commands.html", chatId)
}

func (t *Telegram) onGroupMessageReceived(chatId int64, groupName string, text string) {

	chat := t.chats.GetChat(chatId)
	if chat == nil {
		chat = t.updateChat(&Chat{
			Name:             groupName,
			Id:               chatId,
			Muted:            true,
			AllowExecuteRcon: true,
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

	case "/mute":
		{
			t.muteChat(chatId, true)
			t.sendMessage("Chat muted", chatId)
			return true
		}
	case "/unmute":
		{
			t.muteChat(chatId, false)
			t.sendMessage("Chat unmuted", chatId)
			return true
		}

	case "/info":
		{
			t.onAction(chatId, "status", RconCommand)
			return true
		}
	}
	return false
}

func (t *Telegram) muteChat(chatId int64, mute bool) {

	chat := t.getChat(chatId)
	if chat == nil {
		return
	}
	chat.Muted = mute
	t.updateChat(chat)
}
