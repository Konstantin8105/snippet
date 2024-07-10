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
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/Konstantin8105/compare"
)

// Example folding code part for VSCode:
//
//	// #region Math functions
//	function add(a, b) {
//		return a + b
//	}
//	// #endregion
//
// Example folding code for Vim:
//
//	set foldmethod=marker
//	set foldmarker={{{,}}}
//	{{{
//		fold level here is 2
//	}}}
const (
	prefixName = "//"
	startName  = "snippet"
	endName    = "end"
)

// Position line of code
type Position struct {
	Filename string
	Line     int
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d", p.Filename, p.Line)
}

/*
Snippet inside code
*/
type Snippet struct {
	Name       string
	Start, End Position
	Code       []string
}

func (s Snippet) String() string {
	var out string
	out += fmt.Sprintf("%s %s %s\n", prefixName, startName, s.Name)
	if 0 < len(s.Code) {
		out += strings.Join(s.Code, "\n")
		out += "\n"
	}
	out += fmt.Sprintf("%s %s %s", prefixName, endName, s.Name)
	return out
}

// update snippets in file `filename`
func update(filename string, sn []Snippet) (err error) {
	const op = "Update"

	// snippet deferfunc
	defer func() {
		if err != nil {
			err = errors.Join(
				fmt.Errorf("%s. Filename : `%s`", op, filename),
				err,
			)
		}
	}()
	// end deferfunc

	if len(sn) == 0 {
		return
	}

	actual, err := Get(filename)
	if err != nil {
		return
	}
	if len(actual) == 0 {
		return
	}
	sort.Slice(actual, func(i, j int) bool {
		return actual[i].Start.Line < actual[j].Start.Line
	})

	var differr error // error of diff
	changed := false
	for i := range actual {
		for _, exp := range sn {
			if !strings.EqualFold(actual[i].Name, exp.Name) {
				continue
			}
			if strings.Join(actual[i].Code, "\n") == strings.Join(exp.Code, "\n") {
				continue
			}
			changed = true
			{
				errd := compare.Diff(
					[]byte(strings.Join(actual[i].Code, "\n")),
					[]byte(strings.Join(exp.Code, "\n")),
				)
				if errd != nil {
					differr = errors.Join(differr,
						fmt.Errorf("Snippet name: %s", exp.Name),
						errd,
					)
				}
			}
			actual[i].Code = exp.Code
		}
	}
	if !changed {
		return
	}

	// read file
	dat, err := os.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("%s cannot open. %w", filename, err)
		return
	}
	if bytes.Contains(dat, []byte("\r")) {
		err = fmt.Errorf("not support file `%s` with byte \\r", filename)
		return
	}
	lines := strings.Split(string(dat), "\n")

	var nl []string
	for i := range actual {
		if i == 0 {
			nl = lines[:actual[i].Start.Line-1]
		} else {
			nl = append(nl, lines[actual[i-1].End.Line:actual[i].Start.Line-1]...)
		}
		nl = append(nl, actual[i].String())
		if i == len(actual)-1 {
			nl = append(nl, lines[actual[i].End.Line:]...)
		}
	}

	body := strings.Join(nl, "\n")
	err = os.WriteFile(filename, []byte(body), 0666)
	if err != nil {
		return
	}

	if strings.HasSuffix(filename, ".go") {
		// simplify Go code by `gofmt`
		// error ignored, because it is not change the workflow
		_, _ = exec.Command("gofmt", "-s", "-w", filename).Output()
	}

	return
}

// Get snippets from `filename` and it is folder or file
func Get(filename string) (snippets []Snippet, err error) {
	const op = "Get"

	// snippet deferfunc
	defer func() {
		if err != nil {
			err = errors.Join(
				fmt.Errorf("%s. Filename : `%s`", op, filename),
				err,
			)
		}
	}()
	// end deferfunc

	//
	type (
		status bool
		record struct {
			s    status
			line int
			name string
		}
	)
	const (
		start status = true
		end   status = false
	)

	// check is file
	{
		var gofiles []string
		gofiles, err = paths(filename)
		if err != nil {
			return
		}
		if len(gofiles) == 0 {
			return
		}
		if 1 < len(gofiles) {
			for _, file := range gofiles {
				sn, errSn := Get(file)
				if errSn != nil {
					err = errors.Join(err, errSn)
				} else {
					snippets = append(snippets, sn...)
				}
			}
			return
		}
		filename = gofiles[0]
	}

	// read file
	dat, err := os.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("%s cannot open. %w", filename, err)
		return
	}
	if bytes.Contains(dat, []byte("\r")) {
		err = fmt.Errorf("not support file `%s` with byte \\r", filename)
		return
	}
	lines := strings.Split(string(dat), "\n")

	var records []record
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
					"%s:%d: snipet name cannot be with spaces: `%s`",
					filename, i+1,
					line,
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
		snippets = append(snippets, Snippet{
			Name:  records[i-1].name,
			Start: Position{Filename: filename, Line: records[i-1].line},
			End:   Position{Filename: filename, Line: records[i].line + 1},
			Code:  lines[records[i-1].line:records[i].line],
		})
	}

	// clean code
	for i := range snippets {
		cs := snippets[i].Code
		for i := range cs {
			cs[i] = strings.TrimSpace(cs[i])
		}
		snippets[i].Code = cs
	}

	return
}

