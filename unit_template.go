package main

import (
	"encoding/json"
)

type UnitTemplate struct {
	TypeID int    `json:"typeId"`
	Name   string `json:"name"`
	Hp     int    `json:"hp"`
	Atk    int    `json:"atk"`
	Def    int    `json:"def"`
	Spd    int    `json:"spd"`
}

func UnMarshallUnitTemplatesJson() ([]UnitTemplate, error) {
	var data map[string][]UnitTemplate

	err := json.Unmarshal([]byte(unitTemplatesJson), &data)
	if err != nil {
		return nil, err
	}

	return data["unitTemplates"], nil
}

// Return a random unit template ID.
func RandUnitTemplateID(dc DataCache) int {
	count := len(dc.UnitTemplates)

	return RandInt(0, count)
}
