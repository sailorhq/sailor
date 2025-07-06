package sailor

import (
	"testing"
)

func TestConnect(t *testing.T) {
	err := Connect("http://localhost:7766", "sailor", "backend-core")
	if err != nil {
		t.Error(err)
	}

	s := Instance()
	defer s.Release()

	v, _ := s.Get("app")
	t.Log("some value: ", v)
}
