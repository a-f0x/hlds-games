package game

import (
	logReceiver "hlds-games/internal/log"
	"hlds-games/pkg/common"
	"hlds-games/pkg/messages"
	"hlds-games/pkg/rcon"
	"log"
	"time"
)

const (
	logReceiverPort int64  = 9654
	hldsServerHost  string = "127.0.0.1"
)

type Launcher struct {
	hldsServerPort int64
	gameType       string
	rconPassword   string
	rcon           *rcon.Rcon
	logReceiver    *logReceiver.Receiver
	HeartBeat      chan *messages.Message[messages.HeartBeatMessagePayload]
	Action         chan *messages.Message[messages.ActionMessagePayload]
	isConnected    *common.AtomicBool
}

func NewLauncher(
	hldsServerPort int64,
	gameType string,
	rconPassword string,
	heartBeat chan *messages.Message[messages.HeartBeatMessagePayload],
	action chan *messages.Message[messages.ActionMessagePayload],
) *Launcher {
	return &Launcher{
		hldsServerPort: hldsServerPort,
		gameType:       gameType,
		rconPassword:   rconPassword,
		logReceiver:    logReceiver.NewReceiver(logReceiverPort, make(chan logReceiver.Event)),
		rcon:           rcon.NewRcon(hldsServerHost, hldsServerPort, rconPassword),
		HeartBeat:      heartBeat,
		Action:         action,
		isConnected:    new(common.AtomicBool),
	}
}

// RunGame
//	Суть запуска:
//	Запустить игру
//	Дождаться когда сервер будет готов к игре с помощью функции heartBeat
// После этого поднять logReceiver
func (a *Launcher) RunGame() {
	a.startGame()
	isOnline := make(chan messages.ServerInfo)
	go func() {
		serverInfo := <-isOnline
		a.startLog(serverInfo)
	}()
	a.heartBeat(2, isOnline)
}

func (a *Launcher) startGame() {
	err := newHldsGames(a.rconPassword, a.hldsServerPort, logReceiverPort).runGame(a.gameType)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func (a *Launcher) startLog(info messages.ServerInfo) {
	listenLogEvents := func(info messages.ServerInfo) {
		log.Println("Listen to log events")
		for {
			logEvent := <-a.logReceiver.LogEvent
			message := &messages.Message[messages.ActionMessagePayload]{
				ServerInfo:  info,
				Time:        logEvent.Time,
				MessageType: messages.Action,
				Payload:     newActionMessagePayload(logEvent),
			}
			a.Action <- message
		}
	}
	go listenLogEvents(info)
	go func() {
		err := a.logReceiver.Start()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()
}

func (a *Launcher) heartBeat(initialDelaySec int64, isOnline chan messages.ServerInfo) {
	status := func() *rcon.ServerStatus {
		response, err := a.rcon.GetServerStatus()
		if err != nil {
			log.Printf("Fail heart beat: %s", err)
			return nil
		}
		return response
	}
	go func() {
		initDelay := time.Duration(initialDelaySec) * time.Second
		periodDelay := time.Duration(2) * time.Second
		time.Sleep(initDelay)
		for {
			time.Sleep(periodDelay)
			serverStatus := status()
			if serverStatus != nil {
				serverInfo := messages.ServerInfo{
					Game: a.gameType,
					Name: serverStatus.Name,
					Host: serverStatus.Host,
				}
				a.HeartBeat <- &messages.Message[messages.HeartBeatMessagePayload]{
					ServerInfo:  serverInfo,
					Time:        time.Now().Unix(),
					MessageType: messages.HeartBeat,
					Payload: messages.HeartBeatMessagePayload{
						Players: serverStatus.Players,
						Map:     serverStatus.Map,
					},
				}
				if !a.isConnected.GetAndSet(true) {
					isOnline <- serverInfo
					close(isOnline)
				}
			}
		}
	}()
}
