package main_test

import (
	"testing"

	. "github.com/cdrpl/idlemon"
)

func TestUnitTemplateCount(t *testing.T) {
	db := CreateDBConn()

	count, err := UnitTemplateCount(db)
	if err != nil {
		t.Fatalf("unit template count error: %v", err)
	}

	// get expected template count
	var expected int
	err = db.QueryRow("SELECT COUNT(id) FROM unit_template").Scan(&expected)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	if count != expected {
		t.Errorf("expected count to equal %v, received %v", expected, count)
	}
}
