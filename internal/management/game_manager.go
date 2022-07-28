package management

import (
	"fmt"
	"hlds-games/internal/messages"
	"sync"
	"time"
)

type GameManager struct {
	games      map[string]Game
	mu         sync.Mutex
	externalIp string
}

func NewGameManager(externalIp string, externalPort int64) *GameManager {
	manager := &GameManager{
		games:      make(map[string]Game),
		externalIp: externalIp,
	}
	go func() {
		ticker := time.NewTicker(time.Duration(10) * time.Second)
		for {
			<-ticker.C
			manager.mu.Lock()
			for key, game := range manager.games {
				if game.expired(5) {
					delete(manager.games, key)
				}
			}
			manager.mu.Unlock()
		}
	}()

	return manager
}

func (g *GameManager) RegisterGame(heartBeat messages.Message[messages.HeartBeatMessagePayload]) {
	g.mu.Lock()
	defer g.mu.Unlock()
	game := Game{
		Name:           heartBeat.ServerInfo.GameName,
		GameType:       heartBeat.ServerInfo.GameType,
		GameHost:       heartBeat.Payload.GameHost,
		GamePort:       heartBeat.Payload.GamePort,
		ApiHost:        heartBeat.Payload.ApiHost,
		ApiPort:        heartBeat.Payload.ApiPort,
		ExternalIp:     g.externalIp,
		ExternalPort:   heartBeat.Payload.GamePort,
		registeredTime: heartBeat.Time,
	}
	g.games[game.Key()] = game
}

func (g *GameManager) ListGames() []Game {
	g.mu.Lock()
	defer g.mu.Unlock()
	games := make([]Game, 0, len(g.games))
	for _, value := range g.games {
		games = append(games, value)
	}
	return games
}

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
