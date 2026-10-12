package ruim

import "time"

var (
	fila     []string
	enviados int
)

func Enfileirar(msg string) {
	fila = append(fila, msg)
}

func Iniciar() {
	go func() {
		for {
			if len(fila) > 0 {
				m := fila[0]
				fila = fila[1:]
				enviar(m)
				enviados++
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func Enviados() int { return enviados }

func enviar(m string) {
	if m == "" {
		panic("mensagem vazia")
	}
	time.Sleep(5 * time.Millisecond) // "envio"
}
