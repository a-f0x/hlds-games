package rcon

type ServerStatus struct {
	Name    string
	Host    string
	Port    int64
	Players int32
	Map     string
}
