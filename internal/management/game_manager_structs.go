package management

import (
	"fmt"
	"time"
)

type Game struct {
	Name           string
	GameType       string
	GameHost       string
	GamePort       int64
	ApiHost        string
	ApiPort        int64
	Players        int32
	Map            string
	ExternalIp     string
	ExternalPort   int64
	registeredTime int64
}

func (g Game) Key() string {
	return fmt.Sprintf("%s%s", g.GameType, g.Name)
}
func (g Game) expired(diffSec int64) bool {
	if time.Now().Unix()-g.registeredTime > diffSec {
		return true
	}
	return false
}
func (g Game) getExternalUrl() string {
	return fmt.Sprintf("%s:%d", g.ExternalIp, g.ExternalPort)
}
func (g Game) getApiUrl() string {
	return fmt.Sprintf("%s:%d", g.ApiHost, g.ApiPort)

}

func (g Game) String() string {
	return fmt.Sprintf("GameType: %s\nName: %s\nMap: %s\nPlayers: %d\nip: %s\n", g.GameType, g.Name, g.Map, g.Players, g.getExternalUrl())

}
