// Package snippet compare snippets in golang file.
//
// Snippets lines without specific upcase/lower case of letters.
// Name of snippet is single word.
//
// Format of snippet:
//
//	// Snippet Name
//	some code
//	// End Name
package snippet

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	prefixName = "//"
	startName  = "snippet"
	endName    = "end"
)

type status bool

const (
	start status = true
	end   status = false
)

type record struct {
	s    status
	line int
	name string
}

type snippet struct {
	name string
	pos  string
	code []string
}

func (s snippet) String() string {
	var out string
	out += fmt.Sprintf("%s %s %s\n", prefixName, startName, s.name)
	out += strings.Join(s.code, "\n")
	out += fmt.Sprintf("\n%s %s %s\n", prefixName, endName, s.name)
	return out
}

func Get(filename string) (snippets []snippet, err error) {
	// read file
	dat, err := os.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("%s cannot open. %w", filename, err)
		return
	}
	dat = bytes.ReplaceAll(dat, []byte("\r"), []byte{})

	var records []record

	lines := strings.Split(string(dat), "\n")
	for i := range lines {
		line := strings.TrimSpace(lines[i])
		fs := strings.Fields(line)
		if len(fs) < 2 {
			continue
		}
		fs[1] = strings.ToLower(fs[1])
		if 3 < len(fs) {
			if fs[0] == prefixName && (fs[1] == startName || fs[1] == endName) {
				err = fmt.Errorf(
					"%s:%d: snipet name cannot be with spaces",
					filename, i+1,
				)
				return
			}
		}
		if len(fs) != 3 {
			continue
		}
		if fs[0] == prefixName && fs[1] == startName {
			records = append(records, record{s: start, line: i + 1, name: fs[2]})
			continue
		}
		if fs[0] == prefixName && fs[1] == endName {
			records = append(records, record{s: end, line: i, name: fs[2]})
			continue
		}
	}

	if len(records) == 0 {
		return
	}

	// check start and end
	for i := range records {
		if i%2 == 0 && records[i].s != start {
			err = fmt.Errorf("%s:%d: Error: is not start", filename, i+1)
			return
		}
		if i%2 != 0 && records[i].s != end {
			err = fmt.Errorf("%s:%d: Error: is not end", filename, i+1)
			return
		}
	}

	// check names
	for i := 1; i < len(records); i += 2 {
		if !strings.EqualFold(records[i-1].name, records[i].name) {
			err = errors.Join(
				fmt.Errorf(
					"%s:%d: is not same name with end: `%s`",
					filename,
					records[i-1].line+1,
					records[i-1].name,
				),
				fmt.Errorf(
					"%s:%d: is not same name with start: `%s`",
					filename,
					records[i].line+1,
					records[i].name,
				),
			)
			return
		}
	}

	// amount
	if len(records)%2 != 0 {
		err = fmt.Errorf("%s:%d: not valid for last snippet",
			filename,
			records[len(records)-1].line+1,
		)
		return
	}

	// create snippets
	for i := 1; i < len(records); i += 2 {
		snippets = append(snippets, snippet{
			name: records[i-1].name,
			pos:  fmt.Sprintf("%s:%d:", filename, records[i-1].line),
			code: lines[records[i-1].line:records[i].line],
		})
	}

	// clean code
	for i := range snippets {
		cs := snippets[i].code
		for i := range cs {
			cs[i] = strings.TrimSpace(cs[i])
		}
		snippets[i].code = cs
	}

	return
}

func Compare(expectFilename, actualFilename string) (err error) {
	expect, err := Get(expectFilename)
	if err != nil {
		return
	}

	actual, err := Get(actualFilename)
	if err != nil {
		return
	}

	for _, act := range actual {
		found := false
		index := -1
		for ie, exp := range expect {
			if strings.EqualFold(act.name, exp.name) {
				found = true
				index = ie
			}
		}
		if !found {
			err = errors.Join(err,
				fmt.Errorf("%s cannot find snippet with name `%s`",
					act.pos,
					act.name,
				))
			continue
		}
		ac := strings.Join(act.code, "\n")
		ec := strings.Join(expect[index].code, "\n")
		if ac != ec {
			err = errors.Join(err,
				fmt.Errorf("%s code is not same",
					act.pos,
				))
			continue
		}
	}
	return
}

var ExpectSnippets = "./expect.snippets"

// Test check only '*.go' files in `folder` with subfolders.
// Location with expected snippets in file "ExpectSnippets"
func Test(t interface {
	Errorf(format string, args ...any)
}, folder string) {
	var errs error
	expect := ExpectSnippets
	err := filepath.Walk(folder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			errs = errors.Join(errs, Compare(expect, path))
			return nil
		})
	err = errors.Join(err, errs)
	if err != nil {
		t.Errorf("%w", err)
	}
}
