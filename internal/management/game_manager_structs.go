package management

import (
	"bytes"
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
	return fmt.Sprintf("%s%d", g.GameHost, g.GamePort)
}
func (g Game) expired(diffSec int64) bool {
	if time.Now().Unix()-g.registeredTime > diffSec {
		return true
	}
	return false
}
func (g Game) GetExternalUrl() string {
	return fmt.Sprintf("%s:%d", g.ExternalIp, g.ExternalPort)
}
func (g Game) GetApiUrl() string {
	return fmt.Sprintf("%s:%d", g.ApiHost, g.ApiPort)

}

func (g Game) String() string {
	return fmt.Sprintf("GameType: %s\nName: %s\nMap: %s\nPlayers: %d\nIp: %s\n", g.GameType, g.Name, g.Map, g.Players, g.GetExternalUrl())
}

func BuildGamesText(games []Game) string {
	gl := len(games)
	var buffer bytes.Buffer
	for i, game := range games {
		buffer.WriteString(game.String())
		if i < gl-1 {
			buffer.WriteString("\n")
		}
	}
	return buffer.String()
}
