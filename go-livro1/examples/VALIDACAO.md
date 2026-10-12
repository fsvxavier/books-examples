# Validação dos exemplos: Livro 1 - Go Basico e Completo

Cada diretório com código foi verificado com `go vet`; os programas (`func main`) foram executados com `go run` (limite de 25 s, sem entrada, com os argumentos que o capítulo usa quando precisam de algum) e os testes com `go test`. Go 1.27.1, bancos de dados reais em Docker (PostgreSQL, MySQL, MongoDB, Valkey).

Verificações: 259.

| Resultado | Quantidade | Significado |
|---|---|---|
| passou | 244 | terminou sem erro |
| passou, com falha documentada | 7 | o capítulo ensina um defeito de propósito (erro de compilação, `panic`, deadlock, teste que falha) e a mensagem real aparece no texto |
| fora do módulo | 0 | as ilustrações de outros módulos viraram pacotes de verdade dentro do módulo único de `examples` (ver abaixo) |
| falhou | 0 | |

## Falhas documentadas

| Diretório | Etapa | Primeira linha da saída |
|---|---|---|
| examples/capitulo-1/topico-11 | vet | `arquivo-1.go:7:21: fmt.Printf format %s has arg idade of wrong type int` |
| examples/capitulo-16/topico-3 | test | `--- FAIL: TestErrorfContinua (0.00s)` |
| examples/capitulo-17/topico-6 | vet | `quivo-1.go:22:28: fmt.Printf format %s has arg idade of wrong type int` |
| examples/capitulo-17/topico-8 | vet | `vet: ./arquivo-2.go:5:26: undefined: strings` |
| examples/capitulo-21/topico-2 | vet | `arquivo-1.go:62:2: call to slog.Logger.Info missing a final value` |
| examples/capitulo-23/topico-6 | vet | `arquivo-1.go:29:56: 2006-02-01 should be 2006-01-02` |
| examples/capitulo-4/topico-11 | run | `1. variável de pacote inicializada` |

## Projetos em subpastas

Antes eram módulos com `go.mod` próprio. Agora existe um único `go.mod`, na raiz de `examples`, e cada projeto é um conjunto de pacotes dentro dele; a verificação roda a partir da raiz.

- `examples/capitulo-1/topico-7/projeto-1`
- `examples/capitulo-15/topico-10/projeto-1`
- `examples/capitulo-15/topico-14/projeto-1`
- `examples/capitulo-15/topico-8/projeto-1`
- `examples/capitulo-15/topico-8/projeto-2`
- `examples/capitulo-15/topico-8/projeto-3`
- `examples/capitulo-17/topico-8/projeto-1`
