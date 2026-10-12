package soma

// Soma e tao pequena que o compilador a expande no ponto de chamada.
func Soma(a, b int) int { return a + b }

// Fib e um trabalho de verdade para o laco nao desaparecer.
func Fib(n int) int {
	if n < 2 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}
