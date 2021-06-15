package main_test

import (
	"database/sql"
	"errors"
	"testing"

	. "github.com/cdrpl/idlemon-server"
)

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
			if errors.Is(err, sql.ErrNoRows) {
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
