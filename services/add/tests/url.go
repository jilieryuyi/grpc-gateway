package main

import (
	"strings"
	"fmt"
)

func main() {
	url1 := "/sum/100/200"
	url2 := "/sum/{a}/{b}"

	a1 := strings.Split(url1, "/")
	a2 := strings.Split(url2, "/")

	params := make(map[string]string)
	for i, k := range a2 {
		if strings.Contains(k, "{") {
			k = strings.Trim(k, "{")
			k = strings.Trim(k, "}")
			params[k] = a1[i]
		}
	}

	fmt.Printf("%+v\n", params)
}
