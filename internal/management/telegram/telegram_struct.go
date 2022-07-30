package telegram

import "fmt"

const (
	ListServers BotAction = 0
	RconCommand BotAction = 1
	DirectChat  ChatType  = "direct"
	GroupChat   ChatType  = "group"
)

type ChatType string

type Chat struct {
	Name                string   `json:"name"`
	Id                  int64    `json:"id"`
	ChatType            ChatType `json:"chat_type"`
	PlayerEventsEnabled bool     `json:"player_events_enabled"`
	AllowExecuteRcon    bool     `json:"allow_execute_rcon"`
}

func (c Chat) String() any {
	return fmt.Sprintf("Name:%s, ChatType: %s, Id: %d", c.Name, c.ChatType, c.Id)
}

type BotEvent struct {
	ChatId    int64
	BotAction BotAction
}
type BotAction uint32

type GameButtonType int32

const (
	Rcon GameButtonType = 1
)

type CallbackData struct {
	Type GameButtonType `json:"type"`
	Data string         `json:"data"`
}
