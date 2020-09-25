package main

import (
	"golang.org/x/tour/wc"
	"strings"
)

func WordCount(s string) map[string]int {
	ret:=make(map[string]int);
	words:=strings.Fields(s);
	for _,v:= range words {
		it,ok:=ret[v];
		if ok {
			ret[v]=it+1;
		} else {
			ret[v]=1;
		}
	}
	return ret;
}

func main() {
	wc.Test(WordCount)
}
