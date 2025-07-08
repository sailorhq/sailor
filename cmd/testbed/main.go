package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codekidx/sailor"
	"github.com/codekidx/sailor/internal/types"
)

func main() {
	err := sailor.Connect("http://localhost:7766", "sailor", "backend-core", types.SailorOpts{
		AccessKey:      "sailor-UpYponVq7m68mBMa",
		SecretKey:      "secret-3RpsGcRAsSj9RnhhJPh7ZETfILCxCygE",
		RefreshTimeout: 5 * time.Second,
		Logging:        true,
	})
	if err != nil {
		panic(err)
	}

	s := sailor.Instance()
	v, _ := s.Get("app")
	fmt.Println("some value: ", v)
	s.Release()

	http.ListenAndServe(":8080", nil)
}
