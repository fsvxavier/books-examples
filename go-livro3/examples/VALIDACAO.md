# Validação dos exemplos: Livro 3 - GO e Engenharia de IA

Cada diretório com código foi verificado com `go vet`; os programas (`func main`) foram executados com `go run` (limite de 25 s, sem entrada, com os argumentos que o capítulo usa quando precisam de algum) e os testes com `go test`. Go 1.27.1, bancos de dados reais em Docker (PostgreSQL, MySQL, MongoDB, Valkey).

Verificações: 67.

| Resultado | Quantidade | Significado |
|---|---|---|
| passou | 51 | terminou sem erro |
| passou, com falha documentada | 3 | o capítulo ensina um defeito de propósito (erro de compilação, `panic`, deadlock, teste que falha) e a mensagem real aparece no texto |
| fora do módulo | 0 | as ilustrações de outros módulos viraram pacotes de verdade dentro do módulo único de `examples` (ver abaixo) |
| falhou | 0 | |

## Falhas documentadas

| Diretório | Etapa | Primeira linha da saída |
|---|---|---|
| examples/capitulo-1/topico-6 | run | `0-42ba-8b1c-b14c02b42b93/scratchpad/rc3/3/capitulo-1/topico-6/arquivo-1.go:17 +0xb8` |
| examples/capitulo-11/topico-8/arquivo-1 | run | `c02b42b93/scratchpad/rc3/3/capitulo-11/topico-8/arquivo-1/arquivo-1.go:23 +0x10b` |
| examples/capitulo-9/topico-15 | run | `uso: go run . CAMINHO/libonnxruntime.so soma.onnx` |

## Projetos em subpastas

Antes eram módulos com `go.mod` próprio. Agora existe um único `go.mod`, na raiz de `examples`, e cada projeto é um conjunto de pacotes dentro dele; a verificação roda a partir da raiz.

- `examples/capitulo-1/topico-11/projeto-1`
- `examples/capitulo-11/topico-8/projeto-1`
- `examples/capitulo-12/topico-10/projeto-1`
- `examples/capitulo-5/topico-9/projeto-1`
- `examples/capitulo-7/topico-7/projeto-1`
