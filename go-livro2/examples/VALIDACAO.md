# Validação dos exemplos: Livro 2 - Go Avançado e Design Patterns

Cada diretório com código foi verificado com `go vet`; os programas (`func main`) foram executados com `go run` (limite de 25 s, sem entrada, com os argumentos que o capítulo usa quando precisam de algum) e os testes com `go test`. Go 1.27.1, bancos de dados reais em Docker (PostgreSQL, MySQL, MongoDB, Valkey).

Verificações: 593.

| Resultado | Quantidade | Significado |
|---|---|---|
| passou | 558 | terminou sem erro |
| passou, com falha documentada | 29 | o capítulo ensina um defeito de propósito (erro de compilação, `panic`, deadlock, teste que falha) e a mensagem real aparece no texto |
| fora do módulo | 0 | as ilustrações de outros módulos viraram pacotes de verdade dentro do módulo único de `examples` (ver abaixo) |
| falhou | 0 | |

## Falhas documentadas

| Diretório | Etapa | Primeira linha da saída |
|---|---|---|
| examples/capitulo-1/topico-4/projeto-2 | test | `o (0.00s)` |
| examples/capitulo-11/topico-16/arquivo-2 | vet | `arquivo-2.go:10:2: misuse of unbuffered os.Signal channel as argument to signal.Notify` |
| examples/capitulo-11/topico-3/projeto-1/e01x | vet | `vet: ./main.go:12:4: m.linhas undefined (cannot refer to unexported field linhas)` |
| examples/capitulo-11/topico-3/projeto-2/e14x | vet | `vet: ./main.go:10:23: cannot use falso{} (value of struct type falso) as doc.Snapshot value in variable declaration: falso does not implement doc.Snap` |
| examples/capitulo-13/topico-10/projeto-1 | vet | `mplos/livro2/capitulo-13/topico-10/projeto-1` |
| examples/capitulo-13/topico-10/projeto-2 | test | `--- FAIL: TestDescreverTrataTodosOsTipos (0.00s)` |
| examples/capitulo-13/topico-12/projeto-1 | vet | `vet: ./main.go:16:9: interface method must have no type parameters` |
| examples/capitulo-13/topico-4/projeto-1 | vet | `github.com/fsvxavier/books-examples/go-livro-2/capitulo-13/topico-4/projeto-1` |
| examples/capitulo-14/topico-3 | run | `+0x28` |
| examples/capitulo-15/topico-12 | vet | `48:21: call of copiar copies lock value: github.com/fsvxavier/books-examples/go-livro-2/capitulo-15/topico-12.Contador contains sync.Mutex` |
| examples/capitulo-15/topico-13/projeto-2 | test | `--- FAIL: TestNaoVazaGoroutines (0.05s)` |
| examples/capitulo-15/topico-2/projeto-1 | test | `--- FAIL: TestGlobalMoedaPadrao (0.00s)` |
| examples/capitulo-15/topico-3/projeto-1 | test | `github.com/fsvxavier/books-examples/go-livro-2/capitulo-15/topico-3/projeto-1.Consultar in goroutine 9` |
| examples/capitulo-17/topico-12/erros/b1 | vet | `vet: ./main.go:13:24: Produto does not satisfy ComIdentificador[string] (missing method ObterID)` |
| examples/capitulo-17/topico-12/erros/p1 | vet | `vet: ./main.go:17:32: Conexao does not satisfy Iniciavel (method Iniciar has pointer receiver)` |
| examples/capitulo-17/topico-12/erros/r1 | vet | `vet: ./main.go:22:27: in call to Maximo, T (type Item) does not satisfy Comparavel[T] (wrong type for method Comparar)` |
| examples/capitulo-17/topico-12/erros/t1 | vet | `vet: ./main.go:15:27: in call to Rotulo, T (type Pontos) does not satisfy Codigo (missing method String)` |
| examples/capitulo-17/topico-12/erros/t2 | vet | `vet: ./main.go:13:27: cannot satisfy Impossivel (empty type set)` |
| examples/capitulo-19/topico-4 | vet | `arquivo-1.go:24:14: possible misuse of unsafe.Pointer` |
| examples/capitulo-19/topico-6 | run | `/runtime/mfinal.go:210 +0x107 fp=0x30d014ec1fe0 sp=0x30d014ec1e20 pc=0x424d87` |
| examples/capitulo-20/topico-12/arquivo-2 | run | `a85dd0` |
| examples/capitulo-20/topico-6/arquivo-2 | vet | `arquivo-2.go:18:12: possibly passing Go type with embedded pointer to C` |
| examples/capitulo-4/topico-12/projeto-1/cmd/duplicado | run | `1c-b14c02b42b93/scratchpad/rc3/2/capitulo-4/topico-12/projeto-1/pagamento/pagamento.go:34 +0x125` |
| examples/capitulo-4/topico-4/arquivo-2 | run | `/8be21dc1-0140-42ba-8b1c-b14c02b42b93/scratchpad/rc3/2/capitulo-4/topico-4/arquivo-2/arquivo-2.go:11 +0x2c` |
| examples/capitulo-4/topico-4/arquivo-3 | vet | `arquivo-3.go:25:7: assignment copies lock value to b: github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-4/arquivo-3.Cliente contains sync.Once contains sync.noCopy` |
| examples/capitulo-4/topico-5/projeto-1/cmd/burla | vet | `vet: ./main.go:10:13: undefined: banco.pool` |
| examples/capitulo-4/topico-8/projeto-1 | test | `--- FAIL: TestAmbienteTeste (0.00s)` |
| examples/capitulo-4/topico-8/projeto-2 | test | `--- FAIL: TestParalelosBrigam (0.00s)` |
| examples/capitulo-6/topico-5 | run | `/src/runtime/mfinal.go:210 +0x107 fp=0x676095e07e0 sp=0x676095e0620 pc=0x424d87` |

## Projetos em subpastas

Antes eram módulos com `go.mod` próprio. Agora existe um único `go.mod`, na raiz de `examples`, e cada projeto é um conjunto de pacotes dentro dele; a verificação roda a partir da raiz.

- `examples/capitulo-17/topico-12/erros`
