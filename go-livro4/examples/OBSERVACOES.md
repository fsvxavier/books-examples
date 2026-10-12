## Observações

- Os projetos do encurtador dos capítulos 2 a 7 foram completados com arquivos reconstruídos: ver `RECONSTRUIDOS.md`. No capítulo 7 só existem o domínio e os pacotes mostrados no texto (`mensagens`, `outbox`, `mensageria`, `contagem`, `expiracao`); faltam `criar`, `seguir`, `app` e os testes, que o capítulo só mostra em trechos. Nos capítulos 4 a 6 faltam os testes das pastas `test/` e, no 5, os adaptadores de entrada e a composição.
- Os testes que usam PostgreSQL rodam com `ENCURTADOR_TESTE_DSN=postgres://livro:livro@127.0.0.1:55432/livro` (contêiner `golivro-pg`).
- `capitulo-2/topico-3/consumidor` falha de propósito: o capítulo mostra que o compilador recusa o pacote `internal/segredo` de outro módulo.
