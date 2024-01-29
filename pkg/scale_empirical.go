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

func Predict(x, m, b float64) float64 {
	return (x * m) + b
}

func LinearModelPredict(path string, w io.Writer, indepcol int, m, b float64) (err error) {
	h := handle("LinearModelResiduals: %w")

	r, e := csvh.OpenMaybeGz(path)
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

func WriteModelPath(path string, m, b float64) (err error) {
	out := modelParam{m, b}
	w, e := os.Create(path)
	if e != nil {
		return e
	}
	defer func() {
		e := w.Close()
		if err == nil {
			err = e
		}
	}()

	_, e = fmt.Fprintf(w, "%#v\n", out)
	return e
}

func RescaleData(path string, w io.Writer, modelOutPath string, valcolname, indepcolname string) error {
	h := handle("RescaleData: %w")

	valcol, e := ValCol(path, valcolname)
	if e != nil { return h(e) }

	indepcol, e := ValCol(path, indepcolname)
	if e != nil { return h(e) }

	m, b, e := LinearModel(path, valcol, indepcol)
	if e != nil { return h(e) }

	e = LinearModelPredict(path, w, indepcol, m, b)
	if e != nil { return h(e) }

	if modelOutPath != "" {
		e = WriteModelPath(modelOutPath, m, b)
		if e != nil { return h(e) }
	}

	return nil
}

func RescaleDataResultFile(path string, w io.Writer, modelOutPath string, valcol, indepcol int) error {
	h := handle("RescaleDataResultFile: %w")

	m, b, e := LinearModel(path, valcol, indepcol)
	if e != nil { return h(e) }

	e = LinearModelPredict(path, w, indepcol, m, b)
	if e != nil { return h(e) }

	if modelOutPath != "" {
		e = WriteModelPath(modelOutPath, m, b)
		if e != nil { return h(e) }
	}

	return nil
}

type ScaleEmpiricalFlags struct {
	Valcolname string
	Indepcolname string
	Path string
	ResultFile bool
	ModelOutPath string
}

func RunScaleEmpirical() {
	var f ScaleEmpiricalFlags
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
		e := RescaleData(f.Path, stdout, f.ModelOutPath, f.Valcolname, f.Indepcolname)
		if e != nil {
			panic(h(e))
		}
	} else {
		e := RescaleDataResultFile(f.Path, stdout, f.ModelOutPath, 19, 12)
		if e != nil {
			panic(h(e))
		}
	}
}
