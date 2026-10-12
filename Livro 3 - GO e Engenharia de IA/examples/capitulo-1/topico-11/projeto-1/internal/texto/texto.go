// Package texto reúne o que todo pipeline de busca precisa antes de pensar em
// vetores: normalizar, quebrar em palavras e dividir documentos em trechos.
package texto

import (
	"strings"
	"unicode"
)

var semAcento = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c",
)

// Normalizar deixa o texto em minúsculas e sem acentos, para que "Renovação"
// e "renovacao" sejam a mesma coisa na busca.
func Normalizar(s string) string {
	return semAcento.Replace(strings.ToLower(s))
}

// Canonico devolve a forma canônica de uma pergunta para servir de chave:
// minúsculas, sem acentos, sem pontuação, espaços únicos. Diferente de Tokens,
// MANTÉM todas as palavras: "quando abre" e "onde abre" são perguntas
// diferentes, e uma chave que as confundisse devolveria a resposta errada.
func Canonico(s string) string {
	return strings.Join(strings.FieldsFunc(Normalizar(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), " ")
}

// paradas são palavras tão comuns que não ajudam a distinguir documentos.
var paradas = map[string]struct{}{}

func init() {
	for _, p := range strings.Fields(`a o as os um uma uns umas de da do das dos em no na nos nas
		por para com sem e ou que se ao aos como qual quais quando onde ser sao foi
		pode posso preciso tenho meu minha seu sua isso esse essa este esta`) {
		paradas[p] = struct{}{}
	}
}

// Tokens devolve as palavras úteis do texto: normalizadas, sem pontuação e
// sem palavras de parada. A ordem e as repetições são mantidas.
func Tokens(s string) []string {
	campos := strings.FieldsFunc(Normalizar(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := campos[:0]
	for _, c := range campos {
		if _, ok := paradas[c]; ok {
			continue
		}
		out = append(out, c)
	}
	return out
}

// Trecho é um pedaço de documento. Inicio e Fim contam palavras do original,
// o que permite voltar ao ponto exato de onde o trecho saiu.
type Trecho struct {
	Texto  string
	Inicio int // índice da primeira palavra (inclusivo)
	Fim    int // índice da última palavra (exclusivo)
}

// Dividir corta o texto em janelas de no máximo max palavras, com
// sobreposicao palavras repetidas entre trechos vizinhos. A sobreposição
// evita que uma frase cortada ao meio perca o sentido nos dois lados.
func Dividir(s string, max, sobreposicao int) []Trecho {
	if max <= 0 {
		return nil
	}
	if sobreposicao < 0 || sobreposicao >= max {
		sobreposicao = 0
	}
	palavras := strings.Fields(s)
	var trechos []Trecho
	passo := max - sobreposicao
	for ini := 0; ini < len(palavras); ini += passo {
		fim := min(ini+max, len(palavras))
		trechos = append(trechos, Trecho{
			Texto:  strings.Join(palavras[ini:fim], " "),
			Inicio: ini,
			Fim:    fim,
		})
		if fim == len(palavras) {
			break
		}
	}
	return trechos
}
