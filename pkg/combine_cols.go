package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"strings"
	"fmt"
	"io"
	"encoding/csv"
)

func CombineOne(line []string, cols []int, sep string) (string, error) {
	h := handle("CombineOne: %w")

	tocombine := make([]string, 0, len(cols))

	for _, col := range cols {
		if len(line) <= col { return "", h(fmt.Errorf("line too short")) }
		tocombine = append(tocombine, strings.ReplaceAll(line[col], sep, "."))
	}

	return strings.Join(tocombine, sep), nil
}

func ColCombine(path string, w io.Writer, colsf func([]string, []int) ([]int, error), sep string) error {
	h := handle("ColCombine: %w")

	r, e := csvh.OpenMaybeGz(path)
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

func RunColCombine(path string, w io.Writer, colnames []string, sep string) error {
	h := handle("RunColCombine: %w")

	colsf := NamedColsFunc(colnames)

	if e := ColCombine(path, w, colsf, sep); e != nil {
		return h(e)
	}

	return nil
}
