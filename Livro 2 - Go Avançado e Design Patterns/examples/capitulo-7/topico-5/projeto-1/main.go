package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"exemplos/livro2/capitulo-7/topico-5/projeto-1/subsistema"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(g []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(g) == 0 {
				return slog.Attr{}
			}
			return a
		},
	}))
	est := subsistema.NovoEstoque(map[string]int{"CANETA": 10, "CADERNO": 2})
	pg := &subsistema.Pagamento{}
	aviso := &subsistema.Notificacao{}
	loja := NovaLoja(est, subsistema.Frete{}, pg, subsistema.Fiscal{}, aviso, log)

	pedido := Pedido{Cliente: "ana", Email: "ana@exemplo.com", CEP: "01310100", Cartao: "4111",
		Linhas: []Linha{{"CANETA", 3, 450}, {"CADERNO", 1, 2990}}}

	fmt.Println("1) compra normal")
	rec, err := loja.Finalizar(context.Background(), pedido)
	fmt.Printf("   recibo=%+v err=%v\n   estoque: caneta=%d caderno=%d\n", rec, err, est.Saldo("CANETA"), est.Saldo("CADERNO"))

	fmt.Println("2) caderno acaba: a reserva de caneta e desfeita")
	pedido.Linhas = []Linha{{"CANETA", 2, 450}, {"CADERNO", 5, 2990}}
	_, err = loja.Finalizar(context.Background(), pedido)
	fmt.Printf("   err=%v\n   sem estoque? %v | caneta=%d (nao mudou)\n", err, errors.Is(err, subsistema.ErrSemEstoque), est.Saldo("CANETA"))

	fmt.Println("3) cartao recusado: estoque volta")
	pg.Recusar = true
	pedido.Linhas = []Linha{{"CANETA", 2, 450}}
	_, err = loja.Finalizar(context.Background(), pedido)
	fmt.Printf("   err=%v | caneta=%d\n", err, est.Saldo("CANETA"))
	pg.Recusar = false

	fmt.Println("4) nota fiscal falha: estorna a cobranca e devolve o estoque")
	loja2 := NovaLoja(est, subsistema.Frete{}, pg, subsistema.Fiscal{Fora: true}, aviso, log)
	antes := pg.Ativas()
	_, err = loja2.Finalizar(context.Background(), pedido)
	fmt.Printf("   err=%v | cobrancas ativas: %d -> %d | caneta=%d\n", err, antes, pg.Ativas(), est.Saldo("CANETA"))

	fmt.Println("5) aviso falha: a compra continua valida")
	aviso.Falhar = true
	rec, err = loja.Finalizar(context.Background(), pedido)
	fmt.Printf("   nota=%s err=%v\n", rec.Nota, err)
}
