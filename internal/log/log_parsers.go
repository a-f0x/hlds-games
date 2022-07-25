package log

import (
	"hlds-games/internal/common"
	"regexp"
	"time"
)

type eventMatch struct {
	time    int64
	matches []string
}

const (
	timeLayout = "01/02/2006 - 3:04:05"
)

var (
	/*
	   event: [48 55 47 49 55 47 50 48 50 50 32 45 32 48 51 58 51 51 58 53 52 58 32 34 80 108 97 121 101 114 60 49 62 60 86 65 76 86 69 95 73 68 95 76 65 78 62 60 62 34 32 101 110 116 101 114 101 100 32 116 104 101 32 103 97 109 101]
	   str: 07/17/2022 - 03:35:43: "Player<1><VALVE_ID_LAN><>" entered the game
	*/
	playerConnectedRegexp = regexp.MustCompile("\"(.+)<(\\d+)><(.+)><(.*)>\" entered the game")

	/*
		event: [48 55 47 49 55 47 50 48 50 50 32 45 32 48 51 58 51 57 58 48 49 58 32 34 80 108 97 121 101 114 60 49 62 60 86 65 76 86 69 95 73 68 95 76 65 78 62 60 62 34 32 100 105 115 99 111 110 110 101 99 116 101 100]
		str: 07/17/2022 - 03:39:01: "Player<1><VALVE_ID_LAN><>" disconnected
	*/
	playerDisconnectedRegexp = regexp.MustCompile("\"(.+)<(\\d+)><(.+)><(.*)>\" disconnected")

	/*
		event:[48 55 47 49 55 47 50 48 50 50 32 45 32 49 49 58 51 57 58 51 55 58 32 34 80 108 97 121 101 114 60 50 62 60 86 65 76 86 69 95 73 68 95 76 65 78 62 60 62 34 32 106 111 105 110 101 100 32 116 101 97 109 32 34 84 69 82 82 79 82 73 83 84 34]
		str: 07/17/2022 - 11:39:37: "Player<2><VALVE_ID_LAN><>" joined team "TERRORIST"
	*/
	playerJoinedTeamRegexp = regexp.MustCompile("\"(.+)<(\\d+)><(.+)><(.*)>\" joined team \"(.+)\"")

	/*
		event: [48 55 47 49 55 47 50 48 50 50 32 45 32 48 52 58 49 49 58 51 53 58 32 34 80 108 97 121 101 114 60 49 62 60 86 65 76 86 69 95 73 68 95 76 65 78 62 60 67 84 62 34 32 107 105 108 108 101 100 32 34 97 115 117 115 60 50 62 60 86 65 76 86 69 95 73 68 95 76 65 78 62 60 84 69 82 82 79 82 73 83 84 62 34 32 119 105 116 104 32 34 117 115 112 34]
		str: 07/17/2022 - 04:11:35: "Player<1><VALVE_ID_LAN><CT>" killed "asus<2><VALVE_ID_LAN><TERRORIST>" with "usp"
	*/
	killRegexp = regexp.MustCompile("\"(.+)<(\\d+)><(.+)><([A-Z]+)>\" killed \"(.+)<(\\d+)><(.+)><([A-Z]+)>\" with \"(.+)\"")

	/*
		event: [48 55 47 49 55 47 50 48 50 50 32 45 32 49 49 58 51 51 58 50 56 58 32 34 80 108 97 121 101 114 60 49 62 60 86 65 76 86 69 95 73 68 95 76 65 78 62 60 84 69 82 82 79 82 73 83 84 62 34 32 99 111 109 109 105 116 116 101 100 32 115 117 105 99 105 100 101 32 119 105 116 104 32 34 119 111 114 108 100 115 112 97 119 110 34 32 40 119 111 114 108 100 41]
		str: 07/17/2022 - 11:33:28: "Player<1><VALVE_ID_LAN><TERRORIST>" committed suicide with "worldspawn" (world)
	*/
	suicideRegexp = regexp.MustCompile("\"(.+)<(\\d+)><(.+)><(.+)>\" committed suicide with \"(.+)\"")
)

var playerConnected = func(event []byte) *Event {
	eventMatch := match(event, playerConnectedRegexp)
	if eventMatch != nil {
		return &Event{
			Action: ActionPlayerConnected,
			Time:   eventMatch.time,
			Player: Player{
				NickName: eventMatch.matches[1],
				Id:       eventMatch.matches[2],
				SteamId:  eventMatch.matches[3],
				Team:     common.StringOrNil(eventMatch.matches[4]),
			},
		}
	}
	return nil
}

var playerDisconnected = func(event []byte) *Event {
	eventMatch := match(event, playerDisconnectedRegexp)
	if eventMatch != nil {
		return &Event{
			Action: ActionPayerDisconnected,
			Time:   eventMatch.time,
			Player: Player{
				NickName: eventMatch.matches[1],
				Id:       eventMatch.matches[2],
				SteamId:  eventMatch.matches[3],
				Team:     common.StringOrNil(eventMatch.matches[4]),
			},
		}
	}
	return nil
}

var kill = func(event []byte) *Event {
	eventMatch := match(event, killRegexp)
	if eventMatch != nil {
		return &Event{
			Action: ActionKill,
			Time:   eventMatch.time,
			Player: Player{
				NickName: eventMatch.matches[1],
				Id:       eventMatch.matches[2],
				SteamId:  eventMatch.matches[3],
				Team:     common.StringOrNil(eventMatch.matches[4]),
			},
			Kill: &Kill{
				Victim: Player{
					NickName: eventMatch.matches[5],
					Id:       eventMatch.matches[6],
					SteamId:  eventMatch.matches[7],
					Team:     common.StringOrNil(eventMatch.matches[8]),
				},
				Weapon: eventMatch.matches[9],
			},
		}
	}
	return nil
}

var suicide = func(event []byte) *Event {
	eventMatch := match(event, suicideRegexp)
	if eventMatch != nil {
		return &Event{
			Action: ActionSuicide,
			Time:   eventMatch.time,
			Player: Player{
				NickName: eventMatch.matches[1],
				Id:       eventMatch.matches[2],
				SteamId:  eventMatch.matches[3],
				Team:     common.StringOrNil(eventMatch.matches[4]),
			},
			Suicide: &Suicide{Weapon: eventMatch.matches[5]},
		}
	}
	return nil
}

var joinTeam = func(event []byte) *Event {
	eventMatch := match(event, playerJoinedTeamRegexp)
	if eventMatch != nil {
		return &Event{
			Action: ActionJoinTeam,
			Time:   eventMatch.time,
			Player: Player{
				NickName: eventMatch.matches[1],
				Id:       eventMatch.matches[2],
				SteamId:  eventMatch.matches[3],
				Team:     common.StringOrNil(eventMatch.matches[5]),
			},
		}
	}
	return nil
}

/*
По 21 байт идет таймштамп события
Пейлоад начинается с 23 байта
Весь пейлоад который подходит под наши евенты начинается с ковычки (" = 34 в байтовом представлении)
*/
func match(event []byte, regexp *regexp.Regexp) *eventMatch {
	var quoteByte byte = 34
	if event[23] != quoteByte {
		return nil
	}
	str := string(event[23:])
	matches := regexp.FindAllStringSubmatch(str, -1)
	if matches == nil {
		return nil
	}
	return &eventMatch{
		time:    parseTime(event[0:21]),
		matches: matches[0],
	}
}

func parseTime(timeBytes []byte) int64 {
	t, _ := time.Parse(timeLayout, string(timeBytes))
	return t.Unix()
}
