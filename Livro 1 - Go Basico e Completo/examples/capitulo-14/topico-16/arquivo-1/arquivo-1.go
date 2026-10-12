package lib

// Saudar na versão 2 mudou a assinatura (mudança incompatível): recebe também o idioma.
func Saudar(nome, idioma string) string {
	if idioma == "en" {
		return "Hello, " + nome
	}
	return "Olá, " + nome
}
