package main

import (
	"fmt"
	"net/http"

	"github.com/codekidx/sailor"
)

func main() {
	err := sailor.Connect("http://localhost:7766", "payment", "gateway")
	if err != nil {
		panic(err)
	}

	s := sailor.Instance()
	defer s.Release()

	v, _ := s.Get("name")
	fmt.Println("some value: ", v)

	http.ListenAndServe(":8080", nil)
}
