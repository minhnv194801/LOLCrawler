package model

type Match struct {
	Team []string
	Game []Game
}

func (match *Match) ToCsvData() [][]string {
	var csvData [][]string
	csvData = append(csvData, []string{match.Team[0] + " vs " + match.Team[1]})
	for _, g := range match.Game {
		csvData = append(csvData, g.ToCsvData()...)
	}
	return csvData
}
