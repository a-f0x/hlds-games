package telegram

const (
	ListServers BotAction = 0
	RconCommand BotAction = 1
)

type Chat struct {
	Name                string `json:"name"`
	Id                  int64  `json:"id"`
	PlayerEventsEnabled bool   `json:"player_events_enabled"`
	AllowExecuteRcon    bool   `json:"allow_execute_rcon"`
}

type BotEvent struct {
	ChatId    int64
	BotAction BotAction
	Message   string
}
type BotAction uint32

type GameButtonType string

const (
	Rcon GameButtonType = "rcon"
)

type GameButton struct {
	Type GameButtonType `json:"type"`
	Key  string         `json:"key"`
}
