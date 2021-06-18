package main

type Unit struct {
	ID         string `json:"id"`
	TemplateID int    `json:"templateId"`
	Level      int    `json:"level"`
	Stars      int    `json:"stars"`
	IsLocked   bool   `json:"isLocked"`
}

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
