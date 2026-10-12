package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Preenchidas na compilação com -ldflags "-X main.versao=... -X main.commit=...".
var (
	versao = "dev"
	commit = "desconhecido"
)

func main() {
	fmt.Println("versão:", versao, "| commit:", commit)
	fmt.Println("go:", runtime.Version(), "|", runtime.GOOS+"/"+runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Println("módulo:", info.Main.Path)
		for _, s := range info.Settings {
			switch s.Key {
			case "CGO_ENABLED", "-trimpath", "GOARCH", "GOOS", "-ldflags":
				fmt.Printf("  %s=%s\n", s.Key, s.Value)
			}
		}
	}
}
