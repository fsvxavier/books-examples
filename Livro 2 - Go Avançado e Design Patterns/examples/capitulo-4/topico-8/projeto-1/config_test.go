package config

import "testing"

func TestAmbienteProducao(t *testing.T) {
	t.Setenv("APP_ENV", "producao")
	if got := Obter().Ambiente; got != "producao" {
		t.Fatalf("Ambiente = %q, quero producao", got)
	}
}

func TestAmbienteTeste(t *testing.T) {
	t.Setenv("APP_ENV", "teste")
	if got := Obter().Ambiente; got != "teste" {
		t.Fatalf("Ambiente = %q, quero teste", got)
	}
}
