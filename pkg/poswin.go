package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"strconv"
	"fmt"
	"io"
	"encoding/csv"
)

// Calculate the window that any given position belongs to, assuming tiled windows
func PosWinOne(line []string, col int, winsize int) (int, error) {
	h := handle("CombineOne: %w")

	if len(line) <= col { return 0, h(fmt.Errorf("line too short")) }
	p, e := strconv.ParseInt(line[col], 0, 64)
	if e != nil { return 0, h(e) }

	return (int(p) / winsize) * winsize, nil
}

// CalculatePosWinOne for all lines in rcm
func PosWin(rcm ReadCloserMaker, w io.Writer, colf func([]string, []int) (int, error), winsize int) error {
	h := handle("PosWin: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	line, e := cr.Read()
	if e != nil { return h(e) }
	col, e := colf(line, []int{})
	if e != nil { return h(e) }

	line = append(line, "poswin")
	e = cw.Write(line)
	if e != nil { return h(e) }

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		pw, e := PosWinOne(line, col, winsize)
		if e != nil { continue }
		line = append(line, fmt.Sprint(pw))
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

// Run PosWin with a named column
func RunPosWin(rcm ReadCloserMaker, w io.Writer, colname string, winsize int) error {
	h := handle("RunColCombine: %w")

	colf := ValColFunc(colname)

	if e := PosWin(rcm, w, colf, winsize); e != nil {
		return h(e)
	}

	return nil
}
