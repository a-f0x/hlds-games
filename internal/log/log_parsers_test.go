package log

import (
	"testing"
)

func TestPlayerConnected(t *testing.T) {
	got := playerConnected([]byte("07/17/2022 - 03:35:43: \"Player<1><VALVE_ID_LAN><>\" entered the game"))
	if got.Action != "player_connected" {
		t.Errorf("Actual: %s, Expected: %s", got.Action, "player_connected")
	}
	if got.Time != 1658028943 {
		t.Errorf("Actual: %d, Expected: %d", got.Time, 1658028943)
	}
	if got.Player.NickName != "Player" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.NickName, "Player")
	}
	if got.Player.Id != "1" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.Id, "Id")
	}
	if got.Player.SteamId != "VALVE_ID_LAN" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.SteamId, "VALVE_ID_LAN")
	}

	if got.Player.Team != nil {
		t.Errorf("Actual: %s, Expected: nil", *got.Player.Team)
	}
	if got.Kill != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Kill)
	}
	if got.Suicide != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Suicide)
	}
}
func TestPlayerDisconnected(t *testing.T) {
	got := playerDisconnected([]byte("07/17/2022 - 03:35:43: \"Player<1><VALVE_ID_LAN><TERRORIST>\" disconnected"))
	if got == nil {
		t.Errorf("Result is nil")
	}
	if got.Time != 1658028943 {
		t.Errorf("Actual: %d, Expected: %d", got.Time, 1658028943)
	}
	if got.Action != "player_disconnected" {
		t.Errorf("Actual: %s, Expected: %s", got.Action, "player_disconnected")
	}
	if got.Player.NickName != "Player" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.NickName, "Player")
	}
	if got.Player.Id != "1" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.Id, "Id")
	}
	if got.Player.SteamId != "VALVE_ID_LAN" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.SteamId, "VALVE_ID_LAN")
	}
	if *got.Player.Team != "TERRORIST" {
		t.Errorf("Actual: %s, Expected: TERRORIST", *got.Player.Team)
	}
	if got.Kill != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Kill)
	}
	if got.Suicide != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Suicide)
	}
}

func TestKill(t *testing.T) {
	got := kill([]byte("07/17/2022 - 03:35:43: \"Player<1><VALVE_ID_LAN><CT>\" killed \"asus<2><VALVE_ID_LAN><TERRORIST>\" with \"usp\""))
	if got == nil {
		t.Errorf("Result is nil")
	}
	if got.Time != 1658028943 {
		t.Errorf("Actual: %d, Expected: %d", got.Time, 1658028943)
	}
	if got.Action != "kill" {
		t.Errorf("Actual: %s, Expected: %s", got.Action, "kill")
	}
	if got.Player.NickName != "Player" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.NickName, "Player")
	}
	if *got.Player.Team != "CT" {
		t.Errorf("Actual: %s, Expected: CT", *got.Player.Team)
	}
	if got.Player.Id != "1" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.Id, "Id")
	}
	if got.Player.SteamId != "VALVE_ID_LAN" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.SteamId, "VALVE_ID_LAN")
	}

	if got.Kill == nil {
		t.Errorf("Kill is nil")
	}
	kill := got.Kill
	if kill.Victim.NickName != "asus" {
		t.Errorf("Actual: %s, Expected: %s", kill.Victim.NickName, "asus")
	}
	if kill.Victim.Id != "2" {
		t.Errorf("Actual: %s, Expected: %s", kill.Victim.Id, "Id")
	}
	if kill.Victim.SteamId != "VALVE_ID_LAN" {
		t.Errorf("Actual: %s, Expected: %s", kill.Victim.SteamId, "VALVE_ID_LAN")
	}

	if *kill.Victim.Team != "TERRORIST" {
		t.Errorf("Actual: %s, Expected: %s", *kill.Victim.Team, "TERRORIST")
	}
	if kill.Weapon != "usp" {
		t.Errorf("Actual: %s, Expected: %s", kill.Weapon, "usp")
	}
	if got.Suicide != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Suicide)
	}

}

func TestSuicide(t *testing.T) {
	got := suicide([]byte("07/17/2022 - 03:35:43: \"Player<1><VALVE_ID_LAN><TERRORIST>\" committed suicide with \"worldspawn\" (world)"))
	if got == nil {
		t.Errorf("Result is nil")
	}
	if got.Time != 1658028943 {
		t.Errorf("Actual: %d, Expected: %d", got.Time, 1658028943)
	}
	if got.Action != "suicide" {
		t.Errorf("Actual: %s, Expected: %s", got.Action, "suicide")
	}
	if got.Player.NickName != "Player" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.NickName, "Player")
	}
	if *got.Player.Team != "TERRORIST" {
		t.Errorf("Actual: %s, Expected: TERRORIST", *got.Player.Team)
	}
	if got.Suicide.Weapon != "worldspawn" {
		t.Errorf("Actual: %s, Expected: %s", got.Suicide.Weapon, "worldspawn")
	}
	if got.Kill != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Kill)
	}
}

func TestJoinTeam(t *testing.T) {
	got := joinTeam([]byte("07/17/2022 - 03:35:43: \"Player<1><VALVE_ID_LAN><>\" joined team \"TERRORIST\""))
	if got == nil {
		t.Errorf("Result is nil")
	}
	if got.Time != 1658028943 {
		t.Errorf("Actual: %d, Expected: %d", got.Time, 1658028943)
	}
	if got.Action != "join_team" {
		t.Errorf("Actual: %s, Expected: %s", got.Action, "join_team")
	}
	if got.Player.NickName != "Player" {
		t.Errorf("Actual: %s, Expected: %s", got.Player.NickName, "Player")
	}
	if *got.Player.Team != "TERRORIST" {
		t.Errorf("Actual: %s, Expected: TERRORIST", *got.Player.Team)
	}
	if got.Kill != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Kill)
	}
	if got.Suicide != nil {
		t.Errorf("Actual: %v, Expected: nil", *got.Suicide)
	}
}

func TestParseTime(t *testing.T) {
	//07/17/2022 - 03:35:43
	got := parseTime([]byte{48, 55, 47, 49, 55, 47, 50, 48, 50, 50, 32, 45, 32, 48, 51, 58, 51, 53, 58, 52, 51})
	want := int64(1658028943)
	if got != want {
		t.Errorf("Actual: %d, Expected: %d", got, want)
	}
}
