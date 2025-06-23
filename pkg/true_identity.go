package spstat

import (
	"encoding/csv"
	"fmt"
	"flag"
	"os"
	"io"
	"strconv"
)

// Identifies a sample by its position on a 96-well plate
type ExpSexId struct {
	PlateId string
	Let string
	Num int
}

// Identifies a sample according the experiment done, the sex, and the individual ID
type ExpSexEntry struct {
	Exp string
	Sex string
	Indiv string
}

// A mapping between plate position and experiment, sex, and individual
type ExpSexSet struct {
	M map[ExpSexId]ExpSexEntry
}

// Parse a row of a table of positions and experiment, sex, and individual info
func ParseExpSexLine(line []string) (ExpSexId, ExpSexEntry, error) {
	h := handle("ParseExpSexLine: %w")

	if len(line) < 6 {
		return ExpSexId{}, ExpSexEntry{}, h(fmt.Errorf("len(line) %v < 5", len(line)))
	}

	num, e := strconv.Atoi(line[2])
	if e != nil {
		return ExpSexId{}, ExpSexEntry{}, h(e)
	}

	return ExpSexId{line[0], line[1], num}, ExpSexEntry{line[3], line[4], line[5]}, nil
}

// Parse a full table of positions and experiment, sex, and individual
func GetExperimentSexInfo(rcm ReadCloserMaker) (*ExpSexSet, error) {
	h := handle("GetExperimentSexInfo: %w")

	s := &ExpSexSet{M: map[ExpSexId]ExpSexEntry{}}

	r, e := rcm.NewReadCloser()
	if e != nil { return s, h(e) }
	defer r.Close()

	cr := OpenCsv(r)

	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { return s, h(e) }

		id, entry, e := ParseExpSexLine(l)
		if e != nil { return s, h(e) }
		s.M[id] = entry
	}
	return s, nil
}

// Parse just experiment, sex, and individual from the master data table of sample identities
func (s *ExpSexSet) GetExpSex(line []string) (exp, sex, indiv string, err error) {
	h := handle("GetExpSex: %w")

	if len(line) < 19 {
		return "", "", "", h(fmt.Errorf("len(line) %v < 14", len(line)))
	}

	num, e := strconv.Atoi(line[18])
	if e != nil { return "", "", "", h(e) }

	es, ok := s.M[ExpSexId{line[16], line[17], num}]
	if !ok {
		return line[8], line[9], line[11], nil
	}
	return es.Exp, es.Sex, es.Indiv, nil
}

// Extract the mapping from plate position to identity from
// experimentSexInfoPath, then use that mapping to convert the master data
// table in r to hold the true identities, rather than the original, mistaken
// identities
func TrueIdentity(r io.Reader, w io.Writer, experimentSexInfoPath string) error {
	h := handle("RunTrueIdentity: %w")

	// expSexInfo, e := GetExperimentSexInfo(experimentSexInfoPath)
	// if e != nil { return h(e) }

	cr := OpenCsv(r)

	cw := csv.NewWriter(w)
	defer cw.Flush()
	cw.Comma = rune('\t')

	l, e := cr.Read()
	if e != nil { return h(e) }
	e = cw.Write([]string {
		"probeset_id",
		"Chromosome",
		"Position",
		"Offset",
		"Big_Offset",
		"value",
		"let",
		"num",
		"experiment",
		"sex",
		"tissue",
		"indivs",
		"unique_id",
		"plate_id",
		"name_prefix",
		"spoofed_tissue",
	})
	if e != nil { return h(e) }

	outbuf := []string{}
	for l, e = cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { return h(e) }

		// exp, sex, indiv, e := expSexInfo.GetExpSex(l)
		outbuf = append(outbuf[:0],
			l[0],
			l[1],
			l[2],
			l[3],
			l[4],
			l[5],
			l[17],
			l[18],
			l[8],
			l[9],
			l[10],
			l[11],
			l[12],
			l[16],
			l[21],
			l[20],
		)
		e = cw.Write(outbuf)
		if e != nil { return h(e) }
	}

	return nil
}

// Run TrueIdentity on the command line
func RunTrueIdentities() {
	ipathp := flag.String("si", "", "path to sample exp sex identities, columns should be plate id num, plate letter, plate col num, exp, sex, indiv")
	flag.Parse()
	e := TrueIdentity(os.Stdin, os.Stdout, *ipathp)
	if e != nil { panic(e) }
}
