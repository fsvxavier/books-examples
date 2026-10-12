package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var chave = []byte("chave-hmac-com-pelo-menos-32-bytes!!")

type Claims struct {
	Papel string `json:"papel"`
	jwt.RegisteredClaims
}

func emitir(usuario string, validade time.Duration) string {
	agora := time.Now()
	c := Claims{
		Papel: "leitor",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usuario,
			Issuer:    "loja-api",
			Audience:  jwt.ClaimStrings{"loja-web"},
			IssuedAt:  jwt.NewNumericDate(agora),
			ExpiresAt: jwt.NewNumericDate(agora.Add(validade)),
		},
	}
	s, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(chave)
	return s
}

// validar exige o algoritmo, o emissor e a audiencia: nunca confie no cabecalho do token.
func validar(tokenStr string) (*Claims, error) {
	var c Claims
	t, err := jwt.ParseWithClaims(tokenStr, &c, func(*jwt.Token) (any, error) { return chave, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer("loja-api"),
		jwt.WithAudience("loja-web"),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("token invalido")
	}
	return &c, nil
}

func b64(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }

func main() {
	tok := emitir("ana", time.Hour)
	partes := strings.Split(tok, ".")
	cab, _ := base64.RawURLEncoding.DecodeString(partes[0])
	corpo, _ := base64.RawURLEncoding.DecodeString(partes[1])
	fmt.Println("cabecalho:", string(cab))
	fmt.Println("corpo:    ", string(corpo))
	fmt.Println("(o corpo e apenas base64, qualquer um le: NAO ponha segredos nele)")

	c, err := validar(tok)
	fmt.Printf("\nvalido:        sub=%s papel=%s err=%v\n", c.Subject, c.Papel, err)

	_, err = validar(emitir("ana", -time.Minute))
	fmt.Println("expirado:     ", err)

	// adulteracao: troca o papel no corpo, mantem a assinatura antiga
	falso := partes[0] + "." + b64(strings.Replace(string(corpo), "leitor", "admin", 1)) + "." + partes[2]
	_, err = validar(falso)
	fmt.Println("corpo adulterado:", err)

	// ataque alg=none: token sem assinatura, com papel admin
	none := b64(`{"alg":"none","typ":"JWT"}`) + "." +
		b64(`{"papel":"admin","sub":"ana","iss":"loja-api","aud":"loja-web","exp":9999999999}`) + "."
	_, err = validar(none)
	fmt.Println("alg=none:        ", err)

	// confusao de algoritmo: HS512 assinado com a mesma chave nao e aceito
	t512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{RegisteredClaims: jwt.RegisteredClaims{
		Subject: "ana", Issuer: "loja-api", Audience: jwt.ClaimStrings{"loja-web"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString(chave)
	_, err = validar(t512)
	fmt.Println("HS512 nao esperado:", err)

	// emissor errado
	outro, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{RegisteredClaims: jwt.RegisteredClaims{
		Subject: "ana", Issuer: "outro-servico", Audience: jwt.ClaimStrings{"loja-web"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString(chave)
	_, err = validar(outro)
	fmt.Println("emissor errado:  ", err)

	// o mesmo principio, na mao: assinatura de webhook com HMAC
	corpoHook := []byte(`{"evento":"pagamento.aprovado","id":"p_123"}`)
	ts := fmt.Sprint(time.Now().Unix())
	m := hmac.New(sha256.New, chave)
	m.Write([]byte(ts + "."))
	m.Write(corpoHook)
	fmt.Printf("\nX-Assinatura: t=%s,v1=%x\n", ts, m.Sum(nil))
}
