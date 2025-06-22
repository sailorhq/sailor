package main

import (
	"fmt"
	"net/http"

	"github.com/codekidx/sailor"
)

func main() {
	s := sailor.New("http://localhost:7766", "")
	err := s.Connect("payment", "gateway")
	if err != nil {
		panic(err)
	}

	v, _ := s.Get("name")
	fmt.Println("some value: ", v)

	http.ListenAndServe(":8080", nil)
}
