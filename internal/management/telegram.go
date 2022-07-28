package management

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
)

type Telegram struct {
	config   *TelegramConfig
	bot      *tgbotapi.BotAPI
	chats    []*Chat
	selfName string
	BotEvent chan BotEvent
}

func NewTelegram(config *TelegramConfig) *Telegram {
	return &Telegram{
		config:   config,
		BotEvent: make(chan BotEvent),
	}

}
func (n *Telegram) NotifyAll(message string) {
	for _, group := range n.chats {
		if !group.Muted {
			n.Notify(message, group.Id)
		}
	}
}
func (n *Telegram) Notify(message string, chatId int64) {
	n.sendMessage(message, chatId)
}

func (n *Telegram) sendMessage(message string, chatId int64) {

	msg := tgbotapi.NewMessage(chatId, message)
	_, e := n.bot.Send(msg)
	if e != nil {
		te, ok := e.(tgbotapi.Error)
		if ok {
			if te.Code == 403 || te.Code == 400 {
				n.removeChat(chatId)
				//n.onAction(chatId, "", ChatRemoved)
			}
		}
		log.Println(fmt.Errorf("Error send message : %s \n", e))
	}
}

func (n *Telegram) removeChat(chatId int64) {

	if !n.isChatExist(chatId) {
		return
	}

	chats := make([]*Chat, 0)

	for _, chat := range n.chats {
		if chat.Id == chatId {
			continue
		}
		chats = append(chats, chat)
	}
	n.chats = chats
	n.serializeChats()
}
func (n *Telegram) getChat(chatId int64) *Chat {

	for _, chat := range n.chats {
		if chat.Id == chatId {
			return chat
		}
	}

	return nil
}

func (n *Telegram) isChatExist(chatId int64) bool {
	return n.getChat(chatId) != nil
}

func (n *Telegram) serializeChats() error {
	file, _ := json.MarshalIndent(n.chats, "", " ")
	err := ioutil.WriteFile("./config/chat_groups.json", file, 0644)
	if err != nil {
		return err
	}
	return nil
}
