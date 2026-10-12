package main

import (
	"fmt"
	"net"
	"os/user"
)

func main() {
	addrs, err := net.LookupHost("localhost")
	fmt.Println("localhost:", addrs, err)
	u, err := user.Current()
	if err != nil {
		fmt.Println("user:", err)
		return
	}
	fmt.Println("usuario:", u.Username)
}
