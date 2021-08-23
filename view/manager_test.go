package view_test

import (
	"os"
	"testing"

	"github.com/enorith/http/view"
)

func TestParse(t *testing.T) {
	view.WithDefault(os.DirFS("./test_views"), "html")
	b, e := view.DefaultManager.Parse("test")
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%s", b)
}
