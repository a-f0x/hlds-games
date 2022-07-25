package log

import (
	"fmt"
	"log"
	"net"
)

type Receiver struct {
	port         int64
	LogEvent     chan Event
	eventParsers []func(log []byte) *Event
}

func NewReceiver(port int64, c chan Event) *Receiver {
	return &Receiver{
		port:     port,
		LogEvent: c,
		eventParsers: []func(log []byte) *Event{
			playerConnected,
			playerDisconnected,
			kill,
			joinTeam,
			suicide,
		},
	}
}

func (l *Receiver) Start() error {
	udpServer, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s%d", ":", l.port))
	if err != nil {
		return fmt.Errorf("failed to start start udp server for listen event from cs server: %s ", err)
	}

	connection, err := net.ListenUDP("udp4", udpServer)
	defer connection.Close()
	if err != nil {
		return fmt.Errorf("failed to start start udp server for listet event from cs server: %s ", err)
	}

	buffer := make([]byte, 256)
	log.Println("Log receiver started")

	for {
		n, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			return fmt.Errorf("error read bytes from server: %s", err)
		}
		//log.Printf("onEvent event:%s\n", string(buffer))
		if n > 34 {
			//отрезаем всякий хлам ����log L с начала и перенос строки с конца
			l.onEvent(buffer[10 : n-2])
		}
	}

}

func (l *Receiver) onEvent(event []byte) {
	for _, parser := range l.eventParsers {
		logEvent := parser(event)
		if logEvent != nil {
			l.LogEvent <- *logEvent
			return
		}
	}

}
