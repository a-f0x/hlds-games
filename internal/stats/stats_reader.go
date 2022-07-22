package stats

import (
	"encoding/binary"
	"fmt"
)

// how the statistics file was written
//https://github.com/alliedmodders/amxmodx/blob/ff2b5142f9c4213beee8646d50609cfed315202e/modules/tfcx/CRank.cpp#L308

type Reader struct {
	data   []byte
	offset uint16
}

func NewStatsReader(data []byte) *Reader {
	return &Reader{
		data: data,
	}
}

func (s *Reader) readUint32() uint32 {
	ui := binary.LittleEndian.Uint32(s.data[s.offset:])
	s.offset += 4
	return ui
}

func (s *Reader) readUint16() uint16 {
	ui := binary.LittleEndian.Uint16(s.data[s.offset:])
	s.offset += 2
	return ui
}

func (s *Reader) readString(length uint16) string {
	str := string(s.data[s.offset:(length + s.offset)])
	s.offset += length
	return str
}
func (s *Reader) ReadStats() []PlayerStats {
	stats := func(length uint16) PlayerStats {
		name := s.readString(length)
		length = s.readUint16()
		steamId := s.readString(length)
		teamKills := s.readUint32()
		damage := s.readUint32()
		deaths := s.readUint32()
		kills := s.readUint32()
		shots := s.readUint32()
		hits := s.readUint32()
		headShots := s.readUint32()
		bombDefusions := s.readUint32()
		bombDefused := s.readUint32()
		bombPlants := s.readUint32()
		bombExplosions := s.readUint32()
		var bodyHits [9]uint32
		for i := 0; i < len(bodyHits); i++ {
			bodyHits[i] = s.readUint32()
		}
		fmt.Printf("bodyHits: %v\n", bodyHits)
		return PlayerStats{
			NickName:       name,
			SteamId:        steamId,
			TeamKills:      teamKills,
			Damage:         damage,
			Deaths:         deaths,
			Kills:          kills,
			BodyHits:       bodyHits,
			Shots:          shots,
			Hits:           hits,
			HeadShots:      headShots,
			BombPlants:     bombPlants,
			BombDefusions:  bombDefusions,
			BombDefused:    bombDefused,
			BombExplosions: bombExplosions,
		}

	}
	version := s.readUint16()
	fmt.Printf("stat version: %d\n", version)
	arr := make([]PlayerStats, 0)

	length := s.readUint16()
	for length > 0 {
		r := stats(length)
		arr = append(arr, r)
		length = s.readUint16()
	}

	return arr
}
