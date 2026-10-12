package inseguro

import (
	"crypto/md5"
	"crypto/tls"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
)

const dbPassword = "Sup3rS3cr3t!2024" // credencial no codigo

func HashDeSenha(s string) []byte {
	h := md5.Sum([]byte(s)) // hash quebrado
	return h[:]
}

func TokenDeSessao() int {
	return rand.Int() // gerador previsivel
}

func Buscar(db *sql.DB, nome string) {
	_, _ = db.Query("SELECT * FROM usuarios WHERE nome = '" + nome + "'") // SQL por concatenacao
}

func Executar(arquivo string) {
	_ = exec.Command("sh", "-c", "cat "+arquivo).Run() // injecao de comando
}

func Ler(caminho string) ([]byte, error) {
	return os.ReadFile(caminho) // travessia de diretorio
}

func Cliente() *http.Client {
	return &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // desliga a verificacao do certificado
	}}
}

func Servir() {
	fmt.Println(http.ListenAndServe(":8080", nil)) // sem timeouts
}

func Gravar() {
	_ = os.WriteFile("dados.txt", []byte("x"), 0o777) // permissao ampla demais
}
