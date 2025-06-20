package spstat

import (
	"io"
	"fmt"
	"gonum.org/v1/gonum/stat/distuv"
	"math"
)

// The core of an f test -- calculate sd1^2 / sd2^2
func FTestCore(sd1, sd2 float64) float64 {
	return (sd1 * sd1) / (sd2 * sd2)
}

// Find cases where the degrees of freedom are 0, nan, or inf
func BadDF(df float64) bool {
	return df <= 0 || math.IsNaN(df) || math.IsInf(df, 0)
}

// Get the p-value from the outcome of an f test using the distribution of f.
func FTestP(f, df1, df2 float64) float64 {
	if BadDF(df1) || BadDF(df2) {
		return math.NaN()
	}
	fdist := distuv.F{D1: df1, D2: df2}
	cdf := fdist.CDF(f)
	if cdf > 0.5 {
		return 2 * (1 - cdf)
	}
	return 2 * cdf
}

// Calculate the degrees of freedom in an F test.
func FTestDf(ts1 *TSummary, id1 string, ts2 *TSummary, id2 string) (df1, df2 float64) {
	return ts1.Counts[id1] - 1, ts2.Counts[id2] - 1
}

// Calculate an F test for one TTestSet. Print to w.
func FTest(w io.Writer, tsums []*TSummary, testset TTestSet) error {
	i1, name1 := TsumsSet(tsums, testset.Control)
	i2, name2 := TsumsSet(tsums, testset.Exp)

	mean1 := tsums[i1].Mean(name1)
	mean2 := tsums[i2].Mean(name2)

	sd1 := tsums[i1].Sd(name1)
	sd2 := tsums[i2].Sd(name2)

	count1 := tsums[i1].Counts[name1]
	count2 := tsums[i2].Counts[name2]

	df1, df2 := FTestDf(tsums[i1], name1, tsums[i2], name2)

	f := FTestCore(sd1, sd2)
	p := FTestP(f, df1, df2)

	fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", name1, name2, count1, count2, mean1, mean2, sd1, sd2, f, df1, df2, p)

	return nil
}

// Run FTest on each of testsets.
func FTests(w io.Writer, tsums []*TSummary, testsets []TTestSet) error {
	for _, tset := range testsets {
		e := FTest(w, tsums, tset)
		if e != nil {
			return fmt.Errorf("TTests: %w", e)
		}
	}
	return nil
}

// Run the whole FTest pipeline
func RunFTest(rcm ReadCloserMaker, w io.Writer, valcolname string, idcolsnames []string, controlsetidx, testsetidx int) error {
	h := handle("Run: %w")

	valcol, e := ValCol(rcm, valcolname)
	if e != nil { return h(e) }

	idcols, e := IdCols(rcm, idcolsnames)
	if e != nil { return h(e) }

	tsummaries, testsets, e := CalcTSummary(rcm, valcol, idcolsnames, idcols, controlsetidx, testsetidx)
	if e != nil { return h(e) }

	e = FTests(w, tsummaries, testsets)
	if e != nil { return h(e) }

	return nil
}

// Same as RunFTest, but with controlsetidx and testsetidx set to 0 and 1
func RunFullFTest(rcm ReadCloserMaker, w io.Writer, valcolname, bloodcolname, testcolname string) error {
	idcolsnames := []string{bloodcolname, testcolname}
	return RunFTest(rcm, w, valcolname, idcolsnames, 0, 1)
}
