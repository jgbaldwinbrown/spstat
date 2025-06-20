package spstat

import (
	"strconv"
	"fmt"
	"io"
	"encoding/csv"
	"github.com/jgbaldwinbrown/csvh"
)

// Given a set of strings representing columns in one row, calculate hits and
// counts from the hit column and count column, then return hits / counts
func AddAfracOne(line []string, hitcol, countcol int) (float64, error) {
	h := handle("CombineOne: %w")

	if len(line) <= hitcol { return 0, h(fmt.Errorf("line too short")) }
	hits, e := strconv.ParseFloat(line[hitcol], 64)
	if e != nil { return 0, h(e) }

	if len(line) <= countcol { return 0, h(fmt.Errorf("line too short")) }
	count, e := strconv.ParseFloat(line[countcol], 64)
	if e != nil { return 0, h(e) }

	return hits / count, nil
}

// Take a ReadCloserMaker and a hit column and a count column. For each line in
// the input stream, get hit and count from the hitcol and countcol, then
// append hit/count to the row, forming a new column
func AddAfrac(rcm ReadCloserMaker, w io.Writer, hitcol, countcol int) error {
	h := handle("AddAfrac: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	line, e := cr.Read()
	if e != nil { return h(e) }
	line = append(line, "value")
	e = cw.Write(line)
	if e != nil { return h(e) }

	for line, e = cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		combined, e := AddAfracOne(line, hitcol, countcol)
		if e != nil { continue }
		line = append(line, fmt.Sprint(combined))
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

// Given an input ReadCloserMaker that can be run multiple times, and provides
// a tab-separated table with a header line, identify the hit column and count
// column, then run AddAfrac on it and write the output to 'w'
func RunAddAfrac(rcm ReadCloserMaker, w io.Writer, hitcolname, countcolname string) error {
	h := handle("RunAddAfrac: %w")

	hitcol, e := ValCol(rcm, hitcolname)
	if e != nil { return h(e) }

	countcol, e := ValCol(rcm, countcolname)
	if e != nil { return h(e) }

	e = AddAfrac(rcm, w, hitcol, countcol)
	if e != nil { return h(e) }

	return nil
}
