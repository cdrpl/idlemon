package main_test

import (
	"testing"

	. "github.com/cdrpl/idlemon"
)

func TestInsertUnit(t *testing.T) {
	db := CreateDBConn()

	user, err := InsertRandUser(db)
	if err != nil {
		t.Fatalf("insert rand user error: %v", err)
	}

	templateID := 2

	unit, err := InsertUnit(db, user.ID, templateID)
	if err != nil {
		t.Fatalf("insert unit error: %v", err)
	}

	// returned unit ID should be greater than 0
	if unit.ID <= 0 {
		t.Error("returned unit have an ID less than or equal to 0")
	}

	if unit.UserID != user.ID {
		t.Errorf("returned unit user ID expect: %v, receive: %v", user.ID, unit.ID)
	}

	if unit.TemplateID != templateID {
		t.Errorf("returned unit template ID expect: %v, receive: %v", templateID, unit.TemplateID)
	}

	if unit.Level != 1 {
		t.Errorf("returned unit level should be 1, receive: %v", unit.Level)
	}

	if unit.Stars != 1 {
		t.Errorf("returned unit stars should be 1, receive: %v", unit.Stars)
	}

	if unit.IsLocked {
		t.Error("returned unit is locked, expect unlocked")
	}

	// unit should be in database
	err = db.QueryRow("SELECT * FROM unit WHERE id = ?", unit.ID).Scan(&unit.ID, &unit.UserID, &unit.TemplateID, &unit.Level, &unit.Stars, &unit.IsLocked)
	if err != nil {
		t.Fatalf("query unit error: %v", err)
	}

	if unit.UserID != user.ID {
		t.Errorf("queried unit user ID expect: %v, receive: %v", user.ID, unit.ID)
	}

	if unit.TemplateID != templateID {
		t.Errorf("queried unit template ID expect: %v, receive: %v", templateID, unit.TemplateID)
	}

	if unit.Level != 1 {
		t.Errorf("queried unit level should be 1, receive: %v", unit.Level)
	}

	if unit.Stars != 1 {
		t.Errorf("queried unit stars should be 1, receive: %v", unit.Stars)
	}

	if unit.IsLocked {
		t.Error("queried unit is locked, expect unlocked")
	}
}
