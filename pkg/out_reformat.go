package spstat

import (
	"io"
	"encoding/csv"
	"os"
	"strings"
)

// Reformat output so that l[1] is separated by '_' and appended to the line
func Reformat(r io.Reader, w io.Writer) error {
	h := handle("Reformat: %w")

	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	cr.Comma = rune('\t')

	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer cw.Flush()

	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil { return h(e) }

		split := strings.Split(l[1], "_")
		l = append(l, split...)
		e = cw.Write(l)

		if e != nil { return h(e) }
	}

	return nil
}

// Run Reformat on the command line
func RunReformat() {
	e := Reformat(os.Stdin, os.Stdout)
	if e != nil { panic(e) }
}
