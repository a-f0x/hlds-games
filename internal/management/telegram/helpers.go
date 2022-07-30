package telegram

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"hlds-games/internal/management"
	"html"
)

func NewRconCallbackData(g management.Game) CallbackData {
	return CallbackData{
		Type: Rcon,
		Data: g.GetApiUrl(),
	}
}

func BuildMessagesWithRconConsole(games []management.Game, chatId int64) tgbotapi.MessageConfig {
	rows := make([][]tgbotapi.InlineKeyboardButton, len(games))
	for i, game := range games {
		button := NewRconCallbackData(game)
		jsonButton, _ := json.Marshal(button)
		rows[i] = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s	%s", html.UnescapeString(string(rune(128225))), game.Name), string(jsonButton),
			),
		)
	}
	numericKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatId, management.BuildGamesText(games))
	msg.ReplyMarkup = numericKeyboard
	return msg
}
