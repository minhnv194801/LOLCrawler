package model

import (
	"strconv"
)

type Match struct {
	Team []string
	Game []Game
}

func (match *Match) ToCsvData() [][]string {
	var csvData [][]string
	csvData = append(csvData, []string{match.Team[0] + " vs " + match.Team[1]})
	for index, g := range match.Game {
		csvData = append(csvData, []string{"Game " + strconv.Itoa(index+1)})
		csvData = append(csvData, g.ToCsvData()...)
	}
	return csvData
}
