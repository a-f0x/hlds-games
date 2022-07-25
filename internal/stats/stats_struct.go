package stats

type PlayerStats struct {
	NickName       string
	SteamId        string
	TeamKills      uint32
	Damage         uint32
	Deaths         uint32
	Kills          uint32
	BodyHits       [9]uint32
	Shots          uint32
	Hits           uint32
	HeadShots      uint32
	BombDefusions  uint32
	BombDefused    uint32
	BombPlants     uint32
	BombExplosions uint32
}
