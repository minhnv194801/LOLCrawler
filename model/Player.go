package model

type Player struct {
	Role             string
	Champion         string
	Kills            string
	Deaths           string
	Assists          string
	KDA              string
	CS               string
	Golds            string
	TotalDamage      string
	TotalDamageTaken string
}

func (player *Player) ToCsvData() [][]string {
	var csvData [][]string
	csvData = append(csvData, []string{player.Role, player.Champion, player.Kills, player.Deaths, player.Assists, player.KDA, player.CS, player.Golds, player.TotalDamage, player.TotalDamageTaken})
	return csvData
}
