package main

import (
	"fmt"
	"github.com/jgbaldwinbrown/fasttsv"
	"io"
	"os"
	"errors"
	"flag"
	"strconv"
)

type col_indices struct {
	Sample int
	Gc int
}

type flag_set struct {
	Sample string
	GcCol string
	GcPath string
}

type GcData struct {
	AsStr string
	AsFlo float64
	Na bool
}

func get_flags() (f flag_set) {
	flag.StringVar(&f.Sample, "s", "sample", "Name of column containing sample names.")
	flag.StringVar(&f.GcCol, "g", "gc_col", "Name of column containing per-chromosome GC fraction.")
	flag.StringVar(&f.GcPath, "G", "gc_path", "Path to column-separated file containing correlation with GC for each sample.")
	flag.Parse()
	return f
}

func get_cols(header []string, f flag_set) (cols col_indices, err error) {
	var sample_ok, gc_ok bool
	for i, v := range header {
		if v == f.Sample {
			cols.Sample = i
			sample_ok = true
		}
		if v == f.GcCol {
			cols.Gc = i
			gc_ok = true
		}
	}
	if (!sample_ok) || (!gc_ok) {
		err = errors.New("Error: missing named columns.")
	}
	return cols, err
}

func GetGcData(r io.Reader) (map[string]GcData, error) {
	out := make(map[string]GcData)
	var err error = nil
	s := fasttsv.NewScanner(r)
	s.Scan()

	for s.Scan() {
		var gc_data GcData
		gc_data.AsStr = s.Line()[1]
		gc_data.AsFlo, err = strconv.ParseFloat(s.Line()[1], 64)
		if err != nil {
			gc_data.Na = true
		}
		out[s.Line()[0]] = gc_data
	}
	return out, err
}

func assoc_and_print_all(r io.Reader, w io.Writer, gc_data map[string]GcData, f flag_set) error {
	var err error = nil
	s := fasttsv.NewScanner(r)
	s.Scan()
	header := make([]string, len(s.Line()))
	copy(header, s.Line())
	cols, err := get_cols(header, f)
	if err != nil { return err }
	header = append(header, "sample_gc_corr", "per_chrom_gc_index")

	W := fasttsv.NewWriter(w)
	defer W.Flush()
	W.Write(header)

	outline := make([]string, 0, 0)

	for s.Scan() {
		gc_corr, ok := gc_data[s.Line()[cols.Sample]]
		if !ok {
			gc_corr.AsStr = "NA"
			gc_corr.Na = true
		}

		outline = outline[:0]
		outline = append(outline, s.Line()...)
		outline = append(outline, gc_corr.AsStr)

		chrom_gc_frac, err := strconv.ParseFloat(s.Line()[cols.Gc], 64)
		if (err != nil) || gc_corr.Na {
			outline = append(outline, "NA")
		} else {
			per_chrom_gc_index := chrom_gc_frac * gc_corr.AsFlo
			outline = append(outline, fmt.Sprintf("%v", per_chrom_gc_index))
		}
		W.Write(outline)
	}

	return nil
}

func main() {
	flags := get_flags()

	gc_conn, err := os.Open(flags.GcPath)
	if err != nil {panic(err) }
	defer gc_conn.Close()

	gc_values, err := GetGcData(gc_conn)
	if err != nil { panic(err) }

	assoc_and_print_all(os.Stdin, os.Stdout, gc_values, flags)
}