package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"fmt"
	"io"
)

// Generate a TSummary for each specified column, with transformFunc used to
// convert the strings in the column into expectations.
func CalcFullColTSummaryTransform(rcm ReadCloserMaker, cols []int, transformFunc func(string) float64) ([]*TSummary, error) {
	h := handle("CalcTSummaryTransform: %w")

	var tsums []*TSummary
	for _, col := range cols {
		tsum := NewTSummary()
		tsum.Idx = col
		tsums = append(tsums, tsum)
	}

	r, e := rcm.NewReadCloser()
	if e != nil { return nil, h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return tsums, h(e) }


		for i, tsum := range tsums {
			valcol := cols[i]
			if len(line) <= valcol { continue }
			val := transformFunc(line[valcol])

			if len(line) <= tsum.Idx { continue }
			tsum.Add(val, "")
		}
	}

	return tsums, nil
}

// Calculate a linear model fitting valcol ~ indepcol, and using transformFunc on valcol
func LinearModelTransform(rcm ReadCloserMaker, valcol, indepcol int, transformFunc(func(string)float64)) (m, b float64, err error) {
	h := handle("LinearModel: %w")

	valtsums, e := CalcFullColTSummaryTransform(rcm, []int{valcol}, transformFunc)
	indeptsums, e := CalcFullColTSummary(rcm, []int{indepcol})
	if e != nil { return 0, 0, h(e) }
	vmean := valtsums[0].Mean("")
	imean := indeptsums[0].Mean("")

	m, b, e = LinearModelCore(rcm, valcol, indepcol, vmean, imean)
	if e != nil { return 0, 0, h(e) }
	return m, b, nil
}

// Presume expected chromosome coverage of 1.0, unless chr is a sex chromosome matching [XxYy]
func ChrToExpectation(chr string) float64 {
	chrcov := 1.0
	if chr == "X" || chr == "Y" || chr == "x" || chr == "y" {
		chrcov = 0.5
	}
	return chrcov
}

// Run the whole linear model coverage pipeline
func RunLinearModelCoverage(rcm ReadCloserMaker, w io.Writer, valcolname, indepcolname string) error {
	h := handle("RunLinearModel: %w")

	valcol, e := ValCol(rcm, valcolname)
	if e != nil { return h(e) }

	indepcol, e := ValCol(rcm, indepcolname)
	if e != nil { return h(e) }

	m, b, e := LinearModelTransform(rcm, valcol, indepcol, ChrToExpectation)
	if e != nil { return h(e) }

	fmt.Fprintf(w, "allchromtotals\t%v\t%v\n", b, m)

	return nil
}
