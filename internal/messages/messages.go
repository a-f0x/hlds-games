package messages

type MessageType string

const (
	HeartBeat MessageType = "heart_beat"
	Action    MessageType = "action"
)

type ServerInfo struct {
	Game string `json:"game"`
	Name string `json:"name"`
	Host string `json:"host"`
	Port int64  `json:"port"`
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
	ApiHost string `json:"api_host"`
}

type Player struct {
	NickName string  `json:"nick_name"`
	Id       string  `json:"id"`
	SteamId  string  `json:"steam_id"`
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

type PlayerStats struct {
	NickName       string `json:"nick_name"`
	SteamId        string `json:"steam_id"`
	TeamKills      uint32 `json:"team_kills"`
	Damage         uint32 `json:"damage"`
	Deaths         uint32 `json:"deaths"`
	Kills          uint32 `json:"kills"`
	BodyHits       [9]uint32
	Shots          uint32 `json:"shots"`
	Hits           uint32 `json:"hits"`
	HeadShots      uint32 `json:"head_shots"`
	BombDefusions  uint32 `json:"bomb_defusion"`
	BombDefused    uint32 `json:"bomb_defused"`
	BombPlants     uint32 `json:"bomb_plants"`
	BombExplosions uint32 `json:"bomb_explosions"`
}
