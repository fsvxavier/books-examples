package infra

import (
	"errors"
	"fmt"
	"strings"
)

// ---- Interfaces dos PRODUTOS ----

type Armazenamento interface {
	// Salvar guarda os dados e devolve a URI pela qual eles podem ser lidos.
	Salvar(nome string, dados []byte) (uri string)
	Ler(uri string) ([]byte, error)
}

type Fila interface {
	Publicar(uri string)
	Consumir() (uri string, ok bool)
}

type Notificador interface{ Avisar(uri string) string }

// ---- A FABRICA ABSTRATA: uma operacao de criacao por produto da familia ----

type Fabrica interface {
	NovoArmazenamento() Armazenamento
	NovaFila() Fila
	NovoNotificador() Notificador
}

var ErrEsquema = errors.New("esquema de URI desconhecido")

// ---- Familia 1: Local (tudo em memoria, esquema mem://) ----

type armLocal struct{ m map[string][]byte }

func (a *armLocal) Salvar(nome string, d []byte) string {
	a.m[nome] = d
	return "mem://" + nome
}

func (a *armLocal) Ler(uri string) ([]byte, error) {
	nome, ok := strings.CutPrefix(uri, "mem://")
	if !ok {
		return nil, fmt.Errorf("%w: %q (armazenamento local so le mem://)", ErrEsquema, uri)
	}
	d, ok := a.m[nome]
	if !ok {
		return nil, fmt.Errorf("objeto %q nao encontrado", nome)
	}
	return d, nil
}

type filaSlice struct{ itens []string }

func (f *filaSlice) Publicar(u string) { f.itens = append(f.itens, u) }
func (f *filaSlice) Consumir() (string, bool) {
	if len(f.itens) == 0 {
		return "", false
	}
	u := f.itens[0]
	f.itens = f.itens[1:]
	return u, true
}

type notifLocal struct{}

func (notifLocal) Avisar(u string) string { return "[local] novo objeto em " + u }

type FabricaLocal struct{}

func (FabricaLocal) NovoArmazenamento() Armazenamento { return &armLocal{m: map[string][]byte{}} }
func (FabricaLocal) NovaFila() Fila                   { return &filaSlice{} }
func (FabricaLocal) NovoNotificador() Notificador     { return notifLocal{} }

// ---- Familia 2: Nuvem (simulada; esquema nuvem://bucket/...) ----

type armNuvem struct {
	bucket string
	m      map[string][]byte
}

func (a *armNuvem) Salvar(nome string, d []byte) string {
	a.m[nome] = d
	return "nuvem://" + a.bucket + "/" + nome
}

func (a *armNuvem) Ler(uri string) ([]byte, error) {
	resto, ok := strings.CutPrefix(uri, "nuvem://"+a.bucket+"/")
	if !ok {
		return nil, fmt.Errorf("%w: %q (esperava nuvem://%s/)", ErrEsquema, uri, a.bucket)
	}
	d, ok := a.m[resto]
	if !ok {
		return nil, fmt.Errorf("objeto %q nao encontrado", resto)
	}
	return d, nil
}

type notifNuvem struct{}

func (notifNuvem) Avisar(u string) string { return "[nuvem] evento ObjectCreated " + u }

type FabricaNuvem struct{ Bucket string }

func (f FabricaNuvem) NovoArmazenamento() Armazenamento {
	return &armNuvem{bucket: f.Bucket, m: map[string][]byte{}}
}
func (FabricaNuvem) NovaFila() Fila               { return &filaSlice{} } // a fila nao muda de esquema
func (FabricaNuvem) NovoNotificador() Notificador { return notifNuvem{} }

// ---- O CLIENTE: usa so as interfaces, nunca os tipos concretos ----

// Importar guarda um arquivo, enfileira a URI e avisa. Depois, um consumidor le.
func Importar(f Fabrica, nome string, dados []byte) (aviso string, lido []byte, err error) {
	arm, fila, notif := f.NovoArmazenamento(), f.NovaFila(), f.NovoNotificador()
	uri := arm.Salvar(nome, dados)
	fila.Publicar(uri)
	aviso = notif.Avisar(uri)
	proxima, _ := fila.Consumir()
	lido, err = arm.Ler(proxima)
	return aviso, lido, err
}

// ImportarMisturado monta a pipeline com pecas de FAMILIAS DIFERENTES: o bug que a
// fabrica abstrata existe para evitar.
func ImportarMisturado(nome string, dados []byte) error {
	armazenamentoNuvem := FabricaNuvem{Bucket: "relatorios"}.NovoArmazenamento()
	armazenamentoLocal := FabricaLocal{}.NovoArmazenamento()
	uri := armazenamentoNuvem.Salvar(nome, dados) // grava na nuvem...
	_, err := armazenamentoLocal.Ler(uri)         // ...e tenta ler do local
	return err
}
