package main

import (
	"encoding/json"
)

type UnitType struct {
	Name string `json:"name"`
}

func UnMarshallUnitTypesJson() ([]UnitType, error) {
	var data map[string][]UnitType

	err := json.Unmarshal([]byte(unitTypesJson), &data)
	if err != nil {
		return nil, err
	}

	return data["unitTypes"], nil
}
