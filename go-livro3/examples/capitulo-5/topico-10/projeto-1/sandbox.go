package portao

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// Sandbox executa um comando com os limites que conseguimos impor sem
// privilégios de root: processo em grupo próprio (morto por inteiro no
// prazo), ambiente mínimo, saída limitada e, opcionalmente, sem rede.
type Sandbox struct {
	Dir      string        // diretório de trabalho e HOME do processo
	Prazo    time.Duration // tempo máximo de relógio
	SemRede  bool          // roda dentro de um namespace de rede vazio (unshare -rn)
	Jaula    bool          // raiz somente leitura, só Dir gravável, /home e /root ocultos (unshare -rm)
	MaxSaida int           // bytes de saída guardados (padrão 64 KiB)
	Extra    []string      // variáveis de ambiente adicionais (KEY=valor)
	Limites  []string      // argumentos de prlimit, por exemplo --nproc=64
}

// Resultado de uma execução em sandbox.
type Resultado struct {
	Saida   string
	Codigo  int
	Estouro bool // o prazo acabou e o grupo de processos foi morto
	Duracao time.Duration
}

type limitado struct {
	buf    bytes.Buffer
	max    int
	cortou bool
}

func (l *limitado) Write(p []byte) (int, error) {
	if l.buf.Len() < l.max {
		resta := l.max - l.buf.Len()
		if len(p) > resta {
			l.buf.Write(p[:resta])
			l.cortou = true
		} else {
			l.buf.Write(p)
		}
	} else {
		l.cortou = true
	}
	return len(p), nil // finge que consumiu: não bloqueia quem escreve
}

// scriptJaula roda dentro do novo namespace de montagem: torna a raiz somente
// leitura, devolve a escrita só para o diretório de trabalho e esconde os
// diretórios de usuário sob um tmpfs vazio. Falha com 97 se algo não montar.
const scriptJaula = `D=$1; shift
mount --bind "$D" "$D" && mount -o remount,ro,bind / / && mount -o remount,rw,bind "$D" "$D" &&
mount -t tmpfs tmpfs /home && mount -t tmpfs tmpfs /root || exit 97
cd "$D" && exec "$@"`

func (s Sandbox) flagsUnshare() string {
	switch {
	case s.Jaula && s.SemRede:
		return "-rmn"
	case s.Jaula:
		return "-rm"
	case s.SemRede:
		return "-rn"
	}
	return ""
}

func (s Sandbox) ambiente() []string {
	cache := filepath.Join(os.TempDir(), "guarda-gocache")
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + s.Dir,
		"TMPDIR=" + filepath.Join(s.Dir, "tmp"),
		"LANG=C.UTF-8",
		"GOCACHE=" + cache,
		"GOENV=off",
		"GOFLAGS=-mod=readonly",
		"GOPROXY=off",
		"GOTOOLCHAIN=local",
		"GOWORK=off",
		"GOMODCACHE=" + os.Getenv("GOMODCACHE"),
	}
	return append(env, s.Extra...)
}

// Rodar executa nome com args. Só devolve erro se o comando nem puder ser
// iniciado; falha do comando aparece em Codigo.
func (s Sandbox) Rodar(ctx context.Context, nome string, args ...string) (Resultado, error) {
	if s.Prazo == 0 {
		s.Prazo = 60 * time.Second
	}
	if s.MaxSaida == 0 {
		s.MaxSaida = 64 << 10
	}
	argv := append([]string{nome}, args...)
	if len(s.Limites) > 0 {
		argv = append(append([]string{"prlimit"}, s.Limites...), argv...)
	}
	if s.Jaula {
		argv = append([]string{"sh", "-c", scriptJaula, "sh", s.Dir}, argv...)
	}
	if flags := s.flagsUnshare(); flags != "" {
		argv = append([]string{"unshare", flags, "--"}, argv...)
	}
	ctx, cancela := context.WithTimeout(ctx, s.Prazo)
	defer cancela()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = s.Dir
	cmd.Env = s.ambiente()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second
	saida := &limitado{max: s.MaxSaida}
	cmd.Stdout, cmd.Stderr = saida, saida
	ini := time.Now()
	err := cmd.Run()
	r := Resultado{Saida: saida.buf.String(), Duracao: time.Since(ini)}
	if saida.cortou {
		r.Saida += "\n[saida truncada]"
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		r.Estouro = true
	}
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		r.Codigo = ee.ExitCode()
	case r.Estouro:
		r.Codigo = -1
	default:
		return r, err
	}
	return r, nil
}
