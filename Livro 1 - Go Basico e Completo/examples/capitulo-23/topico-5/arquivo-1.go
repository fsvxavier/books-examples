package main

import (
	"fmt"
	"regexp"
)

// compilar UMA vez, no escopo do pacote
var (
	reData  = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	reLog   = regexp.MustCompile(`^(?P<ip>\S+) \S+ \S+ \[(?P<quando>[^\]]+)\] "(?P<metodo>[A-Z]+) (?P<caminho>\S+) \S+" (?P<status>\d{3}) (?P<bytes>\d+|-)$`)
	reCPF   = regexp.MustCompile(`\b\d{3}\.\d{3}\.\d{3}-\d{2}\b`)
	reEmail = regexp.MustCompile(`(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`)
)

func main() {
	// Match e Find
	fmt.Println("MatchString:      ", reData.MatchString("vence em 2026-10-04"))
	fmt.Println("FindString:       ", reData.FindString("de 2026-10-04 a 2026-12-31"))
	fmt.Println("FindAllString:    ", reData.FindAllString("de 2026-10-04 a 2026-12-31", -1))
	fmt.Println("FindStringSubmatch", reData.FindStringSubmatch("de 2026-10-04"))
	fmt.Println("FindStringIndex:  ", reData.FindStringIndex("de 2026-10-04"))

	// grupos nomeados
	linha := `203.0.113.7 - - [04/Oct/2026:10:15:32 -0300] "GET /api/pedidos/42 HTTP/1.1" 200 1532`
	m := reLog.FindStringSubmatch(linha)
	for i, nome := range reLog.SubexpNames() {
		if nome != "" {
			fmt.Printf("  %-8s = %s\n", nome, m[i])
		}
	}
	fmt.Println("SubexpIndex(status):", reLog.SubexpIndex("status"))

	// substituicao: $1, ${nome}, e funcao
	fmt.Println("Replace $:        ", reData.ReplaceAllString("2026-10-04", "$3/$2/$1"))
	fmt.Println("ReplaceAllFunc:   ", reCPF.ReplaceAllStringFunc("CPF 123.456.789-09 e 987.654.321-00", func(s string) string {
		return "***.***.***-" + s[len(s)-2:]
	}))
	fmt.Println("Email achados:    ", reEmail.FindAllString("fale com Ana@Exemplo.com.br ou bob+x@ex.org.", -1))

	// o ponto NAO casa quebra de linha, a menos que (?s); (?m) muda ^ e $
	fmt.Println("ponto e \\n:       ", regexp.MustCompile(`a.b`).MatchString("a\nb"), regexp.MustCompile(`(?s)a.b`).MatchString("a\nb"))
	fmt.Println("multilinha:       ", regexp.MustCompile(`(?m)^\w+$`).FindAllString("um\ndois\ntres", -1))

	// guloso x preguicoso
	html := `<b>um</b> e <b>dois</b>`
	fmt.Println("guloso  <b>.*</b>:  ", regexp.MustCompile(`<b>.*</b>`).FindString(html))
	fmt.Println("preguicoso .*?:     ", regexp.MustCompile(`<b>.*?</b>`).FindString(html))

	// escapar texto do usuario antes de embutir em um padrao
	busca := "preco (R$) [1+1]"
	fmt.Println("QuoteMeta:        ", regexp.QuoteMeta(busca))
	fmt.Println("literal casa:     ", regexp.MustCompile(regexp.QuoteMeta(busca)).MatchString("tabela de "+busca))

	// o que o RE2 NAO tem
	_, err := regexp.Compile(`(\w+) \1`)
	fmt.Println("retrovisor \\1:    ", err)
	_, err = regexp.Compile(`foo(?=bar)`)
	fmt.Println("lookahead:        ", err)
	_, err = regexp.Compile(`(abc`)
	fmt.Println("padrao invalido:  ", err)

	// Split com regex
	fmt.Printf("Split:            %q\n", regexp.MustCompile(`\s*[,;]\s*`).Split("a , b;c ,d", -1))
}
