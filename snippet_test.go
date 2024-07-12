package snippet_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
					bs := []byte(
						fmt.Sprintf("%s\n%s\n%s\n",
							sns[i].Start,
							sns[i],
							sns[i].End,
						))
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
		err := snippet.Compare("Not exist 2", "Not exist 1", true)
		if err == nil {
			t.Errorf("shall be error")
		}
		t.Logf("%v", err)
	})
	t.Run("not.valid2", func(t *testing.T) {
		err := snippet.Compare("Not exist 3", td+"compare.expect", true)
		if err == nil {
			t.Errorf("shall be error")
		}
		t.Logf("%v", err)
	})
	t.Run("not.valid3", func(t *testing.T) {
		err := snippet.Compare(td+"compare.fail.actual", td+"compare.expect", true)
		if err == nil {
			t.Errorf("shall be error")
		}
		t.Logf("%v", err)
	})
	t.Run("valid", func(t *testing.T) {
		err := snippet.Compare(td+"compare.actual", td+"compare.expect", true)
		if err != nil {
			t.Error(err)
		}
	})
}

type mockTest struct {
	log string
	res error
}

func (m *mockTest) Errorf(format string, args ...any) {
	m.res = fmt.Errorf(format, args...)
}

func (m *mockTest) Logf(format string, args ...any) {
	m.log += fmt.Sprintf(format, args...)
}

func (m mockTest) String() string {
	return fmt.Sprintf("Error: %v\nLog: %s", m.res, m.log)
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
		t.Logf("%s", m)
	})
	t.Run("not.valid.snippet.folder", func(t *testing.T) {
		// snippet B
		old := snippet.ExpectSnippets
		defer func() {
			snippet.ExpectSnippets = old
		}()
		snippet.ExpectSnippets = "./fail.snippets"
		// end B

		m := new(mockTest)
		snippet.Test(m, ".")
		if m.res == nil {
			t.Fatalf("shall be error")
		}
		t.Logf("%s", m)
	})
	t.Run("not.valid.snippet.file", func(t *testing.T) {
		// snippet B
		old := snippet.ExpectSnippets
		defer func() {
			snippet.ExpectSnippets = old
		}()
		snippet.ExpectSnippets = "./fail.snippets"
		// end B

		m := new(mockTest)
		snippet.Test(m, "snippet_test.go")
		if m.res == nil {
			t.Fatalf("shall be error")
		}
		t.Logf("%s", m)
	})
}

func TestUpdate(t *testing.T) {
	var err error

	_, err = exec.Command("cp", "./testdata/cli.actual", "./testdata/cli.actual.1").Output()
	if err != nil {
		t.Fatal(err)
	}

	{
		act, err := os.ReadFile("./testdata/cli.result")
		if err != nil {
			t.Fatal(err)
		}
		act1, err := os.ReadFile("./testdata/cli.actual.1")
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(act, act1) {
			t.Fatal("files are same")
		}
	}

	err = snippet.Compare("./testdata/cli.actual.1", "./testdata/cli.expect", true)
	if err == nil {
		t.Fatal("cannot find diff")
	}
	{
		act1 := []byte(fmt.Sprintf("%v", err))
		compare.Test(t, "./testdata/cli.diff", act1)
	}

	err = snippet.Compare("./testdata/cli.actual.1", "./testdata/cli.expect", false)
	if err != nil {
		t.Fatal(err)
	}
	{
		act1, err := os.ReadFile("./testdata/cli.actual.1")
		if err != nil {
			t.Fatal(err)
		}
		compare.Test(t, "./testdata/cli.result", act1)
	}

	_, err = exec.Command("rm", "-f", "./testdata/cli.actual.1").Output()
	if err != nil {
		t.Fatal(err)
	}
}
