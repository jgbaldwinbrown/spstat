package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"fmt"
	"strconv"
	"io"
	"encoding/csv"
)

func SubOne(line []string, valcol, tosubcol int) (float64, error) {
	h := handle("SubOne: %w")

	if len(line) <= valcol { return 0, h(fmt.Errorf("line too short")) }
	val, e := strconv.ParseFloat(line[valcol], 64)
	if e != nil { return 0, h(e) }

	if len(line) <= tosubcol { return 0, h(fmt.Errorf("line too short")) }
	tosub, e := strconv.ParseFloat(line[tosubcol], 64)
	if e != nil { return 0, h(e) }

	return val - tosub, nil
}

func ColSub(rcm ReadCloserMaker, w io.Writer, valcol, tosubcol int) error {
	h := handle("RunColSub: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	line, e := cr.Read()
	if e != nil { return h(e) }
	line = append(line, "sub")
	e = cw.Write(line)
	if e != nil { return h(e) }

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		subbed, e := SubOne(line, valcol, tosubcol)
		if e != nil { continue }
		line = append(line, fmt.Sprintf("%f", subbed))
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

func RunColSub(rcm ReadCloserMaker, w io.Writer, valcolname, tosubcolname string) error {
	h := handle("RunColSub: %w")

	valcol, e := ValCol(rcm, valcolname)
	if e != nil { return h(e) }

	tosubcol, e := ValCol(rcm, tosubcolname)
	if e != nil { return h(e) }

	e = ColSub(rcm, w, valcol, tosubcol)
	if e != nil { return h(e) }

	return nil
}
