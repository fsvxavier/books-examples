package main

import (
	"bytes"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	"image/png"
)

func main() {
	// 1. image.Decode: os formatos se REGISTRAM nos imports com "_"
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.RGBA{255, 0, 0, 255})

	var pngBuf, gifBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	gif.Encode(&gifBuf, img, nil)

	for _, c := range []struct {
		nome  string
		dados []byte
	}{{"png", pngBuf.Bytes()}, {"gif", gifBuf.Bytes()}} {
		_, formato, err := image.Decode(bytes.NewReader(c.dados))
		fmt.Printf("image.Decode(%s) -> formato=%q err=%v\n", c.nome, formato, err)
	}
	_, formato, err := image.Decode(bytes.NewReader([]byte("isto nao e uma imagem")))
	fmt.Printf("image.Decode(lixo) -> formato=%q err=%v\n", formato, err)

	// 2. crypto.Hash: um enum que FABRICA o hash correspondente
	for _, h := range []crypto.Hash{crypto.SHA256, crypto.SHA512, crypto.MD5} {
		fmt.Printf("%-8s disponivel=%v", h, h.Available())
		if h.Available() {
			fmt.Printf(" tamanho=%d bytes", h.New().Size())
		}
		fmt.Println()
	}

	// 3. fabricar a partir de um nome vindo de configuracao
	nomes := []string{"sha256", "sha512", "blake3"}
	porNome := map[string]crypto.Hash{"sha256": crypto.SHA256, "sha512": crypto.SHA512}
	for _, n := range nomes {
		h, ok := porNome[n]
		if !ok || !h.Available() {
			fmt.Printf("hash %q: nao suportado\n", n)
			continue
		}
		sum := h.New()
		sum.Write([]byte("go"))
		fmt.Printf("hash %q: %x...\n", n, sum.Sum(nil)[:6])
	}
}
