# Infraestrutura do Livro 2

O Livro 2 **não precisa de banco, fila ou outro serviço**: todos os exemplos rodam só com a toolchain Go.

O único uso de Docker é o capítulo 20 (cgo): o serviço `musl` roda binários Linux numa imagem Alpine (libc musl) para mostrar a diferença entre glibc e musl. A imagem é `alpine:3`; o livro usa `postgres:18-alpine` só como base Alpine, e o resultado é o mesmo.

## Uso

```bash
cd go-livro2/examples/capitulo-20/topico-XX      # onde está o main.go do exemplo
go build -o app-cgo . && CGO_ENABLED=0 go build -o app-pure .
cd -
cd go-livro2/examples/infrastructure
BINARIOS=/caminho/dos/binarios docker compose --profile cgo run --rm musl /app/app-cgo
```

`BINARIOS` é a pasta com os binários (padrão: a pasta atual). O binário dinâmico (`app-cgo`) falha no Alpine por falta da glibc; o `app-pure` e o estático rodam.
