// Package digitos guarda o cálculo dos dígitos verificadores. É privado ao módulo.
package digitos

// Extrair devolve só os dígitos de s.
func Extrair(s string) []int {
	var ds []int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			ds = append(ds, int(r-'0'))
		}
	}
	return ds
}

// Verificador calcula um dígito verificador a partir dos dígitos já conhecidos.
func Verificador(base []int) int {
	peso := len(base) + 1
	soma := 0
	for _, d := range base {
		soma += d * peso
		peso--
	}
	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}
