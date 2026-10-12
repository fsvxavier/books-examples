# Validação dos exemplos: Livro 4 - Arquiteturas Go

Cada diretório com código foi verificado com `go vet`; os programas (`func main`) foram executados com `go run` (limite de 25 s, sem entrada, com os argumentos que o capítulo usa quando precisam de algum) e os testes com `go test`. Go 1.27.1, bancos de dados reais em Docker (PostgreSQL, MySQL, MongoDB, Valkey).

Verificações: 103.

| Resultado | Quantidade | Significado |
|---|---|---|
| passou | 102 | terminou sem erro |
| passou, com falha documentada | 1 | o capítulo ensina um defeito de propósito (erro de compilação, `panic`, deadlock, teste que falha) e a mensagem real aparece no texto |
| fora do módulo | 0 | as ilustrações de outros módulos viraram módulos de verdade (ver abaixo) |
| falhou | 0 | |

## Falhas documentadas

| Diretório | Etapa | Primeira linha da saída |
|---|---|---|
| examples/capitulo-2/topico-3/consumidor | vet | `package exemplos/livro4/capitulo-2/topico-3/consumidor` |


## Observações

- Os projetos do encurtador dos capítulos 2 a 7 foram completados com arquivos reconstruídos: ver `RECONSTRUIDOS.md`. No capítulo 7 só existem o domínio e os pacotes mostrados no texto (`mensagens`, `outbox`, `mensageria`, `contagem`, `expiracao`); faltam `criar`, `seguir`, `app` e os testes, que o capítulo só mostra em trechos. Nos capítulos 4 a 6 faltam os testes das pastas `test/` e, no 5, os adaptadores de entrada e a composição.
- Os testes que usam PostgreSQL rodam com `ENCURTADOR_TESTE_DSN=postgres://livro:livro@127.0.0.1:55432/livro` (contêiner `golivro-pg`).
- `capitulo-2/topico-3/consumidor` falha de propósito: o capítulo mostra que o compilador recusa o pacote `internal/segredo` de outro módulo.
