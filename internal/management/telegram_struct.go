package management

const (
	ChatAdded   BotAction = 0
	ChatRemoved BotAction = 1
	RconCommand BotAction = 3
)

type TelegramProxyConfig struct {
	Enabled bool
	Url     string
}
type TelegramBotConfig struct {
	Token            string
	ReconnectTimeout int64
	AdminPassword    string
}
type TelegramConfig struct {
	Proxy TelegramProxyConfig
	Bot   TelegramBotConfig
}

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
