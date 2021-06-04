package main_test

import (
	"testing"

	. "github.com/cdrpl/idlemon"
)

func BenchmarkCreateDBConn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateDBConn()
	}
}
