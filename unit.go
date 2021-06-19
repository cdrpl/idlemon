package main

type Unit struct {
	ID         int  `json:"id"`
	TemplateID int  `json:"templateId"`
	Level      int  `json:"level"`
	Stars      int  `json:"stars"`
	IsLocked   bool `json:"isLocked"`
}

// Create a unit with the given template ID.
func CreateUnit(templateID int) Unit {
	return Unit{
		TemplateID: templateID,
		Level:      1,
		Stars:      1,
	}
}

// Will return a random 1 star unit at level 1.
func RandUnit(dc DataCache) Unit {
	template := RandUnitTemplateID(dc)
	return CreateUnit(template)
}

// Add unit to the user's units map. Will set the unit's ID and return it.
func AddUnitToUser(user *User, unit Unit) Unit {
	id := user.Data.UnitSerial + 1

	// if ID is taken, increment ID until empty one is found
	_, ok := user.Data.Units[id]
	for ok {
		id++
		_, ok = user.Data.Units[id]
	}

	unit.ID = id
	user.Data.Units[id] = unit
	user.Data.UnitSerial = id

	return unit
}
