package main

import "database/sql"

type Unit struct {
	ID         int  `json:"id"`
	UserID     int  `json:"userId"`
	TemplateID int  `json:"templateId"`
	Level      int  `json:"level"`
	Stars      int  `json:"stars"`
	IsLocked   bool `json:"isLocked"`
}

// Find all units owned by a user.
func Units(db *sql.DB, userID int) ([]Unit, error) {
	units := make([]Unit, 0)

	rows, err := db.Query("SELECT * FROM unit WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		unit := Unit{}

		err := rows.Scan(&unit.ID, &unit.UserID, &unit.TemplateID, &unit.Level, &unit.Stars, &unit.IsLocked)
		if err != nil {
			return nil, err
		}

		units = append(units, unit)
	}

	return units, nil
}
