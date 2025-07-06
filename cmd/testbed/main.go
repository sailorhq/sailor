package main

import (
	"fmt"
	"net/http"

	"github.com/codekidx/sailor"
)

func main() {
	err := sailor.Connect("http://localhost:7766", "sailor", "backend-core")
	if err != nil {
		panic(err)
	}

	s := sailor.Instance()
	defer s.Release()

	v, _ := s.Get("app")
	fmt.Println("some value: ", v)

	http.ListenAndServe(":8080", nil)
}
