package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"os"
	"encoding/csv"
	"flag"
	"bufio"
	"io"
	"fmt"
	"strconv"
)

// Predict a y value based on an x value and coefficients for the model y ~ x
func Predict(x, m, b float64) float64 {
	return (x * m) + b
}

// Predict y for based on an x column for the linear model y ~ x
func LinearModelPredict(rcm ReadCloserMaker, w io.Writer, indepcol int, m, b float64) (err error) {
	h := handle("LinearModelResiduals: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	line, e := cr.Read()
	if e != nil { return h(e) }
	line = append(line, "predicted")
	e = cw.Write(line)
	if e != nil { return h(e) }

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		if len(line) <= indepcol { continue }
		indep, e := strconv.ParseFloat(line[indepcol], 64)
		if e != nil { continue }

		pred := Predict(indep, m, b)
		line = append(line, fmt.Sprint(pred))
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

type modelParam struct {
	M float64
	B float64
}

// Write the coefficients of a linear model to a path
func WriteModelPath(path string, m, b float64) (err error) {
	return os.WriteFile(path, []byte(fmt.Sprintf("{M: %v, X: %v}\n", m, b)), 0644)
}

// Rescale a set of data by generating a linear model val ~ indep, then writing the predictions of that linear model
func RescaleData(rcm ReadCloserMaker, w io.Writer, modelOutPath string, valcolname, indepcolname string) error {
	h := handle("RescaleData: %w")

	valcol, e := ValCol(rcm, valcolname)
	if e != nil { return h(e) }

	indepcol, e := ValCol(rcm, indepcolname)
	if e != nil { return h(e) }

	m, b, e := LinearModel(rcm, valcol, indepcol)
	if e != nil { return h(e) }

	e = LinearModelPredict(rcm, w, indepcol, m, b)
	if e != nil { return h(e) }

	if modelOutPath != "" {
		e = WriteModelPath(modelOutPath, m, b)
		if e != nil { return h(e) }
	}

	return nil
}

// Like RescaleData, but for numbered columns
func RescaleDataResultFile(rcm ReadCloserMaker, w io.Writer, modelOutPath string, valcol, indepcol int) error {
	h := handle("RescaleDataResultFile: %w")

	m, b, e := LinearModel(rcm, valcol, indepcol)
	if e != nil { return h(e) }

	e = LinearModelPredict(rcm, w, indepcol, m, b)
	if e != nil { return h(e) }

	if modelOutPath != "" {
		e = WriteModelPath(modelOutPath, m, b)
		if e != nil { return h(e) }
	}

	return nil
}

type scaleEmpiricalFlags struct {
	Valcolname string
	Indepcolname string
	Path string
	ResultFile bool
	ModelOutPath string
}

// Scale data to match empirical results
func RunScaleEmpirical() {
	var f scaleEmpiricalFlags
	flag.StringVar(&f.Valcolname, "v", "", "Name of column with empirical, known values, i.e., 100% x representation for females")
	flag.StringVar(&f.Indepcolname, "i", "", "Name of column with estimated values")
	flag.StringVar(&f.Path, "p", "", "Input path")
	flag.BoolVar(&f.ResultFile, "r", false, "Interpret input file as results, not data")
	flag.StringVar(&f.ModelOutPath, "mo", "", "path to output model parameters")
	flag.Parse()

	h := handle("RunLinearModel: %w")

	stdout := bufio.NewWriter(os.Stdout)
	defer func() {
		if e := stdout.Flush(); e != nil {
			panic(h(e))
		}
	}()

	if !f.ResultFile {
		e := RescaleData(MaybeGzPath(f.Path), stdout, f.ModelOutPath, f.Valcolname, f.Indepcolname)
		if e != nil {
			panic(h(e))
		}
	} else {
		e := RescaleDataResultFile(MaybeGzPath(f.Path), stdout, f.ModelOutPath, 19, 12)
		if e != nil {
			panic(h(e))
		}
	}
}