// Compare snippers from expectFilenames and files/folders from actualFilename
func Compare(
	actualFilename, expectFilename string,
	diffOnly bool,
) (err error) {
	const op = "Compare"

	defer func() {
		if err != nil {
			err = errors.Join(
				fmt.Errorf("%s", op),
				fmt.Errorf("Error: expect `%s`, actual `%s`",
					expectFilename,
					actualFilename,
				),
				err,
			)
		}
	}()

	expect, err := Get(expectFilename)
	if err != nil {
		return
	}
	// check expect snippets
	for i := range expect {
		for j := range expect {
			if i <= j {
				continue
			}
			if strings.EqualFold(expect[i].Name, expect[j].Name) {
				err = errors.Join(err,
					fmt.Errorf("same snippets names `%s`", expect[i].Name),
				)
			}
		}
	}
	if err != nil {
		return
	}

	actual, err := Get(actualFilename)
	if err != nil {
		return
	}

	var differr []errorDiff // error of diff
	for _, act := range actual {
		found := false
		index := -1
		for ie, exp := range expect {
			if strings.EqualFold(act.Name, exp.Name) {
				found = true
				index = ie
				break
			}
		}
		if !found {
			err = errors.Join(err,
				fmt.Errorf("%s: cannot find snippet with name `%s`",
					act.Start,
					act.Name,
				))
			continue
		}
		ac := strings.Join(act.Code, "\n")
		ec := strings.Join(expect[index].Code, "\n")
		if ac != ec {
			de := errorDiff{Actual: act, Expect: expect[index]}
			differr = append(differr, de)
			err = errors.Join(err,
				de,
				compare.Diff([]byte(ac), []byte(ec)),
			)
		}
	}

	if diffOnly {
		return
	}
	err = nil // ignore errors

	var filenames []string
	for i := range differr {
		found := false
		for _, f := range filenames {
			if f == differr[i].Actual.Start.Filename {
				found = true
			}
		}
		if found {
			continue
		}
		filenames = append(filenames, differr[i].Actual.Start.Filename)
	}

	for _, filename := range filenames {
		err = errors.Join(err, update(filename, expect))
	}

	return
}

type errorDiff struct {
	Expect, Actual Snippet
}

func (e errorDiff) Error() string {
	return fmt.Sprintf(
		"%s: code is not same as in snippet `%s` in file: %s",
		e.Actual.Start,
		e.Expect.Name,
		e.Expect.Start,
	)
}

// ExpectSnippets is location of expect snippets
var ExpectSnippets = "./expect.snippets"

// Test and update snippets in all acceptable files in `folder` with subfolders.
// Location with expected snippets in file "ExpectSnippets"
//
// for update snippets run in console:
// UPDATE=true go test
// or see package `compare`
func Test(t interface {
	Errorf(format string, args ...any)
	Logf(format string, args ...any)
}, folder string) {
	diffOnly := true
	if os.Getenv(compare.Key) == compare.KeyValid {
		diffOnly = false
	}
	files, err := paths(folder)
	if err != nil {
		t.Errorf("cannot find expect files in `%s`: %v", folder, err)
		return
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	var errs []error
	for _, file := range files {
		err = Compare(file, ExpectSnippets, diffOnly)
		if err != nil {
			errs = append(errs, err)
		} else {
			errs = append(errs, nil)
		}
		fmt.Fprintf(w, "%s\t%v\n", file, err)
	}
	w.Flush()
	t.Logf("result:\n%s", buf.String())
	err = errors.Join(errs...)
	if err != nil {
		t.Errorf("%v", err)
	}
}

// SuffixFiles store list of acceptable fileformat
var SuffixFiles = []string{".go"}

// paths return only filenames with specific suffix
func paths(paths ...string) (files []string, err error) {
	const op = "Path"

	defer func() {
		if err != nil {
			err = errors.Join(
				fmt.Errorf("%s. %v", op, paths),
				err,
			)
		}
	}()

	for _, path := range paths {
		fileInfo, errF := os.Stat(path)
		if errF != nil {
			err = errors.Join(err, errF)
			return
		}
		if fileInfo.IsDir() {
			// is a directory
			errW := filepath.Walk(path,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}
					found := false
					for _, sf := range SuffixFiles {
						if strings.HasSuffix(path, sf) {
							found = true
						}
					}
					if !found {
						return nil
					}
					files = append(files, path)
					return nil
				})
			err = errors.Join(err, errW)
		} else {
			// is file
			files = append(files, path)
		}
	}
	return
}
