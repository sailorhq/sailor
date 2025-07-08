package sailor

import (
	"testing"

	"github.com/codekidx/sailor/internal/types"
)

func TestConnect(t *testing.T) {
	err := Connect("http://localhost:7766", "sailor", "backend-core", types.SailorOpts{
		AccessKey: "sailor-UpYponVq7m68mBMa",
		SecretKey: "secret-3RpsGcRAsSj9RnhhJPh7ZETfILCxCygE",
	})
	if err != nil {
		t.Error(err)
	}

	s := Instance()
	defer s.Release()

	v, err := s.Get("app")
	if err != nil {
		t.Error(err)
	} else {
		t.Log("some value: ", v)
	}
}
