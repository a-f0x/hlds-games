package messages

type MessageType string

const (
	HeartBeat MessageType = "heart_beat"
	Action    MessageType = "action"
)

type ServerInfo struct {
	Game string `json:"game"`
	Name string `json:"name"`
	Host string `json:"game_host"`
}

type Message[P messageConstraint] struct {
	ServerInfo  ServerInfo  `json:"server_info"`
	Time        int64       `json:"time"`
	MessageType MessageType `json:"message_type"`
	Payload     P           `json:"payload"`
}

type messageConstraint interface {
	HeartBeatMessagePayload | ActionMessagePayload
}

type HeartBeatMessagePayload struct {
	Players int32  `json:"players"`
	Map     string `json:"map"`
}

type Player struct {
	NickName string  `json:"nick_name"`
	Team     *string `json:"team"`
}
type Kill struct {
	Victim Player `json:"victim"`
	Weapon string `json:"weapon"`
}

type Suicide struct {
	Weapon string `json:"weapon"`
}

type ActionMessagePayload struct {
	ActionType string   `json:"action_type"`
	Player     Player   `json:"player"`
	Kill       *Kill    `json:"kill"`
	Suicide    *Suicide `json:"suicide"`
}
