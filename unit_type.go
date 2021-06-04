package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
)

type UnitType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func InsertUnitTypes(db *sql.DB) {
	var data map[string][]UnitType

	err := json.Unmarshal([]byte(unitTypesJson), &data)
	if err != nil {
		log.Fatalln("insert unit types error:", err)
	}

	// insert unit types if they don't exist
	for _, unitType := range data["unitTypes"] {
		var id int

		err := db.QueryRow("SELECT id FROM unit_type WHERE id = ?", unitType.ID).Scan(&id)
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
