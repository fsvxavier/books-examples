//go:build falha

package calc

import "testing"

func TestErrorfContinua(t *testing.T) {
	t.Errorf("primeira falha")
	t.Errorf("segunda falha (o teste continuou)")
}

func TestFatalfPara(t *testing.T) {
	t.Fatalf("falha fatal")
	t.Errorf("nunca executa")
}
