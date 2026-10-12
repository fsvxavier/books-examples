# Infraestrutura do Livro 4

Serviços do encurtador de links (arquitetura limpa, hexagonal, DDD, vertical slice e eventos).

## Serviços

| Serviço | Imagem | Porta (127.0.0.1) | Para quê |
|---|---|---|---|
| PostgreSQL 18 | `postgres:18-alpine` | `55432` | repositório de links e testes de integração |
| Valkey | `valkey/valkey:latest` | `56390` | Streams (protocolo do Redis) usados pelo `watermill-redisstream` |

Todas as portas ficam presas a `127.0.0.1` e usam números altos para não colidir com instalações locais. Usuário, senha e banco são `livro` quando o serviço pede credenciais.

## Uso

```bash
cd go-livro4/examples/infrastructure
docker compose up -d --wait     # sobe e espera os healthchecks
docker compose ps
docker compose down             # para e remove os contêineres
```

Para apagar também os dados: `docker compose down -v`.

## Conexão

```bash
export ENCURTADOR_DSN=postgres://livro:livro@127.0.0.1:55432/livro?sslmode=disable
export ENCURTADOR_TESTE_DSN=$ENCURTADOR_DSN     # sem ela, os testes de PostgreSQL são pulados
```

Os testes criam um esquema isolado por teste, então podem rodar contra o mesmo banco do servidor.


Os livros repetem os mesmos nomes de contêiner (`golivro-pg`, `golivro-valkey`...) e portas. Suba a infraestrutura de **um livro por vez**, ou rode `docker compose down` no anterior antes de subir outro.
