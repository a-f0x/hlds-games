package rcon

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var (
	statusRegexp = regexp.MustCompile("hostname: {2}(.+)\nversion : {2}(.+)\ntcp/ip {2}: {2}(.+):(.+)\nmap {5}: {2}(.+) at:(.+)\nplayers : {2}(.+) active (.+)\n")
)

func parseServerStatus(response []byte) (*ServerStatus, error) {
	status := string(response)
	matches := statusRegexp.FindAllStringSubmatch(status, -1)
	if matches == nil {
		return nil, errors.New(fmt.Sprint("Invalid Rcon response"))
	}
	players, err := strconv.ParseInt(matches[0][7], 10, 32)
	port, err := strconv.ParseInt(matches[0][4], 10, 64)

	if err != nil {
		return nil, errors.New(fmt.Sprint("Invalid Rcon response"))
	}
	return &ServerStatus{
		Name:    matches[0][1],
		Host:    matches[0][3],
		Port:    port,
		Players: int32(players),
		Map:     matches[0][5],
	}, nil
}
