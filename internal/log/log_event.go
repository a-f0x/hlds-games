package log

const (
	ActionPlayerConnected   = "player_connected"
	ActionPayerDisconnected = "player_disconnected"
	ActionKill              = "kill"
	ActionSuicide           = "suicide"
	ActionJoinTeam          = "join_team"
)

type Event struct {
	Action  string
	Time    int64
	Player  Player
	Kill    *Kill
	Suicide *Suicide
}
type Player struct {
	NickName string
	Id       string
	SteamId  string
	Team     *string
}

type Kill struct {
	Victim Player
	Weapon string
}

type Suicide struct {
	Weapon string
}
