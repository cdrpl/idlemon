package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
)

type UnitType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// insert unit types from the embeded json file.
func InsertUnitTypes(ctx context.Context, db *sql.DB, dc *DataCache) {
	for _, unitType := range dc.UnitTypes {
		var id int

		err := db.QueryRowContext(ctx, "SELECT id FROM unit_type WHERE id = ?", unitType.ID).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_, err := db.Exec("INSERT INTO unit_type (id, name) VALUES (?, ?)", unitType.ID, unitType.Name)
				if err != nil {
					log.Fatalln("insert unit types error:", err)
				}

				log.Printf("insert unit type %+v\n", unitType)
			} else {
				log.Fatalln("insert unit types error:", err)
			}
		}
	}
}

func UnMarshallUnitTypesJson() ([]UnitType, error) {
	var data map[string][]UnitType

	err := json.Unmarshal([]byte(unitTypesJson), &data)
	if err != nil {
		return nil, err
	}

	return data["unitTypes"], nil
}
