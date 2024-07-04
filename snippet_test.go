package snippet

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Konstantin8105/compare"
)

func Test(t *testing.T) {
	td := "./testdata/" // test directory

	entries, err := os.ReadDir(td)
	if err != nil {
		t.Fatal(err)
	}
	for _, ent := range entries {
		name := ent.Name()
		if strings.HasSuffix(name, "ok") {
			t.Run(name, func(t *testing.T) {
				sns, err := Get(td + name)
				if err != nil {
					t.Fatal(err)
				}
				for i := range sns {
					filename := fmt.Sprintf("%s%s.view%d", td, name, i)
					bs := []byte(sns[i].String())
					compare.Test(t, filename, bs)
				}
			})
		}
		if strings.HasSuffix(name, "fail") {
			t.Run(name, func(t *testing.T) {
				_, err := Get(td + name)
				if err == nil {
					t.Errorf("not fail test")
				}
			})
		}
	}
	{
		// wrong filename
		_, err := Get("not valid filename")
		if err == nil {
			t.Fatalf("haven`t error for wrong filename")
		}
	}
	// Compare(t, "./snippet_test.go")
}
