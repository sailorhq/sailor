package sailor

import "testing"

func TestConnect(t *testing.T) {
	s := New("http://localhost:7766", "")
	err := s.Connect("payment", "gateway")
	if err != nil {
		t.Error(err)
	}
}
