package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
)

type Resource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Insert resources into the table if they don't exist.
func InsertResources(db *sql.DB) {
	resources, err := UnmarshallResourcesJson()
	if err != nil {
		log.Fatalf("insert resources error: %v\n", err)
	}

	for _, resource := range resources {
		var id int

		err := db.QueryRow("SELECT id FROM resource WHERE id = ?", resource.ID).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				query := "INSERT INTO resource (id, name) VALUES (?, ?)"
				_, err := db.Exec(query, resource.ID, resource.Name)
				if err != nil {
					log.Fatalln("insert resources error:", err)
				}

				log.Printf("insert resources %+v\n", resource)
			} else {
				log.Fatalln("insert resources error:", err)
			}
		}
	}
}

// Unmarshall the embeded resourcesJson string.
func UnmarshallResourcesJson() ([]Resource, error) {
	var data map[string][]Resource

	err := json.Unmarshal([]byte(resourcesJson), &data)
	if err != nil {
		return nil, err
	}

	return data["resources"], nil
}
