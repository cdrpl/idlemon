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
	var data map[string][]Resource

	err := json.Unmarshal([]byte(resourcesJson), &data)
	if err != nil {
		log.Fatalln("unmarshall resources json error:", err)
	}

	for _, resource := range data["resources"] {
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
