package game

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
)

const (
	halfLife               = "half-life"
	counterStrike          = "cs-classic"
	counterStrikeDeadMatch = "cs-dm"
)

type hldsGames struct {
	games           []string
	rconPassword    string
	hldsServerPort  int64
	logReceiverPort int64
}

func newHldsGames(rconPassword string, hldsServerPort int64, logReceiverPort int64) *hldsGames {
	return &hldsGames{
		games:           []string{halfLife, counterStrike, counterStrikeDeadMatch},
		rconPassword:    rconPassword,
		hldsServerPort:  hldsServerPort,
		logReceiverPort: logReceiverPort,
	}
}
func (h *hldsGames) runGame(gameType string, gMap string) error {
	var command string
	switch gameType {
	case counterStrike, counterStrikeDeadMatch:
		command = fmt.Sprintf("./hlds_run -game cstrike +rcon_password %s +port %d +maxplayers 32 +map %s +logaddress 127.0.0.1 %d",
			h.rconPassword, h.hldsServerPort, gMap, h.logReceiverPort)
	case halfLife:
		command = fmt.Sprintf("./hlds_run -game valve +rcon_password %s +port %d +maxplayers 32 +map %s +logaddress 127.0.0.1 %d",
			h.rconPassword, h.hldsServerPort, gMap, h.logReceiverPort)
	default:
		return errors.New(fmt.Sprintf("Unknown game type: %s. Available type is %v", gameType, h.games))
	}
	go func() {
		cmd := exec.Command("sh", "-c", command)
		_, err := cmd.Output()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()
	return nil
}
