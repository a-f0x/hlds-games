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
