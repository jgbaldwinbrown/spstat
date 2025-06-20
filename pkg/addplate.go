package spstat

import (
	"encoding/csv"
	"io"
	"os"
	"flag"
	"errors"
	"fmt"
)

// Read a tab-separated table from rcm, then associate columns 5 and 7 such
// that ids[l[5]] = l[7]. This finds the plate identity (l[7]) associated with
// each individual identity (l[5]).
func ReadIdents(rcm ReadCloserMaker) (map[string]string, error) {
	h := handle("ReadIdents: %w")
	ids := map[string]string{}

	r, e := rcm.NewReadCloser()
	if e != nil { return nil, h(e) }
	defer r.Close()
	cr := OpenCsv(r)

	cr.Read()
	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { return nil, h(e) }
		if len(l) < 8 { return nil, h(fmt.Errorf("len(l) %v < 8", len(l))) }

		ids[l[5]] = l[7]
	}
	return ids, nil
}

// Assign plate identities to a data table, assuming col contains the individual identities.
func AssignIdents(r io.Reader, w io.Writer, idents map[string]string, col int, header bool) error {
	h := handle("AssignIdents: %w")
	cr := OpenCsv(r)

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	if header {
		l, e := cr.Read()
		if e != nil { return h(e) }
		l = append(l, "plate")
		e = cw.Write(l)
		if e != nil { return h(e) }
	}

	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { return h(e) }
		if len(l) <= col { return h(fmt.Errorf("len(l) %v <= col %v", len(l), col)) }

		l = append(l, idents[l[col]])
		e = cw.Write(l)
		if e != nil { return h(e) }
	}
	return nil
}

// Wrapper around ReadIdents and AssignIdents
func AssignPlate(r io.Reader, w io.Writer, identrcm ReadCloserMaker, col int, header bool) error {
	h := handle("Assign Plate: %w")

	idents, e := ReadIdents(identrcm)
	if e != nil { return h(e) }

	e = AssignIdents(r, w, idents, col, header)
	if e != nil { return h(e) }

	return nil
}

// Run AssignPlate on the command line
func RunAddPlate() {
	headerp := flag.Bool("h", false, "Table contains a header")
	colp := flag.Int("c", -1, "column to use as indiv identities")
	identp := flag.String("i", "", "path to identities table")
	flag.Parse()

	if *colp == -1 {
		panic(errors.New("missing -c"))
	}
	if *identp == "" {
		panic(errors.New("missing -i"))
	}

	e := AssignPlate(os.Stdin, os.Stdout, MaybeGzPath(*identp), *colp, *headerp)
	if e != nil {
		panic(e)
	}
}

