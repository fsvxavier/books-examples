package e06

import (
	"bytes"
	"strconv"
	"strings"
)

func linhasDe(n int) []string {
	l := make([]string, n)
	for i := range l {
		l[i] = "linha número " + strconv.Itoa(i)
	}
	return l
}

func Concat(linhas []string) string {
	r := ""
	for _, l := range linhas {
		r += l + "\n"
	}
	return r
}

func Builder(linhas []string) string {
	var sb strings.Builder
	for _, l := range linhas {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	return sb.String()
}

func BuilderGrow(linhas []string) string {
	n := 0
	for _, l := range linhas {
		n += len(l) + 1
	}
	var sb strings.Builder
	sb.Grow(n)
	for _, l := range linhas {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	return sb.String()
}

func Buffer(linhas []string) string {
	var b bytes.Buffer
	for _, l := range linhas {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return b.String()
}

func Join(linhas []string) string {
	return strings.Join(linhas, "\n") + "\n"
}
