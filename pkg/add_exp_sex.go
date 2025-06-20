package spstat

import (
	"strings"
	"encoding/csv"
	"fmt"
	"flag"
	"os"
	"io"
	"strconv"
)

// For all rows of an info table, extract a key from keycol. Then, make a map
// with that key as the map key and that line as the map value
func GetInfoMap(rcm ReadCloserMaker, keycol int) (map[string][]string, error) {
	h := handle("GetInfoMap: %w")

	m := map[string][]string{}

	r, e := rcm.NewReadCloser()
	if e != nil { return nil, h(e) }
	defer r.Close()

	cr := OpenCsv(r)
	cr.ReuseRecord = false

	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { return nil, h(e) }
		if len(l) <= keycol {
			return nil, h(fmt.Errorf("len(l) %v < keycol %v; l %v", len(l), keycol, l))
		}

		key := l[keycol]
		m[key] = l
	}
	return m, nil
}

// Get a key from line[keycol], then find that key in idmap and extract all of
// the values at the indices of valcols.
func GetColsToAppend(line []string, idmap map[string][]string, keycol int, valcols []int) ([]string, error) {
	h := handle("GetColsToAppend: %w")

	if len(line) <= keycol {
		return nil, h(fmt.Errorf("len(line) %v < keycol %v", len(line), keycol))
	}


	idline, ok := idmap[line[keycol]]
	if !ok {
		return make([]string, len(valcols)), nil
	}

	var out []string
	for _, valcol := range valcols {
		if len(idline) <= valcol {
			return nil, h(fmt.Errorf("len(idline) %v < valcol %v", len(idline), valcol))
		}
		out = append(out, idline[valcol])
	}

	return out, nil
}

// Append all the requested id columns to the data read in r, using line[keycol] to identify the right info in infoMap.
func AddIdCols(r io.Reader, w io.Writer, keycol int, idcolnames []string, idcols []int, infoMap map[string][]string) error {
	h := handle("RunTrueIdentity: %w")

	cr := OpenCsv(r)

	cw := csv.NewWriter(w)
	defer cw.Flush()
	cw.Comma = rune('\t')

	l, e := cr.Read()
	if e != nil { return h(e) }
	l = append(l, idcolnames...)
	e = cw.Write(l)
	if e != nil { return h(e) }

	for l, e = cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { return h(e) }

		newcols, e := GetColsToAppend(l, infoMap, keycol, idcols)
		if e != nil { return h(e) }
		l = append(l, newcols...)

		e = cw.Write(l)
		if e != nil { return h(e) }
	}

	return nil
}

// Get comma-separated integers from a string
func ParseValCols(str string) ([]int, error) {
	spl := strings.Split(str, ",")
	out := make([]int, 0, len(spl))
	for _, vcstr := range spl {
		vc, e := strconv.Atoi(vcstr)
		if e != nil {
			return nil, fmt.Errorf("ParseValCols: %w", e)
		}
		out = append(out, vc)
	}
	return out, nil
}

// Run AddIdCols on the command line
func RunAddIdCols() {
	ipathp := flag.String("si", "", "path to sample ID file")
	idKeyColp := flag.Int("idkey", -1, "Key column in ID file")
	keyColp := flag.Int("key", -1, "Key column in data file")
	idValColsp := flag.String("valcols", "", "comma-separated columns in ID file to add")
	idValNamesp := flag.String("valnames", "", "comma-separated names of columns to add")
	flag.Parse()

	if *ipathp == "" {
		panic(fmt.Errorf("missing -si"))
	}
	if *idKeyColp == -1 {
		panic(fmt.Errorf("missing -idkey"))
	}
	if *keyColp == -1 {
		panic(fmt.Errorf("missing -key"))
	}
	if *idValNamesp == "" {
		panic(fmt.Errorf("missing -valnames"))
	}
	if *idValColsp == "" {
		panic(fmt.Errorf("missing -valcols"))
	}

	valcols, e := ParseValCols(*idValColsp)
	if e != nil {
		panic(e)
	}

	infoMap, e := GetInfoMap(MaybeGzPath(*ipathp), *idKeyColp)
	if e != nil {
		panic(e)
	}

	e = AddIdCols(os.Stdin, os.Stdout, *keyColp, strings.Split(*idValNamesp, ","), valcols, infoMap)
	if e != nil { panic(e) }
}
