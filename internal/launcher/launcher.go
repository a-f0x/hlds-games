package launcher

import (
	"fmt"
	"hlds-games/internal/common"
	"hlds-games/internal/config"
	logReceiver "hlds-games/internal/log"
	"hlds-games/internal/messages"
	"hlds-games/internal/rcon"
	"log"
	"os/exec"
	"time"
)

type Launcher struct {
	hldsGameConfig   *config.HldsGameConfig
	rcon             *rcon.Rcon
	logReceiver      *logReceiver.Receiver
	heartBeatChannel chan messages.Message[messages.HeartBeatMessagePayload]
	actionChannel    chan messages.Message[messages.ActionMessagePayload]
	isConnected      *common.AtomicBool
}

func NewLauncher(hldsGameConfig *config.HldsGameConfig) *Launcher {
	return &Launcher{
		hldsGameConfig:   hldsGameConfig,
		logReceiver:      logReceiver.NewReceiver(hldsGameConfig.LogReceiverPort, make(chan logReceiver.Event)),
		rcon:             rcon.NewRcon(hldsGameConfig.Host, hldsGameConfig.HldsGamePort, hldsGameConfig.RconPassword),
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
func (l *Launcher) RunGame(gameMap string) (
	<-chan messages.Message[messages.HeartBeatMessagePayload],
	<-chan messages.Message[messages.ActionMessagePayload],
) {
	l.startGame(gameMap)
	isOnline := make(chan messages.ServerInfo)
	go func() {
		serverInfo := <-isOnline
		l.startLog(serverInfo)
	}()
	l.heartBeat(2, isOnline)
	return l.heartBeatChannel, l.actionChannel
}

func (l *Launcher) startLog(info messages.ServerInfo) {
	listenLogEvents := func(info messages.ServerInfo) {
		log.Println("Listen to log events")
		for {
			logEvent := <-l.logReceiver.LogEvent
			message := messages.Message[messages.ActionMessagePayload]{
				ServerInfo:  info,
				Time:        logEvent.Time,
				MessageType: messages.Action,
				Payload:     newActionMessagePayload(logEvent),
			}
			l.actionChannel <- message
		}
	}
	go listenLogEvents(info)
	go func() {
		err := l.logReceiver.Start()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()
}

func (l *Launcher) heartBeat(initialDelaySec int64, isOnline chan messages.ServerInfo) {
	status := func() *rcon.ServerStatus {
		response, err := l.rcon.GetServerStatus()
		if err != nil {
			log.Printf("Fail heart beat: %s", err)
			return nil
		}
		return response
	}
	go func() {
		ticker := time.NewTicker(time.Duration(2) * time.Second)
		for {
			<-ticker.C
			serverStatus := status()
			if serverStatus != nil {
				serverInfo := messages.ServerInfo{
					GameType: l.hldsGameConfig.GameType,
					GameName: serverStatus.Name,
				}
				l.heartBeatChannel <- messages.Message[messages.HeartBeatMessagePayload]{
					ServerInfo:  serverInfo,
					Time:        time.Now().Unix(),
					MessageType: messages.HeartBeat,
					Payload: messages.HeartBeatMessagePayload{
						Players:  serverStatus.Players,
						Map:      serverStatus.Map,
						GameHost: serverStatus.Host,
						GamePort: serverStatus.Port,
					},
				}
				if !l.isConnected.GetAndSet(true) {
					isOnline <- serverInfo
					close(isOnline)
				}
			}
		}
	}()
}

func (l *Launcher) startGame(gameMap string) {
	halfLife := "half-life"
	counterStrike := "cs-classic"
	counterStrikeDeadMatch := "cs-dm"
	games := []string{halfLife, counterStrike, counterStrikeDeadMatch}

	var command string
	switch l.hldsGameConfig.GameType {
	case counterStrike, counterStrikeDeadMatch:
		command = fmt.Sprintf("./hlds_run -game cstrike +rcon_password %s +port %d +maxplayers 32 +map %s +logaddress 127.0.0.1 %d",
			l.hldsGameConfig.RconPassword, l.hldsGameConfig.HldsGamePort, gameMap, l.hldsGameConfig.LogReceiverPort)
	case halfLife:
		command = fmt.Sprintf("./hlds_run -game valve +rcon_password %s +port %d +maxplayers 32 +map %s +logaddress 127.0.0.1 %d",
			l.hldsGameConfig.RconPassword, l.hldsGameConfig.HldsGamePort, gameMap, l.hldsGameConfig.LogReceiverPort)
	default:
		log.Fatalf(fmt.Sprintf("Unknown game type: %s. Available type is %v", l.hldsGameConfig.GameType, games))
	}
	go func() {
		cmd := exec.Command("sh", "-c", command)
		_, err := cmd.Output()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()
}
