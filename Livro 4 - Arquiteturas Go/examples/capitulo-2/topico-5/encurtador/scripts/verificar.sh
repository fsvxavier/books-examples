#!/usr/bin/env bash
# Verificação que a integração contínua e o desenvolvedor rodam do mesmo jeito.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "== gofmt"
fora=$(gofmt -l .)
if [ -n "$fora" ]; then
	echo "arquivos fora do padrão:"
	echo "$fora"
	exit 1
fi

echo "== go vet"
go vet ./...

echo "== go test -race"
go test -race -count=1 ./...
