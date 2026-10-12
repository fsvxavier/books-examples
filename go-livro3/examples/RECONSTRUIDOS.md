# Arquivos adicionados e movidos: Livro 3 - GO e Engenharia de IA

Os capítulos mostram parte do código em pedaços, ou mostram um arquivo em um tópico e o usam em outro. Para que os exemplos compilem e rodem, foram **adicionados** os arquivos abaixo (cópias de tópicos irmãos, trechos do capítulo reunidos, arquivos de apoio como `main.go`, SQL e modelos, ou código escrito a partir do que o texto descreve). Não aparecem como bloco de código no livro. Arquivos movidos (por serem versões alternativas ou por pertencerem a outro módulo) aparecem na segunda lista.

## Adicionados

- `capitulo-4/topico-6/main.go`
- `capitulo-8/topico-3/projeto-1/main.go`

## Módulos de ilustração montados

As ilustrações de projetos fictícios (`exemplo.com/...`) viraram subpastas do módulo único de `examples` (há um só `go.mod`, na raiz), e os arquivos foram movidos para a estrutura de pastas que o capítulo descreve (as legendas apontam para o novo caminho). Foi escrito, a partir do que o texto descreve, o que não aparece em bloco de código:

- capitulo-12/topico-5 (agora capitulo-12/topico-10/projeto-1)/internal/longo/embedder.go e extrator.go (Embedder, Hash, Fato, Extrair)
- capitulo-7/topico-7/projeto-1/internal/janela/janela.go (tipos e construtores das janelas ingênua, exata e de baldes, com os métodos do capítulo)
- capitulo-7/topico-7/projeto-1/cmd/lab/main.go (dados sintéticos, despachador do subcomando exportar)
- capitulo-5/topico-9/projeto-1/internal/achado/achado.go, internal/portao/portao.go, internal/analisadores/errodescartado/ (Analyzer, isenções e teste), testdata e semctx_test.go
- capitulo-11/topico-8/projeto-1/internal/equipe/equipe.go (reúne os trechos do capítulo, mais Opcoes, BaseDemo, ErrReprovado e pesquisarLocal)
- capitulo-4/topico-6/main.go (despachador de subcomandos, cortaTxt, encurta)
