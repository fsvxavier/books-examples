# Infraestrutura do Livro 3

Serviços dos capítulos de RAG, feature stores, gateway de IA e memória de agentes.

## Serviços

| Serviço | Imagem | Porta (127.0.0.1) | Para quê |
|---|---|---|---|
| PostgreSQL + pgvector | `pgvector/pgvector:pg18` | `55433` | busca vetorial (RAG e memória longa); a extensão `vector` é criada em `initdb/` |
| Valkey | `valkey/valkey:latest` | `56390` | limitador de taxa, cache e feature store online |
| Valkey bundle | `valkey/valkey-bundle:latest` | `56391` | busca vetorial com o módulo `search` |
| Ollama (opcional) | `ollama/ollama:latest` | `11434` | modelos locais; perfil `ollama` |

Todas as portas ficam presas a `127.0.0.1` e usam números altos para não colidir com instalações locais. Usuário, senha e banco são `livro` quando o serviço pede credenciais.

## Uso

```bash
cd go-livro3/examples/infrastructure
docker compose up -d --wait     # sobe e espera os healthchecks
docker compose ps
docker compose down             # para e remove os contêineres
```

Para apagar também os dados: `docker compose down -v`.

## Conexão

```bash
export PG=postgres://livro:livro@127.0.0.1:55433/livro
export VALKEY=127.0.0.1:56390 VALKEY_ADDR=127.0.0.1:56390
export VALKEY_SEARCH_ADDR=127.0.0.1:56391
```

### Ollama

Fica fora da subida padrão porque baixa modelos grandes. Se já tem o Ollama no host, não precisa dele.

```bash
docker compose --profile ollama up -d
docker exec golivro-ollama ollama pull gemma3:1b
docker exec golivro-ollama ollama pull embeddinggemma:300m
```

O Ollama aqui roda só em CPU. Para usar GPU no contêiner, é preciso adicionar `deploy.resources.reservations.devices` ao serviço, conforme o seu driver.


Os livros repetem os mesmos nomes de contêiner (`golivro-pg`, `golivro-valkey`...) e portas. Suba a infraestrutura de **um livro por vez**, ou rode `docker compose down` no anterior antes de subir outro.
