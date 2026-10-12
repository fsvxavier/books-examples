VERSAO ?= dev
PKG    := exemplos/capitulo-40/topico-5/encurtador
LDFLAGS := -X $(PKG)/internal/versao.Versao=$(VERSAO)

.PHONY: build test verificar limpar

build:
	go build -ldflags "$(LDFLAGS)" -o bin/ ./cmd/...

test:
	go test -race ./...

verificar:
	./scripts/verificar.sh

limpar:
	rm -rf bin
