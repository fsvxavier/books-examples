package molde

type Passos interface {
	Preparar() error
	Executar() error
	Validar() error
}

type Padrao struct{}

func (Padrao) Validar() error { return nil }

func Rodar(p Passos) error {
	if err := p.Preparar(); err != nil {
		return err
	}
	if err := p.Executar(); err != nil {
		return err
	}
	return p.Validar()
}
