package main

import (
	"errors"
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.Atoi("123")
	fmt.Println(n+1, err)

	_, err = strconv.Atoi("12a")
	fmt.Println(err)
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		fmt.Println("func:", numErr.Func, "| entrada:", numErr.Num, "| causa:", numErr.Err)
	}

	_, err = strconv.Atoi("99999999999999999999")
	fmt.Println(err, errors.Is(err, strconv.ErrRange))

	f, _ := strconv.ParseFloat("3.14", 64)
	b, _ := strconv.ParseBool("true")
	i, _ := strconv.ParseInt("ff", 16, 64)
	u, err := strconv.ParseUint("-1", 10, 64)
	fmt.Println(f, b, i, u, err)
	_, err = strconv.ParseInt("200", 10, 8)
	fmt.Println(err)

	fmt.Println(strconv.Itoa(456)+" unidades", strconv.FormatInt(255, 2), strconv.FormatFloat(3.14159, 'f', 2, 64))
	fmt.Println(strconv.Quote("olá\n"), strconv.FormatBool(false))
}
