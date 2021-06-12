package main_test

import (
	"database/sql"
	"testing"

	. "github.com/cdrpl/idlemon-server"
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

func TestRandUnitTemplateID(t *testing.T) {
	db := CreateDBConn()

	for i := 0; i < 100; i++ {
		id, err := RandUnitTemplateID(db)
		if err != nil {
			t.Fatalf("rand unit template ID error: %v", err)
		}

		// id must be valid
		var scan int
		err = db.QueryRow("SELECT id FROM unit_template WHERE id = ?", id).Scan(&scan)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("received an invalid ID: %v", id)
			} else {
				t.Fatalf("query error: %v", err)
			}
		}
	}
}

func BenchmarkRandUnitTemplateID(b *testing.B) {
	db := CreateDBConn()

	for i := 0; i < b.N; i++ {
		_, err := RandUnitTemplateID(db)
		if err != nil {
			b.Fatalf("rand unit template ID error: %v", err)
		}
	}
}
