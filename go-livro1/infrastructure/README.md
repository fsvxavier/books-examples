# Infraestrutura do Livro 1

Bancos reais do capítulo 20 (Database). Os exemplos criam as próprias tabelas e coleções.

## Serviços

| Serviço | Imagem | Porta (127.0.0.1) | Para quê |
|---|---|---|---|
| PostgreSQL 18 | `postgres:18-alpine` | `55432` | pgx, pool, transações, migrações |
| MySQL 9 | `mysql:9` | `53306` | `database/sql` com `go-sql-driver/mysql` |
| MongoDB 8 | `mongo:8` | `57017` | `mongo-driver/v2` |
| Valkey | `valkey/valkey:latest` | `56390` | `valkey-go` e `go-redis` (cache, trava, Pub/Sub) |

Todas as portas ficam presas a `127.0.0.1` e usam números altos para não colidir com instalações locais. Usuário, senha e banco são `livro` quando o serviço pede credenciais.

## Uso

```bash
cd go-livro1/examples/infrastructure
docker compose up -d --wait     # sobe e espera os healthchecks
docker compose ps
docker compose down             # para e remove os contêineres
```

Para apagar também os dados: `docker compose down -v`.

## Conexão

```text
postgres://livro:livro@127.0.0.1:55432/livro?sslmode=disable
livro:livro@tcp(127.0.0.1:53306)/livro?parseTime=true
mongodb://127.0.0.1:57017
127.0.0.1:56390
```


Os livros repetem os mesmos nomes de contêiner (`golivro-pg`, `golivro-valkey`...) e portas. Suba a infraestrutura de **um livro por vez**, ou rode `docker compose down` no anterior antes de subir outro.
