package game

import (
	"hlds-games/internal/log"
	"hlds-games/pkg/messages"
)

func newActionMessagePayload(logEvent log.Event) messages.ActionMessagePayload {
	return messages.ActionMessagePayload{
		ActionType: logEvent.Action,
		Player:     newMessagePlayer(logEvent.Player),
		Kill:       newMessageKill(logEvent.Kill),
		Suicide:    newMessageSuicide(logEvent.Suicide),
	}
}

func newMessagePlayer(player log.Player) messages.Player {
	return messages.Player{
		NickName: player.NickName,
		Team:     player.Team,
	}
}

func newMessageKill(kill *log.Kill) *messages.Kill {
	if kill == nil {
		return nil
	}
	return &messages.Kill{
		Victim: newMessagePlayer(kill.Victim),
		Weapon: kill.Weapon,
	}
}

func newMessageSuicide(suicide *log.Suicide) *messages.Suicide {
	if suicide == nil {
		return nil
	}
	return &messages.Suicide{
		Weapon: suicide.Weapon,
	}
}
