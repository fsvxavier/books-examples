# Arquivos adicionados e movidos: Livro 1 - Go Basico e Completo

Os capítulos mostram parte do código em pedaços, ou mostram um arquivo em um tópico e o usam em outro. Para que os exemplos compilem e rodem, foram **adicionados** os arquivos abaixo (cópias de tópicos irmãos, trechos do capítulo reunidos, arquivos de apoio como `main.go`, SQL e modelos, ou código escrito a partir do que o texto descreve). Não aparecem como bloco de código no livro. Arquivos movidos (por serem versões alternativas ou por pertencerem a outro módulo) aparecem na segunda lista.

## Adicionados

- `capitulo-1/topico-11/projeto-1/main.go`
- `capitulo-11/topico-11/arquivo-1/arquivo-4_test.go`
- `capitulo-12/topico-13/arquivo-2/arquivo-5_test.go`
- `capitulo-12/topico-13/main.go`
- `capitulo-14/topico-16/arquivo-2/arquivo-2.go`
- `capitulo-14/topico-16/arquivo-2/arquivo-3_test.go`
- `capitulo-15/topico-10/projeto-1/arq.go`
- `capitulo-15/topico-10/projeto-1/arq_externo_test.go`
- `capitulo-15/topico-10/projeto-1/assets/email.tmpl`
- `capitulo-15/topico-10/projeto-1/assets/sql/001_pedidos.sql`
- `capitulo-15/topico-10/projeto-1/assets/sql/002_valor.sql`
- `capitulo-15/topico-10/projeto-1/embutidos.go`
- `capitulo-15/topico-10/projeto-1/go.mod`
- `capitulo-15/topico-10/projeto-1/nome_darwin.go`
- `capitulo-15/topico-10/projeto-1/nome_linux.go`
- `capitulo-15/topico-10/projeto-1/nome_outro.go`
- `capitulo-15/topico-10/projeto-1/nome_windows.go`
- `capitulo-15/topico-10/projeto-1/testdata/caso1.json`
- `capitulo-15/topico-8/arquivo-6/arquivo-6.go`
- `capitulo-15/topico-8/arquivo-6/ref.go`
- `capitulo-17/topico-11/arquivo-2/estado_string.go`
- `capitulo-17/topico-11/arquivo-2/gerador/main.go`
- `capitulo-17/topico-11/arquivo-2/versao_gerada.go`
- `capitulo-17/topico-5/projeto-1/main.go`
- `capitulo-20/topico-12/migracoes/0001_esquema.sql`
- `capitulo-7/topico-16/arquivo-3/arquivo-4_test.go`

## Que deixaram o caminho original (movidos)

- `capitulo-11/topico-11/arquivo-4_test.go`
- `capitulo-12/topico-13/arquivo-5_test.go`
- `capitulo-14/topico-16/arquivo-1/arquivo-2.go`
- `capitulo-14/topico-16/arquivo-1/arquivo-3_test.go`
- `capitulo-15/topico-10/arquivo-1.go`
- `capitulo-15/topico-10/arquivo-2.go`
- `capitulo-15/topico-10/arquivo-3.go`
- `capitulo-15/topico-10/arquivo-4.go`
- `capitulo-15/topico-11/arquivo-1/arquivo-1_test.go`
- `capitulo-15/topico-8/arquivo-3/arquivo-6.go`
- `capitulo-7/topico-16/arquivo-4_test.go`

## Módulos de ilustração montados

As ilustrações de projetos fictícios (`exemplo.com/...`) viraram módulos aninhados, cada um com o seu `go.mod`, e os arquivos foram movidos para a estrutura de pastas que o capítulo descreve (as legendas apontam para o novo caminho). Foi escrito, a partir do que o texto descreve, o que não aparece em bloco de código:

- capitulo-15/topico-8/projeto-2/models/pedido.go
- capitulo-17/topico-8/projeto-1/interno/interno.go
- go.mod de cada projeto-N (cpf, calc, ciclo, camadas, dominio, fm)
