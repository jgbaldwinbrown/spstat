package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"math"
	"strconv"
	"encoding/csv"
	"io"
	"fmt"
)

func handle(form string) func(...any) error {
	return func(args ...any) error {
		return fmt.Errorf(form, args...)
	}
}

type NamedValSet struct {
	ColName string
	Idx int
	Names []string
	Sums map[string]float64
	Counts map[string]float64
}

func (s *NamedValSet) Mean(id string) float64 {
	return s.Sums[id] / s.Counts[id]
}

func OpenCsv(r io.Reader) (*csv.Reader) {
	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.FieldsPerRecord = -1
	cr.ReuseRecord = true
	cr.LazyQuotes = true
	return cr
}


// func Open(rcm ReadCloserMaker) (*csv.Reader, *gzip.Reader, *os.File, error) {
// 	h := handle("Open: %w")
// 
// 	f, e := os.Open(rcm)
// 	if e != nil { return nil, nil, nil, h(e) }
// 
// 	gr, e := gzip.NewReader(f)
// 	if e != nil {
// 		f.Close()
// 		return nil, nil, nil, h(e)
// 	}
// 
// 	cr := OpenCsv(gr)
// 
// 	return cr, gr, f, nil
// }

func NamedColsFunc(names []string) func(line []string, outbuf []int) ([]int, error) {
	h := handle("NamedColsFunc: %w")

	return func(line []string, outbuf []int) ([]int, error) {
		idxs := outbuf[:0]

		for _, name := range names {
			for i, col := range line {
				if name == col {
					idxs = append(idxs, i)
					break
				}
			}
		}

		if len(idxs) != len(names) {
			return nil, h(fmt.Errorf("len(idxs) %v != len(names) %v", len(idxs), len(names)))
		}

		return idxs, nil
	}
}

func NamedCols(rcm ReadCloserMaker, names []string) ([]int, error) {
	h := handle("NamedCols: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return nil, h(fmt.Errorf("path %v; %w", rcm, e)) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	line, e := cr.Read()
	if e != nil { return nil, h(e) }

	var idxs []int

	for _, name := range names {
		for i, col := range line {
			if name == col {
				idxs = append(idxs, i)
				break
			}
		}
	}

	if len(idxs) != len(names) {
		return nil, h(fmt.Errorf("len(idxs) %v != len(names) %v", len(idxs), len(names)))
	}

	return idxs, nil
}

func ValCol(rcm ReadCloserMaker, valname string) (int, error) {
	h := handle("ValCol: %w")

	cols, e := NamedCols(rcm, []string{valname})
	if e != nil { return 0, h(e) }

	return cols[0], nil
}

func ValColFunc(valname string) func(line []string, buf []int) (int, error) {
	colsf := NamedColsFunc([]string{valname})
	h := handle("ValColFunc: %w")

	return func(line []string, buf []int) (int, error) {
		cols, e := colsf(line, buf[:0])
		if e != nil { return 0, h(e) }

		return cols[0], nil
	}
}

func IdCols(rcm ReadCloserMaker, idnames []string) ([]int, error) {
	return NamedCols(rcm, idnames)
}

func IdColsFunc(idnames []string) func(line []string, buf []int) ([]int, error) {
	return NamedColsFunc(idnames)
}

func NewNamedValSet() *NamedValSet {
	s := new(NamedValSet)
	s.Sums = make(map[string]float64)
	s.Counts = make(map[string]float64)
	return s
}

func CalcMeans(rcm ReadCloserMaker, valcol int, idnames []string, idcols []int) ([]*NamedValSet, error) {
	h := handle("CalcMeans: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return nil, h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	sets := []*NamedValSet{}
	for i, name := range idnames {
		s := NewNamedValSet()
		s.ColName = name
		s.Idx = idcols[i]
		sets = append(sets, s)
	}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return sets, h(e) }

		if len(line) <= valcol { continue }
		val, e := strconv.ParseFloat(line[valcol], 64)
		if e != nil { continue }

		for _, set := range sets {
			if len(line) <= set.Idx { continue }
			set.Add(val, line[set.Idx])
		}
	}

	return sets, nil
}

func (s *NamedValSet) Add(val float64, id string) {
	if !math.IsNaN(val) {
		s.Sums[id] += val
		s.Counts[id]++
	}
}

func (s *NamedValSet) AddResid(val float64, line []string, means []*NamedValSet, id string) {
	resid := val
	for _, mean := range means {
		resid -= mean.Mean(line[mean.Idx])
	}
	s.Add(resid, id)
}

func CalcSerialMean(rcm ReadCloserMaker, valcol int, means []*NamedValSet, idname string, idcol int) (*NamedValSet, error) {
	h := handle("CalcSerialMean: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return nil, h(e) }
	defer r.Close()
	cr := csvh.CsvIn(r)

	s := NewNamedValSet()
	s.ColName = idname
	s.Idx = idcol

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return s, h(e) }

		if len(line) <= valcol { continue }
		val, e := strconv.ParseFloat(line[valcol], 64)
		if e != nil { continue }

		if len(line) <= s.Idx { continue }
		s.AddResid(val, line, means, line[s.Idx])
	}

	return s, nil
}

func CalcSerialMeans(rcm ReadCloserMaker, valcol int, idnames []string, idcols []int) ([]*NamedValSet, error) {
	h := handle("CalcSerialMeans: %w")

	var means []*NamedValSet
	for i, name := range idnames {
		mean, e := CalcSerialMean(rcm, valcol, means, name, idcols[i])
		if e != nil { return nil, h(e) }
		means = append(means, mean)
	}
	return means, nil
}

func NormOne(line []string, valcol int, means []*NamedValSet) (float64, error) {
	h := handle("NormOne: %w")

	if len(line) <= valcol { return 0, h(fmt.Errorf("line too short")) }
	val, e := strconv.ParseFloat(line[valcol], 64)
	if e != nil { return 0, h(e) }

	resid := val

	for _, mean := range means {
		if len(line) <= mean.Idx { return 0, h(fmt.Errorf("line too short")) }
		resid -= mean.Mean(line[mean.Idx])
	}
	return resid, nil
}

func Norm(rcm ReadCloserMaker, w io.Writer, valcol int, means []*NamedValSet) error {
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
	line = append(line, "norm")
	e = cw.Write(line)
	if e != nil { return h(e) }

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return h(e) }

		norm, e := NormOne(line, valcol, means)
		if e != nil { continue }
		line = append(line, fmt.Sprintf("%f", norm))
		e = cw.Write(line)
		if e != nil { continue }
	}

	return nil
}

func Run(rcm ReadCloserMaker, w io.Writer, valcolname string, idcolsnames []string) error {
	h := handle("Run: %w")

	valcol, e := ValCol(rcm, valcolname)
	if e != nil { return h(e) }

	idcols, e := IdCols(rcm, idcolsnames)
	if e != nil { return h(e) }

	means, e := CalcSerialMeans(rcm, valcol, idcolsnames, idcols)
	if e != nil { return h(e) }

	e = Norm(rcm, w, valcol, means)
	if e != nil { return h(e) }

	return nil
}
