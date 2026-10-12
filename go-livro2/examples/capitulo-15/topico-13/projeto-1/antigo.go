package desconto

// Antigo é o código original, com float64 e números mágicos.
func Antigo(valorTotal float64) float64 {
	if valorTotal > 1000 {
		return valorTotal * 0.1
	}
	return valorTotal * 0.05
}

const (
	limiarPremium     = 100_000 // R$ 1.000,00, em centavos
	percentualPremium = 10
	percentualPadrao  = 5
)

// NovoTruncado calcula em centavos e descarta a fração.
func NovoTruncado(centavos int64) int64 {
	if centavos > limiarPremium {
		return centavos * percentualPremium / 100
	}
	return centavos * percentualPadrao / 100
}

// NovoArredondado calcula em centavos e arredonda meio para cima.
func NovoArredondado(centavos int64) int64 {
	if centavos > limiarPremium {
		return (centavos*percentualPremium + 50) / 100
	}
	return (centavos*percentualPadrao + 50) / 100
}
