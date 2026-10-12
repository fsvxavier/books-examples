// Package curto guarda o histórico de conversas no Valkey, uma lista por conversa.
package curto

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-12/topico-10/projeto-1/internal/janela"
)

// anexar faz, de forma atômica no servidor: RPUSH, LTRIM e EXPIRE.
// Se três comandos separados fossem enviados, outro cliente poderia se
// intrometer entre eles, e uma queda no meio deixaria a lista sem TTL.
var anexar = valkey.NewLuaScript(`
redis.call('RPUSH', KEYS[1], ARGV[1])
redis.call('LTRIM', KEYS[1], -tonumber(ARGV[2]), -1)
redis.call('EXPIRE', KEYS[1], tonumber(ARGV[3]))
return redis.call('LLEN', KEYS[1])
`)

// Historico guarda as últimas Max mensagens de cada conversa, por TTL deslizante.
type Historico struct {
	Cliente valkey.Client
	Max     int           // mensagens guardadas por conversa
	TTL     time.Duration // some depois desse tempo sem novas mensagens
}

func chave(conversa string) string { return "conv:" + conversa }

// Anexar acrescenta uma mensagem e devolve quantas a conversa tem agora.
func (h *Historico) Anexar(ctx context.Context, conversa string, m janela.Mensagem) (int64, error) {
	dados, err := json.Marshal(m)
	if err != nil {
		return 0, err
	}
	ttl := strconv.Itoa(max(int(h.TTL.Seconds()), 1))
	n, err := anexar.Exec(ctx, h.Cliente, []string{chave(conversa)}, []string{string(dados), strconv.Itoa(h.Max), ttl}).AsInt64()
	if err != nil {
		return 0, fmt.Errorf("anexar em %q: %w", conversa, err)
	}
	return n, nil
}

// Ler devolve a conversa inteira, da mais antiga para a mais nova.
func (h *Historico) Ler(ctx context.Context, conversa string) ([]janela.Mensagem, error) {
	itens, err := h.Cliente.Do(ctx, h.Cliente.B().Lrange().Key(chave(conversa)).Start(0).Stop(-1).Build()).AsStrSlice()
	if err != nil {
		return nil, fmt.Errorf("ler %q: %w", conversa, err)
	}
	msgs := make([]janela.Mensagem, 0, len(itens))
	for _, it := range itens {
		var m janela.Mensagem
		if err := json.Unmarshal([]byte(it), &m); err != nil {
			return nil, fmt.Errorf("mensagem corrompida em %q: %w", conversa, err)
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// Apagar remove a conversa (por exemplo, a pedido do usuário).
func (h *Historico) Apagar(ctx context.Context, conversa string) error {
	return h.Cliente.Do(ctx, h.Cliente.B().Del().Key(chave(conversa)).Build()).Error()
}

// ApagarPrefixo remove todas as conversas cujo identificador começa com prefixo
// e devolve quantas eram. Usa SCAN (em lotes), nunca KEYS, que trava o servidor.
func (h *Historico) ApagarPrefixo(ctx context.Context, prefixo string) (int64, error) {
	var total int64
	var cursor uint64
	for {
		e, err := h.Cliente.Do(ctx, h.Cliente.B().Scan().Cursor(cursor).Match(chave(prefixo)+"*").Count(100).Build()).AsScanEntry()
		if err != nil {
			return total, err
		}
		if len(e.Elements) > 0 {
			n, err := h.Cliente.Do(ctx, h.Cliente.B().Del().Key(e.Elements...).Build()).AsInt64()
			if err != nil {
				return total, err
			}
			total += n
		}
		cursor = e.Cursor
		if cursor == 0 {
			return total, nil
		}
	}
}
