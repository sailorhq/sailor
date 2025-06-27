package sailor

import (
	"testing"
)

func TestConnect(t *testing.T) {
	err := Connect("http://localhost:7766", "payment", "gateway")
	if err != nil {
		t.Error(err)
	}
}
