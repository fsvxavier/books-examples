// lib/lib_test.go
package lib

import "testing"

func TestSaudar(t *testing.T) {
	if got := Saudar("Ana"); got != "Oi, Ana" {
		t.Errorf("Saudar = %q", got)
	}
}
