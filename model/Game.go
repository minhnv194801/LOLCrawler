package model

import (
	"reflect"
)

type Game struct {
	Team    []string
	Time    string
	Players []Player
}

func (game *Game) ToCsvData() [][]string {
	var csvData [][]string
	csvData = append(csvData, []string{"Time", game.Time})
	csvData = append(csvData, []string{game.Team[0]})
	val := reflect.ValueOf(&game.Players[0]).Elem()
	var playerCsvMetadata []string
	for i := 0; i < val.NumField(); i++ {
		playerCsvMetadata = append(playerCsvMetadata, val.Type().Field(i).Name)
	}
	csvData = append(csvData, playerCsvMetadata)
	for i := 0; i < 5; i++ {
		csvData = append(csvData, game.Players[i].ToCsvData()...)
	}
	csvData = append(csvData, []string{game.Team[1]})
	csvData = append(csvData, playerCsvMetadata)
	for i := 5; i < 10; i++ {
		csvData = append(csvData, game.Players[i].ToCsvData()...)
	}

	return csvData
}
