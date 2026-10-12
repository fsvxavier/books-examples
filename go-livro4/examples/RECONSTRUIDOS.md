# Arquivos reconstruídos do Livro 4

Os capítulos mostram apenas parte dos projetos do encurtador (os demais arquivos existiam só no ambiente de escrita). Os arquivos abaixo foram **reconstruídos** a partir do que o texto de cada capítulo descreve, do projeto do capítulo anterior (que é a base de cada novo estilo de arquitetura) e do que os arquivos mostrados exigem para compilar. Eles não aparecem como bloco de código no livro.

Foram verificados com `go vet` e `go test -race`. Os testes que dependem de um PostgreSQL (`ENCURTADOR_TESTE_DSN`) são pulados sem o banco; não havia Docker nesta máquina para exercitá-los.

## capitulo-2/topico-3

- `capitulo-2/topico-3/lib/internal/segredo/segredo.go`
- `capitulo-2/topico-3/lib/palavra/palavra.go`
- `capitulo-2/topico-3/lib/pkg/numero/numero.go`

## capitulo-2/topico-5

- `capitulo-2/topico-5/encurtador/internal/armazem/armazem.go`
- `capitulo-2/topico-5/encurtador/internal/config/config.go`
- `capitulo-2/topico-5/encurtador/internal/servidor/codigo.go`
- `capitulo-2/topico-5/encurtador/internal/servidor/servidor.go`
- `capitulo-2/topico-5/encurtador/internal/servidor/validacao.go`

## capitulo-3/topico-2

- `capitulo-3/topico-2/encurtador/cmd/encurtador/main.go`
- `capitulo-3/topico-2/encurtador/internal/config/config.go`
- `capitulo-3/topico-2/encurtador/internal/contrato/contrato.go`
- `capitulo-3/topico-2/encurtador/internal/versao/versao.go`

## capitulo-4/topico-2

- `capitulo-4/topico-2/encurtador/cmd/encurtador/main.go`
- `capitulo-4/topico-2/encurtador/internal/adaptadores/entrada/web/web.go`
- `capitulo-4/topico-2/encurtador/internal/adaptadores/entrada/web/web_test.go`
- `capitulo-4/topico-2/encurtador/internal/adaptadores/saida/memoria/memoria.go`
- `capitulo-4/topico-2/encurtador/internal/adaptadores/saida/memoria/memoria_test.go`
- `capitulo-4/topico-2/encurtador/internal/adaptadores/saida/postgres/postgres.go`
- `capitulo-4/topico-2/encurtador/internal/adaptadores/saida/postgres/postgres_test.go`
- `capitulo-4/topico-2/encurtador/internal/config/config.go`
- `capitulo-4/topico-2/encurtador/internal/contrato/contrato.go`
- `capitulo-4/topico-2/encurtador/internal/nucleo/codigo.go`
- `capitulo-4/topico-2/encurtador/internal/nucleo/dominio/link.go`
- `capitulo-4/topico-2/encurtador/internal/nucleo/dominio/link_test.go`
- `capitulo-4/topico-2/encurtador/internal/nucleo/repositoriotest/repositoriotest.go`
- `capitulo-4/topico-2/encurtador/internal/versao/versao.go`

## capitulo-5/topico-3

- `capitulo-5/topico-3/encurtador/internal/nucleo/codigo.go`
- `capitulo-5/topico-3/encurtador/internal/nucleo/dominio/erros.go`
- `capitulo-5/topico-3/encurtador/internal/nucleo/porta_entrada.go`

## capitulo-6/topico-4

- `capitulo-6/topico-4/encurtador/cmd/encurtador/main.go`
- `capitulo-6/topico-4/encurtador/internal/app/app.go`
- `capitulo-6/topico-4/encurtador/internal/auditoria/auditoria.go`
- `capitulo-6/topico-4/encurtador/internal/config/config.go`
- `capitulo-6/topico-4/encurtador/internal/contrato/contrato.go`
- `capitulo-6/topico-4/encurtador/internal/criar/codigo.go`
- `capitulo-6/topico-4/encurtador/internal/dominio/codigo.go`
- `capitulo-6/topico-4/encurtador/internal/dominio/destino.go`
- `capitulo-6/topico-4/encurtador/internal/dominio/erros.go`
- `capitulo-6/topico-4/encurtador/internal/dominio/evento.go`
- `capitulo-6/topico-4/encurtador/internal/dominio/link.go`
- `capitulo-6/topico-4/encurtador/internal/dominio/link_test.go`
- `capitulo-6/topico-4/encurtador/internal/infra/banco/banco.go`
- `capitulo-6/topico-4/encurtador/internal/infra/barramento/barramento.go`
- `capitulo-6/topico-4/encurtador/internal/infra/barramento/barramento_test.go`
- `capitulo-6/topico-4/encurtador/internal/infra/resposta/resposta.go`
- `capitulo-6/topico-4/encurtador/internal/versao/versao.go`

## capitulo-7/topico-2

- `capitulo-7/topico-2/encurtador/internal/dominio/codigo.go`
- `capitulo-7/topico-2/encurtador/internal/dominio/destino.go`
- `capitulo-7/topico-2/encurtador/internal/dominio/erros.go`
- `capitulo-7/topico-2/encurtador/internal/dominio/evento.go`
- `capitulo-7/topico-2/encurtador/internal/dominio/link.go`
- `capitulo-7/topico-2/encurtador/internal/dominio/link_test.go`
