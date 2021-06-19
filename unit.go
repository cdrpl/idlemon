package main

import "errors"

type Unit struct {
	ID         string `json:"id"`
	TemplateID int    `json:"templateId"`
	Level      int    `json:"level"`
	Stars      int    `json:"stars"`
	IsLocked   bool   `json:"isLocked"`
}

// Create a unit with the given template ID.
func CreateUnit(templateID int) (Unit, error) {
	id, err := GenerateToken(UNIT_ID_LEN)
	if err != nil {
		return Unit{}, err
	}

	return Unit{
		ID:         id,
		TemplateID: templateID,
		Level:      1,
		Stars:      1,
	}, nil
}

// Will return a random 1 star unit at level 1.
func RandUnit(dc DataCache) (Unit, error) {
	template := RandUnitTemplateID(dc)
	return CreateUnit(template)
}

// Add unit to the user's units map, will return error if the unit ID is a duplicate.
func AddUnitToUser(user *User, unit Unit) error {
	if _, ok := user.Data.Units[unit.ID]; ok {
		return errors.New("duplicate unit ID")
	}

	user.Data.Units[unit.ID] = unit

	return nil
}
