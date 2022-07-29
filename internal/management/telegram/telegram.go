package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"hlds-games/internal/config"
	"log"
	"os"
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
func (t *Telegram) Start() {

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

func (t *Telegram) addChat(chatName string, chatId int64, muted bool, allowRcon bool) {

	err := t.chats.AddChat(Chat{
		chatName,
		chatId,
		muted,
		allowRcon})
	if err != nil {
		log.Fatalf("fail to add chat. %s", err.Error())
	}
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

func (t *Telegram) isChatExist(chatId int64) bool {
	return t.getChat(chatId) != nil
}
