package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"os"
	"flag"
	"strconv"
	"encoding/csv"
	"io"
	"fmt"
)

func NormVarOne(line []string, valcol int, tsum *TSummary) (float64, error) {
	h := handle("NormOne: %w")

	if len(line) <= valcol { return 0, h(fmt.Errorf("line too short")) }
	val, e := strconv.ParseFloat(line[valcol], 64)
	if e != nil { return 0, h(e) }

	resid := (val - tsum.Mean(line[tsum.Idx])) / tsum.Sd(line[tsum.Idx])

	return resid, nil
}

func NormVar(rcm ReadCloserMaker, w io.Writer, valcol int, tsum *TSummary) error {
	h := handle("Norm: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	line, e := cr.Read()
	if e != nil { return h(e) }
	line = append(line, "normvar")
	e = cw.Write(line)
	if e != nil { return h(e) }

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		norm, e := NormVarOne(line, valcol, tsum)
		if e != nil { continue }
		line = append(line, fmt.Sprintf("%f", norm))
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

func RunNormVar(rcm ReadCloserMaker, w io.Writer, valcolname string, idcolname string) error {
	h := handle("RunNormVar: Step: %v; %w")

	valcol, e := ValCol(rcm, valcolname)
	if e != nil { return h("valcol", e) }

	idcol, e := ValCol(rcm, idcolname)
	if e != nil { return h("idcol", e) }

	tsums, _, e := CalcTSummary(rcm, valcol, []string{idcolname}, []int{idcol}, 0, 0)
	if e != nil { return h("tsums", e) }
	tsum := tsums[0]

	// fmt.Println("means and vars:")
	// for name, _ := range tsum.Sums {
	// 	fmt.Printf("name: %v; mean: %v; var: %v; sd: %v\n", name, tsum.Mean(name), tsum.Var(name), tsum.Sd(name))
	// }

	e = NormVar(rcm, w, valcol, tsum)
	if e != nil { return h("normvar", e) }

	return nil
}

func RunFullNormVar() {
	inpp := flag.String("i", "", "input .gz file")
	valcolp := flag.String("v", "", "value column name")
	idcolp := flag.String("id", "", "id column name")
	flag.Parse()
	if *inpp == "" {
		panic(fmt.Errorf("missing -i"))
	}
	if *valcolp == "" {
		panic(fmt.Errorf("missing -v"))
	}
	if *idcolp == "" {
		panic(fmt.Errorf("missing -id"))
	}

	e := RunNormVar(MaybeGzPath(*inpp), os.Stdout, *valcolp, *idcolp)
	if e != nil { panic(e) }
}
