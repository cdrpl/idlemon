package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Unit struct {
	Id       uuid.UUID `json:"id"`
	Template int       `json:"template"`
	Level    int       `json:"level"`
	Stars    int       `json:"stars"`
	IsLocked bool      `json:"isLocked"`
}

// Create a unit with the given template ID.
func CreateUnit(templateID int) Unit {
	return Unit{
		Id:       uuid.New(),
		Template: templateID,
		Level:    1,
		Stars:    1,
	}
}

// Will return a random 1 star unit at level 1.
func RandUnit(dc *DataCache) Unit {
	template := RandUnitTemplateID(dc)
	return CreateUnit(template)
}

// Will insert a unit into the database and return the unit ID.
func InsertUnit(ctx context.Context, tx pgx.Tx, userId uuid.UUID, unit Unit) error {
	query := "INSERT INTO units (id, user_id, template) VALUES ($1, $2, $3)"
	_, err := tx.Exec(ctx, query, unit.Id, userId, unit.Template)

	return err
}

func FindUnits(ctx context.Context, db *pgxpool.Pool, userId uuid.UUID) ([]Unit, error) {
	units := make([]Unit, 0)

	query := "SELECT id, template, level, stars, is_locked FROM units WHERE user_id = $1"
	rows, err := db.Query(ctx, query, userId)
	if err != nil {
		return units, fmt.Errorf("fail to query units table: %w", err)
	}

	for rows.Next() {
		var unit Unit

		err := rows.Scan(&unit.Id, &unit.Template, &unit.Level, &unit.Stars, &unit.IsLocked)
		if err != nil {
			return units, fmt.Errorf("fail to scan into unit: %w", err)
		}

		units = append(units, unit)
	}

	return units, nil
}
