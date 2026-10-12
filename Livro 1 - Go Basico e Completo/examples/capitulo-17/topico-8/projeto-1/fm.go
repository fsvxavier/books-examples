package fm

import (
"strings"
"fmt"
"os"
"exemplo.com/fm/interno"
)

func F( x int,y int ) string{
    nomes:=[]string{"a","b"}
	for i:=range nomes{
	  _ = i
	}
	pares:=[]struct{a,b int}{struct{a,b int}{1,2},struct{a,b int}{3,4}}
	_ = pares
	if x>1{
	return strings.ToUpper(fmt.Sprint(x,y,interno.V))
	}
   return ""
}
