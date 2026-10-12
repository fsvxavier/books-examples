package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	memoria = 19 * 1024
	passes  = 2
)

func hashSenha(senha string, sal []byte) []byte {
	return argon2.IDKey([]byte(senha), sal, passes, memoria, 1, 32)
}

type usuario struct {
	sal  []byte
	hash []byte
}

type Servico struct {
	mu       sync.Mutex
	usuarios map[string]usuario
	sessoes  map[[32]byte]sessao // chave: SHA-256 do token (o token em si nunca e guardado)
	falhas   map[string]int
	// hash falso, para gastar o mesmo tempo quando o usuario nao existe
	salFalso, hashFalso []byte
}

type sessao struct {
	usuario string
	expira  time.Time
}

func NovoServico() *Servico {
	s := &Servico{usuarios: map[string]usuario{}, sessoes: map[[32]byte]sessao{}, falhas: map[string]int{}}
	s.salFalso = make([]byte, 16)
	rand.Read(s.salFalso)
	s.hashFalso = hashSenha("senha-que-ninguem-usa", s.salFalso)
	return s
}

func (s *Servico) Cadastrar(nome, senha string) {
	sal := make([]byte, 16)
	rand.Read(sal)
	s.usuarios[nome] = usuario{sal: sal, hash: hashSenha(senha, sal)}
}

func novoToken() (token string, indice [32]byte) {
	b := make([]byte, 32)
	rand.Read(b)
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, sha256.Sum256([]byte(token))
}

// Login devolve a MESMA resposta e gasta o MESMO tempo para "usuario inexistente"
// e "senha errada", para nao revelar quais usuarios existem.
func (s *Servico) Login(nome, senha string, comDummy bool) (string, bool) {
	s.mu.Lock()
	bloqueado := s.falhas[nome] >= 5
	u, existe := s.usuarios[nome]
	s.mu.Unlock()
	if bloqueado {
		return "", false
	}

	var calc, esperado []byte
	switch {
	case existe:
		calc, esperado = hashSenha(senha, u.sal), u.hash
	case comDummy:
		calc, esperado = hashSenha(senha, s.salFalso), s.hashFalso
		esperado = append([]byte{^esperado[0]}, esperado[1:]...) // nunca confere
	default:
		return "", false // atalho: responde rapido e entrega que o usuario nao existe
	}
	if subtle.ConstantTimeCompare(calc, esperado) != 1 || !existe {
		s.mu.Lock()
		s.falhas[nome]++
		s.mu.Unlock()
		return "", false
	}
	token, indice := novoToken()
	s.mu.Lock()
	s.falhas[nome] = 0
	s.sessoes[indice] = sessao{usuario: nome, expira: time.Now().Add(time.Hour)}
	s.mu.Unlock()
	return token, true
}

func (s *Servico) Autenticar(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ses, ok := s.sessoes[sha256.Sum256([]byte(token))]
	if !ok || time.Now().After(ses.expira) {
		return "", false
	}
	return ses.usuario, true
}

func (s *Servico) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Usuario, Senha string }
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
			http.Error(w, "requisicao invalida", http.StatusBadRequest)
			return
		}
		token, ok := s.Login(req.Usuario, req.Senha, true)
		if !ok {
			http.Error(w, "credenciais invalidas", http.StatusUnauthorized) // mensagem unica
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	})
	mux.HandleFunc("GET /perfil", func(w http.ResponseWriter, r *http.Request) {
		u, ok := s.Autenticar(r)
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "nao autenticado", http.StatusUnauthorized)
			return
		}
		fmt.Fprintf(w, "ola, %s\n", u)
	})
	return mux
}

func tempo(f func()) time.Duration {
	ini := time.Now()
	f()
	return time.Since(ini).Round(time.Millisecond)
}

func main() {
	s := NovoServico()
	s.Cadastrar("ana", "correct horse battery staple")

	// 1. enumeracao de usuarios pelo tempo de resposta
	fmt.Println("tempo de login (sem a defesa):")
	fmt.Println("  usuario existe, senha errada:", tempo(func() { s.Login("ana", "errada", false) }))
	fmt.Println("  usuario NAO existe:          ", tempo(func() { s.Login("fantasma", "errada", false) }), " <- entrega que nao existe")
	fmt.Println("tempo de login (com hash falso):")
	fmt.Println("  usuario existe, senha errada:", tempo(func() { s.Login("ana", "errada", true) }))
	fmt.Println("  usuario NAO existe:          ", tempo(func() { s.Login("fantasma", "errada", true) }))

	// 2. fluxo HTTP
	s.falhas = map[string]int{}
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()
	post := func(u, p string) (int, string) {
		corpo := fmt.Sprintf(`{"usuario":%q,"senha":%q}`, u, p)
		resp, err := http.Post(srv.URL+"/login", "application/json", strings.NewReader(corpo))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		var m map[string]string
		json.NewDecoder(resp.Body).Decode(&m)
		return resp.StatusCode, m["token"]
	}
	code, _ := post("ana", "errada")
	fmt.Println("\nlogin errado:      ", code)
	code, _ = post("fantasma", "x")
	fmt.Println("usuario inexistente:", code, "(mesma resposta)")
	code, token := post("ana", "correct horse battery staple")
	fmt.Println("login certo:       ", code, "token de", len(token), "caracteres")

	req, _ := http.NewRequest("GET", srv.URL+"/perfil", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err0 := http.DefaultClient.Do(req)
	if err0 != nil {
		panic(err0)
	}
	var b strings.Builder
	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	b.Write(buf[:n])
	resp.Body.Close()
	fmt.Printf("GET /perfil com token: %d %q\n", resp.StatusCode, strings.TrimSpace(b.String()))
	resp, err := http.Get(srv.URL + "/perfil")
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	fmt.Println("GET /perfil sem token:", resp.StatusCode, "WWW-Authenticate:", resp.Header.Get("WWW-Authenticate"))

	// 3. bloqueio apos 5 erros
	for range 5 {
		post("ana", "tentativa")
	}
	code, _ = post("ana", "correct horse battery staple")
	fmt.Println("senha certa apos 5 erros:", code, "(conta bloqueada)")
}
