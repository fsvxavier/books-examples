package a

import (
	"context"
	"net/http"
	"os/exec"
)

func busca(ctx context.Context, url string) error {
	resp, err := http.Get(url) // want `net/http.Get não aceita context.Context; use http.NewRequestWithContext e Client.Do`
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	return err
}

func roda(ctx context.Context) error {
	if err := exec.Command("ls").Run(); err != nil { // want `os/exec.Command não aceita context.Context; use exec.CommandContext`
		return err
	}
	return exec.CommandContext(ctx, "ls").Run()
}
