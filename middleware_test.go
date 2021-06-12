package main_test

import (
	"fmt"
	"testing"

	. "github.com/cdrpl/idlemon-server"
)

func TestParseAuthHeader(t *testing.T) {
	idExpect := "1234"
	tokenExpect := "2ac18df3f143f971baa13289fa5a5b04"
	auth := fmt.Sprintf("%v:%v", idExpect, tokenExpect)

	id, token := ParseAuthHeader(auth)

	if id != idExpect {
		t.Errorf("expected id to equal %v but received: %v", idExpect, id)
	}

	if token != tokenExpect {
		t.Errorf("expected token to equal %v but received: %v", tokenExpect, token)
	}
}

func BenchmarkParseAuthHeader(b *testing.B) {
	token := "1:2ac18df3f143f971baa13289fa5a5b04"

	for i := 0; i < b.N; i++ {
		ParseAuthHeader(token)
	}
}
