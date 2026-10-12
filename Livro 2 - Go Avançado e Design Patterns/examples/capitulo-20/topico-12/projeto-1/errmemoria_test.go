package texto

import "testing"

func TestRepetir(t *testing.T) {
	casos := []struct {
		entrada string
		vezes   int
		quer    string
	}{
		{"ab", 3, "ababab"},
		{"x", 0, ""},
		{"", 5, ""},
		{"ção", 2, "çãoção"},
	}
	for _, c := range casos {
		t.Run(c.entrada, func(t *testing.T) {
			got, err := Repetir(c.entrada, c.vezes)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.quer {
				t.Errorf("Repetir(%q, %d) = %q, quer %q", c.entrada, c.vezes, got, c.quer)
			}
		})
	}
}
