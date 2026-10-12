package main

import "os"

func lerTodos(caminhos []string) error {
	for _, c := range caminhos {
		f, err := os.Open(c)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	for i := 0; i < 3; i++ {
		f, err := os.Open("x")
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

func main() { _ = lerTodos(nil) }
