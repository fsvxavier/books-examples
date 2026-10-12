package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func init() { registra("cliente", cliente) }

// cliente sobe o servidor como subprocesso (stdio) e exercita as ferramentas.
func cliente(args []string) error {
	fs := flag.NewFlagSet("cliente", flag.ExitOnError)
	bin := fs.String("servidor", "./mcpserver", "binário do servidor MCP")
	base := fs.String("base", "dados/livro.db", "base de conhecimento")
	fs.Parse(args)
	ctx, cancela := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancela()

	cmd := exec.Command(*bin, "-base", *base)
	cmd.Stderr = nil // o log estruturado do servidor sairia aqui; descartamos
	cli := mcp.NewClient(&mcp.Implementation{Name: "lab-cliente", Version: "v1.0.0"}, nil)
	cs, err := cli.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		return err
	}
	defer cs.Close()
	ini := cs.InitializeResult()
	fmt.Printf("servidor: %s %s, protocolo %s\n", ini.ServerInfo.Name, ini.ServerInfo.Version, ini.ProtocolVersion)

	fmt.Println("\nferramentas:")
	for t, err := range cs.Tools(ctx, nil) {
		if err != nil {
			return err
		}
		a := t.Annotations
		var marcas []string
		if a != nil && a.ReadOnlyHint {
			marcas = append(marcas, "só leitura")
		}
		if a != nil && a.IdempotentHint {
			marcas = append(marcas, "idempotente")
		}
		if a != nil && a.DestructiveHint != nil && *a.DestructiveHint {
			marcas = append(marcas, "destrutiva")
		}
		fmt.Printf("  %-18s %s\n", t.Name, strings.Join(marcas, ", "))
	}

	chama := func(titulo, nome string, args map[string]any) {
		ini := time.Now()
		r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: nome, Arguments: args})
		d := time.Since(ini).Round(time.Millisecond)
		fmt.Printf("\n%s (%v)\n", titulo, d)
		if err != nil {
			fmt.Println("  erro de protocolo:", err)
			return
		}
		prefixo := "  resultado"
		if r.IsError {
			prefixo = "  erro da ferramenta"
		}
		var txt string
		for _, c := range r.Content {
			if t, ok := c.(*mcp.TextContent); ok {
				txt += t.Text
			}
		}
		fmt.Printf("%s: %s\n", prefixo, resumir(txt))
	}
	fmt.Println("\nmodelos de recurso:")
	for t, err := range cs.ResourceTemplates(ctx, nil) {
		if err != nil {
			return err
		}
		fmt.Printf("  %s  (%s)\n", t.URITemplate, t.MIMEType)
	}
	fmt.Println("\nprompts:")
	for p, err := range cs.Prompts(ctx, nil) {
		if err != nil {
			return err
		}
		var args []string
		for _, a := range p.Arguments {
			if a.Required {
				args = append(args, a.Name+"*")
			} else {
				args = append(args, a.Name)
			}
		}
		fmt.Printf("  %s(%s)\n", p.Name, strings.Join(args, ", "))
	}
	chama("1. buscar no livro", "buscar_trechos", map[string]any{"consulta": "Como limitar quantas goroutines o errgroup executa ao mesmo tempo?", "quantidade": 2})
	chama("2. abrir chamado", "abrir_chamado", map[string]any{"titulo": "Erro no login", "descricao": "não entra", "chave_idempotencia": "pedido-0001"})
	chama("3. repetir o mesmo pedido", "abrir_chamado", map[string]any{"titulo": "Erro no login", "descricao": "não entra", "chave_idempotencia": "pedido-0001"})
	chama("4. mesma chave, outros dados", "abrir_chamado", map[string]any{"titulo": "Outro assunto", "chave_idempotencia": "pedido-0001"})
	chama("5. sem a chave obrigatória", "abrir_chamado", map[string]any{"titulo": "Sem chave"})
	chama("6. cancelar um chamado que não existe", "cancelar_chamado", map[string]any{"id": 99, "motivo": "teste"})
	chama("7. ferramenta inexistente", "formatar_disco", map[string]any{})
	rr, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "livro://documento/capitulo-12-context.md"})
	if err != nil {
		return err
	}
	fmt.Printf("\n8. ler o recurso livro://documento/capitulo-12-context.md\n  %d bytes de %s; começa com: %q\n", len(rr.Contents[0].Text), rr.Contents[0].MIMEType, cortaTxt(rr.Contents[0].Text, 70))
	_, err = cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "livro://documento/../../etc/passwd"})
	fmt.Printf("\n9. recurso fora do formato\n  erro: %v\n", err)
	pr, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "explicar_conceito", Arguments: map[string]string{"conceito": "context.WithTimeout"}})
	if err != nil {
		return err
	}
	fmt.Printf("\n10. pedir o prompt explicar_conceito\n  %s: %s\n", pr.Messages[0].Role, pr.Messages[0].Content.(*mcp.TextContent).Text)
	return nil
}

func resumir(s string) string {
	var v any
	if json.Unmarshal([]byte(s), &v) == nil {
		b, _ := json.Marshal(encurta(v))
		s = string(b)
	}
	if r := []rune(s); len(r) > 330 {
		s = string(r[:330]) + "..."
	}
	return s
}
