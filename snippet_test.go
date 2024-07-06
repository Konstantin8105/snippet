package snippet_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Konstantin8105/compare"
	"github.com/Konstantin8105/snippet"
)

const td = "./testdata/" // test directory

func Test(t *testing.T) {
	entries, err := os.ReadDir(td)
	if err != nil {
		t.Fatal(err)
	}
	for _, ent := range entries {
		name := ent.Name()
		if strings.HasSuffix(name, "ok") {
			t.Run(name, func(t *testing.T) {
				sns, err := snippet.Get(td + name)
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
				_, err := snippet.Get(td + name)
				if err == nil {
					t.Errorf("not fail test")
				}
				t.Logf("%v", err)
			})
		}
	}
	{
		// wrong filename
		_, err := snippet.Get("not valid filename")
		if err == nil {
			t.Fatalf("haven`t error for wrong filename")
		}
		t.Logf("%v", err)
	}
}
func TestCompare(t *testing.T) {
	t.Run("not.valid1", func(t *testing.T) {
		err := snippet.Compare("Not exist 1", "Not exist 2")
		if err == nil {
			t.Errorf("shall be error")
		}
		t.Logf("%v", err)
	})
	t.Run("not.valid2", func(t *testing.T) {
		err := snippet.Compare(td+"compare.expect", "Not exist 3")
		if err == nil {
			t.Errorf("shall be error")
		}
		t.Logf("%v", err)
	})
	t.Run("not.valid3", func(t *testing.T) {
		err := snippet.Compare(td+"compare.expect", td+"compare.fail.actual")
		if err == nil {
			t.Errorf("shall be error")
		}
		t.Logf("%v", err)
	})
	t.Run("valid", func(t *testing.T) {
		err := snippet.Compare(td+"compare.expect", td+"compare.actual")
		if err != nil {
			t.Error(err)
		}
	})
}

type mockTest struct {
	res error
}

func (m *mockTest) Errorf(format string, args ...any) {
	m.res = fmt.Errorf(format, args...)
}

func TestTest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		// snippet A
		snippet.Test(t, ".")
		// end A
	})
	t.Run("not.valid.file", func(t *testing.T) {
		m := new(mockTest)
		snippet.Test(m, "wrong data")
		if m.res == nil {
			t.Fatalf("shall be error")
		}
		t.Logf("%v", m.res)
	})
	t.Run("not.valid.snippet", func(t *testing.T) {
		old := snippet.ExpectSnippets
		defer func() {
			snippet.ExpectSnippets = old
		}()
		snippet.ExpectSnippets = "./fail.snippets"

		m := new(mockTest)
		snippet.Test(m, ".")
		if m.res == nil {
			t.Fatalf("shall be error")
		}
		t.Logf("%v", m.res)
	})
}
