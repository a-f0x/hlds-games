package telegram

const (
	ChatAdded   BotAction = 0
	ChatRemoved BotAction = 1
	RconCommand BotAction = 3
)

type Chat struct {
	Name             string `json:"name"`
	Id               int64  `json:"id"`
	Muted            bool   `json:"muted"`
	AllowExecuteRcon bool   `json:"allow_execute_rcon"`
}

type BotEvent struct {
	ChatId    int64
	BotAction BotAction
	Message   string
}
type BotAction uint32
