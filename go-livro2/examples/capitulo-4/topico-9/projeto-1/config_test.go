package config

import "testing"

func fonte(ambiente string) func(string) string {
	return func(chave string) string {
		if chave == "APP_ENV" {
			return ambiente
		}
		return ""
	}
}

func TestAmbienteProducao(t *testing.T) {
	t.Parallel()
	if got := Carregar(fonte("producao")).Ambiente; got != "producao" {
		t.Fatalf("Ambiente = %q, quero producao", got)
	}
}

func TestAmbienteTeste(t *testing.T) {
	t.Parallel()
	if got := Carregar(fonte("teste")).Ambiente; got != "teste" {
		t.Fatalf("Ambiente = %q, quero teste", got)
	}
}
