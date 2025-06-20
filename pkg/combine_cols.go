package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"strings"
	"fmt"
	"io"
	"encoding/csv"
)

// Extract the specified columns from line, then join them using sep. If the
// column entry contains sep, replace all instances of sep with '.'.
func CombineOne(line []string, cols []int, sep string) (string, error) {
	h := handle("CombineOne: %w")

	tocombine := make([]string, 0, len(cols))

	for _, col := range cols {
		if len(line) <= col { return "", h(fmt.Errorf("line too short")) }
		tocombine = append(tocombine, strings.ReplaceAll(line[col], sep, "."))
	}

	return strings.Join(tocombine, sep), nil
}

// Take a tab-separated table and a function that reads the header and finds
// the correct columns to combine. Combine those columns with sep, then append to the existing line.
func ColCombine(rcm ReadCloserMaker, w io.Writer, colsf func([]string, []int) ([]int, error), sep string) error {
	h := handle("ColCombine: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	line, e := cr.Read()
	if e == io.EOF {
		return nil
	}
	cols, e := colsf(line, []int{})
	if e != nil {
		return h(e)
	}

	for ; e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		combined, e := CombineOne(line, cols, sep)
		if e != nil { continue }
		line = append(line, combined)
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

// Run ColCombine on the named columns
func RunColCombine(rcm ReadCloserMaker, w io.Writer, colnames []string, sep string) error {
	h := handle("RunColCombine: %w")

	colsf := NamedColsFunc(colnames)

	if e := ColCombine(rcm, w, colsf, sep); e != nil {
		return h(e)
	}

	return nil
}
