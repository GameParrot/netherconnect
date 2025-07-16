package utils

import (
	"errors"
	"strconv"
)

type ServerStatus struct {
	// ServerName is the name or MOTD of the server, as shown in the server list.
	ServerName string
	// ProtocolID is the main protocol id accepted by the server
	ProtocolID int
	// Version is the main version that the server supports
	Version string
	// PlayerCount is the current amount of players displayed in the list.
	PlayerCount int
	// MaxPlayers is the maximum amount of players in the server. If set to 0, MaxPlayers is set to
	// PlayerCount + 1.
	MaxPlayers int
	// RaknetID is the server's Raknet ID
	RaknetID int
	// WorldName is the world name of the server
	WorldName string
	// GameMode is the gamemode of the server
	GameMode string
}

func splitPong(s string) []string {
	var runes []rune
	var tokens []string
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\\':
			inEscape = true
		case r == ';':
			tokens = append(tokens, string(runes))
			runes = runes[:0]
		case inEscape:
			inEscape = false
			fallthrough
		default:
			runes = append(runes, r)
		}
	}
	return append(tokens, string(runes))
}

// parsePongData parses the unconnected pong data passed into the relevant fields of a ServerStatus struct.
func parsePongData(pong []byte) (ServerStatus, error) {
	frag := splitPong(string(pong))
	if len(frag) < 7 {
		return ServerStatus{}, errors.New("invalid pong data")

	}
	serverName := frag[1]
	protocol, _ := strconv.Atoi(frag[2])
	version := frag[3]
	online, err := strconv.Atoi(frag[4])
	if err != nil {
		return ServerStatus{}, errors.New("invalid player count")

	}
	max, err := strconv.Atoi(frag[5])
	if err != nil {
		return ServerStatus{}, errors.New("invalid max player count")
	}
	raknetID, _ := strconv.Atoi(frag[6])
	mapName := ""
	gamemode := "Unknown"
	if len(frag) >= 9 {
		mapName = frag[7]
		gamemode = frag[8]
	}
	return ServerStatus{
		ServerName:  serverName,
		ProtocolID:  protocol,
		Version:     version,
		PlayerCount: online,
		MaxPlayers:  max,
		RaknetID:    raknetID,
		WorldName:   mapName,
		GameMode:    gamemode,
	}, nil
}
