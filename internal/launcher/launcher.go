package launcher

import (
	"hlds-games/internal/common"
	logReceiver "hlds-games/internal/log"
	"hlds-games/internal/messages"
	"hlds-games/internal/rcon"
	"log"
	"time"
)

const (
	hldsServerHost string = "127.0.0.1"
)

type Launcher struct {
	hldsServerPort   int64
	logReceiverPort  int64
	gMap             string
	gameType         string
	rconPassword     string
	rcon             *rcon.Rcon
	logReceiver      *logReceiver.Receiver
	heartBeatChannel chan messages.Message[messages.HeartBeatMessagePayload]
	actionChannel    chan messages.Message[messages.ActionMessagePayload]
	isConnected      *common.AtomicBool
}

func NewLauncher(
	hldsServerPort int64,
	gMap string,
	gameType string,
	rconPassword string,
) *Launcher {
	logReceiverPort := hldsServerPort - 100
	return &Launcher{
		hldsServerPort:   hldsServerPort,
		logReceiverPort:  logReceiverPort,
		gMap:             gMap,
		gameType:         gameType,
		rconPassword:     rconPassword,
		logReceiver:      logReceiver.NewReceiver(logReceiverPort, make(chan logReceiver.Event)),
		rcon:             rcon.NewRcon(hldsServerHost, hldsServerPort, rconPassword),
		heartBeatChannel: make(chan messages.Message[messages.HeartBeatMessagePayload]),
		actionChannel:    make(chan messages.Message[messages.ActionMessagePayload]),
		isConnected:      new(common.AtomicBool),
	}
}

// RunGame
//	Суть запуска:
//	Запустить игру
//	Дождаться когда сервер будет готов к игре с помощью функции heartBeat
// После этого поднять logReceiver
func (a *Launcher) RunGame() (
	<-chan messages.Message[messages.HeartBeatMessagePayload],
	<-chan messages.Message[messages.ActionMessagePayload],
) {
	a.startGame()
	isOnline := make(chan messages.ServerInfo)
	go func() {
		serverInfo := <-isOnline
		a.startLog(serverInfo)
	}()
	a.heartBeat(2, isOnline)
	return a.heartBeatChannel, a.actionChannel
}

func (a *Launcher) startGame() {
	err := newHldsGames(a.rconPassword, a.hldsServerPort, a.logReceiverPort).runGame(a.gameType, a.gMap)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func (a *Launcher) startLog(info messages.ServerInfo) {
	listenLogEvents := func(info messages.ServerInfo) {
		log.Println("Listen to log events")
		for {
			logEvent := <-a.logReceiver.LogEvent
			message := messages.Message[messages.ActionMessagePayload]{
				ServerInfo:  info,
				Time:        logEvent.Time,
				MessageType: messages.Action,
				Payload:     newActionMessagePayload(logEvent),
			}
			a.actionChannel <- message
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
				a.heartBeatChannel <- messages.Message[messages.HeartBeatMessagePayload]{
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
