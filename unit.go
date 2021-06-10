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

func InsertUnit(db *sql.DB, userID int, templateID int) (Unit, error) {
	unit := Unit{
		UserID:     userID,
		TemplateID: templateID,
		Level:      1,
		Stars:      1,
	}

	query := "INSERT INTO unit (user_id, template_id) VALUES (?, ?)"
	result, err := db.Exec(query, unit.UserID, unit.TemplateID)
	if err != nil {
		return unit, err
	}

	unitID, err := result.LastInsertId()
	if err != nil {
		return unit, err
	}

	unit.ID = int(unitID)
	return unit, nil
}
