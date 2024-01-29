package spstat

import (
	"log"
	"errors"
	"flag"
	"os"
	"encoding/csv"
	"bufio"
	"io"
	"github.com/jgbaldwinbrown/csvh"
)

func FExpectation(sex, experiment, tissue, chrom string) string {
	// if sex == "female" {
	// 	return "1.0"
	// }
	switch experiment {
	case "control": return "0.5"
	case "control_0": return "0.5"
	case "control_1": return "0.51"
	case "control_2": return "0.52"
	case "control_4": return "0.54"
	case "control_8": return "0.58"
	}
	if tissue == "blood" && sex != "female" {
		return "0.5"
	}
	return ""
}

func TExpectation(sex, experiment, tissue, chrom string) string {
	if sex == "female" && chrom == "X" {
		return "1.0"
	}
	if tissue == "blood" && chrom == "X" {
		return "0.5"
	}
	return ""
}

func Expectation(t bool, sex, experiment, tissue, chrom string) string {
	if t {
		return TExpectation(sex, experiment, tissue, chrom)
	}
	return FExpectation(sex, experiment, tissue, chrom)
}

func LinearModelAppendExpectation(rcm ReadCloserMaker, w io.Writer, t bool, sexcol, experimentcol, tissuecol, chromcol int, header bool) (err error) {
	h := handle("LinearModelAppendExpectations: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return h(e) }
	cr := csvh.CsvIn(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	if header {
		line, e := cr.Read()
		if e != nil { return h(e) }
		line = append(line, "expected")
		e = cw.Write(line)
		if e != nil { return h(e) }
	}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		if len(line) <= sexcol { continue }
		sex := line[sexcol]

		if len(line) <= experimentcol { continue }
		experiment := line[experimentcol]

		if len(line) <= tissuecol { continue }
		tissue := line[tissuecol]

		if len(line) <= chromcol { continue }
		chrom := line[chromcol]

		expect := Expectation(t, sex, experiment, tissue, chrom)
		line = append(line, expect)
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

func FullAppendExpectation(rcm ReadCloserMaker, w io.Writer, t bool, sexcolname, experimentcolname, tissuecolname, chromcolname string) error {
	h := handle("RunLinearModel: %w")

	sexcol, e := ValCol(rcm, sexcolname)
	if e != nil { return h(e) }

	experimentcol, e := ValCol(rcm, experimentcolname)
	if e != nil { return h(e) }

	tissuecol, e := ValCol(rcm, tissuecolname)
	if e != nil { return h(e) }

	chromcol, e := ValCol(rcm, chromcolname)
	if e != nil { return h(e) }

	e = LinearModelAppendExpectation(rcm, w, t, sexcol, experimentcol, tissuecol, chromcol, true)
	if e != nil { return h(e) }

	return nil
}

type AppendExpectationFlags struct {
	Path string
	ResultFile bool
	T bool
}

func RunAppendExpectation() {
	h := handle("RunAppendExpectation: %w")

	var f AppendExpectationFlags
	flag.StringVar(&f.Path, "i", "", "inpath")
	flag.BoolVar(&f.ResultFile, "r", false, "interpret input file as a results file, not a data file")
	flag.BoolVar(&f.T, "t", false, "Append t test expectations")
	flag.Parse()
	if f.Path == "" {
		log.Fatal(errors.New("missing -i"))
	}

	stdout := bufio.NewWriter(os.Stdout)
	defer func() {
		if e := stdout.Flush(); e != nil {
			panic(h(e))
		}
	}()

	if !f.ResultFile {
		e := FullAppendExpectation(MaybeGzPath(f.Path), stdout, f.T, "sex", "experiment", "tissue", "chrom")
		if e != nil {
			log.Fatal(h(e))
		}
	} else {
		e := LinearModelAppendExpectation(MaybeGzPath(f.Path), stdout, f.T, 18, 17, 3, -1, false)
		if e != nil {
			log.Fatal(h(e))
		}
	}
}
