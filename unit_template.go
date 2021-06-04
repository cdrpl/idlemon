package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
)

type UnitTemplate struct {
	ID     int    `json:"id"`
	TypeID int    `json:"typeId"`
	Name   string `json:"name"`
	Hp     int    `json:"hp"`
	Atk    int    `json:"atk"`
	Def    int    `json:"def"`
	Spd    int    `json:"spd"`
}

// Fetch all the unit templates from the table.
func UnitTemplates(db *sql.DB) ([]UnitTemplate, error) {
	templates := make([]UnitTemplate, 0)

	rows, err := db.Query("SELECT * FROM unit_template")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		template := UnitTemplate{}

		err := rows.Scan(&template.ID, &template.TypeID, &template.Name, &template.Hp, &template.Atk, &template.Def, &template.Spd)
		if err != nil {
			return nil, err
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func InsertUnitTemplates(db *sql.DB) {
	var data map[string][]UnitTemplate

	err := json.Unmarshal([]byte(unitTemplatesJson), &data)
	if err != nil {
		log.Fatalln("unmarshall unit templates error:", err)
	}

	// insert unit templates if they don't exist
	for _, template := range data["unitTemplates"] {
		var id int

		err := db.QueryRow("SELECT id FROM unit_template WHERE id = ?", template.ID).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				query := "INSERT INTO unit_template (type_id, name, hp, atk, def, spd) VALUES (?, ?, ?, ?, ?, ?)"
				_, err := db.Exec(query, template.TypeID, template.Name, template.Hp, template.Atk, template.Def, template.Spd)
				if err != nil {
					log.Fatalln("insert unit templates error:", err)
				}

				log.Printf("insert unit template %+v\n", template)
			} else {
				log.Fatalln("insert unit templates error:", err)
			}
		}
	}
}
