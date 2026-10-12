package main

import "fmt"

// ---------- IMPLEMENTACAO: os dispositivos ----------

// Dispositivo declara os passos primitivos que toda "plataforma" precisa oferecer.
type Dispositivo interface {
	Ligado() bool
	Ligar()
	Desligar()
	Volume() int
	DefinirVolume(v int)
	Canal() int
	DefinirCanal(c int)
}

type base struct {
	ligado bool
	volume int
	canal  int
}

func (b *base) Ligado() bool        { return b.ligado }
func (b *base) Ligar()              { b.ligado = true }
func (b *base) Desligar()           { b.ligado = false }
func (b *base) Volume() int         { return b.volume }
func (b *base) DefinirVolume(v int) { b.volume = max(0, min(100, v)) }
func (b *base) Canal() int          { return b.canal }
func (b *base) DefinirCanal(c int)  { b.canal = c }

type TV struct{ base }
type Radio struct{ base }

func (t *TV) String() string {
	return fmt.Sprintf("TV(ligada=%v vol=%d canal=%d)", t.ligado, t.volume, t.canal)
}
func (r *Radio) String() string {
	return fmt.Sprintf("Radio(ligado=%v vol=%d sintonia=%d)", r.ligado, r.volume, r.canal)
}

// ---------- ABSTRACAO: os controles remotos ----------

// Controle e a abstracao basica; guarda o dispositivo (a ponte).
type Controle struct{ disp Dispositivo }

func NovoControle(d Dispositivo) *Controle { return &Controle{disp: d} }

func (c *Controle) AlternarEnergia() {
	if c.disp.Ligado() {
		c.disp.Desligar()
	} else {
		c.disp.Ligar()
	}
}
func (c *Controle) VolumeMais()  { c.disp.DefinirVolume(c.disp.Volume() + 10) }
func (c *Controle) VolumeMenos() { c.disp.DefinirVolume(c.disp.Volume() - 10) }
func (c *Controle) CanalMais()   { c.disp.DefinirCanal(c.disp.Canal() + 1) }
func (c *Controle) CanalMenos()  { c.disp.DefinirCanal(c.disp.Canal() - 1) }

// ControleAvancado e uma abstracao REFINADA: acrescenta "mudo" sem tocar nos dispositivos.
type ControleAvancado struct{ *Controle }

func (c ControleAvancado) Mudo() { c.disp.DefinirVolume(0) }

func main() {
	tv, radio := &TV{}, &Radio{}

	basico := NovoControle(tv)
	basico.AlternarEnergia()
	basico.VolumeMais()
	basico.VolumeMais()
	basico.CanalMais()
	fmt.Println(tv)

	avancado := ControleAvancado{NovoControle(radio)}
	avancado.AlternarEnergia()
	avancado.VolumeMais()
	avancado.CanalMais()
	avancado.CanalMais()
	avancado.Mudo()
	fmt.Println(radio)

	// O mesmo controle avancado, agora ligado a outro dispositivo, em tempo de execucao.
	avancado.disp = tv
	avancado.Mudo()
	fmt.Println(tv)
}
