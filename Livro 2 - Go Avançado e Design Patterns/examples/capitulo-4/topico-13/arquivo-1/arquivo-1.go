package main

import (
	"bytes"
	"crypto"
	_ "crypto/sha256" // registra SHA-256 em crypto.Hash
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"mime"
)

func main() {
	// database/sql: sem importar nenhum driver, o registro esta vazio.
	fmt.Println("sql.Drivers():", sql.Drivers())
	_, err := sql.Open("postgres", "x")
	fmt.Println("sql.Open sem driver:", err)

	// image: importar image/png (para gerar o PNG) tambem registra o decodificador.
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	_ = png.Encode(&buf, img)
	_, formato, err := image.Decode(bytes.NewReader(buf.Bytes()))
	fmt.Println("image.Decode PNG:", formato, err)

	_, _, err = image.Decode(bytes.NewReader([]byte("\xff\xd8\xff\xe0 isto parece JPEG")))
	fmt.Println("image.Decode JPEG sem importar image/jpeg:", err)

	// crypto: Hash.Available diz se a implementacao foi registrada por um import.
	fmt.Println("SHA256 disponivel:", crypto.SHA256.Available())
	fmt.Println("SHA512 disponivel:", crypto.SHA512.Available())
	fmt.Println("BLAKE2b_256 disponivel:", crypto.BLAKE2b_256.Available())

	// mime: tipos por extensao vem de uma tabela global que voce pode estender.
	fmt.Printf(".golivro antes: %q\n", mime.TypeByExtension(".golivro"))
	_ = mime.AddExtensionType(".golivro", "application/x-golivro")
	fmt.Println(".golivro depois:", mime.TypeByExtension(".golivro"))
}
