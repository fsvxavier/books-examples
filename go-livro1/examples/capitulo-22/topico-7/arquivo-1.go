package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func main() {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	srv.TLS = &tls.Config{MinVersion: tls.VersionTLS13}
	srv.StartTLS()
	defer srv.Close()

	tentar := func(rotulo string, cfg *tls.Config) {
		cfg.RootCAs = nil
		cliente := srv.Client()
		tr := cliente.Transport.(*http.Transport).Clone()
		tr.TLSClientConfig.MinVersion = cfg.MinVersion
		tr.TLSClientConfig.MaxVersion = cfg.MaxVersion
		tr.TLSClientConfig.CurvePreferences = cfg.CurvePreferences
		resp, err := (&http.Client{Transport: tr}).Get(srv.URL)
		if err != nil {
			fmt.Printf("%-34s erro: %v\n", rotulo, err)
			return
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		st := resp.TLS
		fmt.Printf("%-34s %s  suite=%s  troca=%s\n", rotulo, tls.VersionName(st.Version), tls.CipherSuiteName(st.CipherSuite), st.CurveID)
	}

	tentar("padrao do cliente", &tls.Config{})
	tentar("so X25519 (classico)", &tls.Config{CurvePreferences: []tls.CurveID{tls.X25519}})
	tentar("cliente limitado a TLS 1.2", &tls.Config{MaxVersion: tls.VersionTLS12})
}
