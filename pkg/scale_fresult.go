package spstat

import (
	"github.com/jgbaldwinbrown/csvh"
	"os"
	"flag"
	"strings"
	"bufio"
	"io"
	"fmt"
	"strconv"
)

// A named linear model with an arbitrary number of coefficients, one per named probe
type Model struct {
	Name string
	Coeffs []float64
}

// Parse all coefficients associated with a named probe, usually two coefficients where coeff[0] is the intercept and coeff[1] is the slope
func ParseCoeffs(fields []string) ([]float64, error) {
	h := handle("ParseCoeffs: %w")
	var coeffs []float64

	for _, field := range fields {
		c, e := strconv.ParseFloat(field, 64)
		if e != nil { return nil, h(e) }
		coeffs = append(coeffs, c)
	}
	return coeffs, nil
}

// Parse a full model, usually from a line where l[0] == name, l[1] == intercept, and l[2] == slope
func ParseModel(line []string) (Model, error) {
	h := handle("ParseModel: %w")
	if len(line) < 1 {
		return Model{}, h(fmt.Errorf("len(line) < 1"))
	}

	coeffs, e := ParseCoeffs(line[1:])
	if e != nil { return Model{}, h(e) }

	return Model{line[0], coeffs}, nil
}

// Parse a model for each named probe using ParseModel
func ReadModel(r io.Reader) ([]Model, error) {
	h := handle("ReadModel: %w")
	cr := OpenCsv(r)
	models := []Model{}

	for line, e := cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return nil, h(e) }

		m, e := ParseModel(line)
		if e != nil { return nil, h(e) }
		models = append(models, m)
	}
	return models, nil
}

// Run ReadModel on rcm.NewReadCloser()
func ReadModelPath(rcm ReadCloserMaker) ([]Model, error) {
	h := handle("ReadModelPath: %w")

	r, e := rcm.NewReadCloser()
	if e != nil { return nil, h(e) }
	defer r.Close()

	return ReadModel(r)
}

// Generate a mapping of probe name to coefficient slice for each model
func MapProbeToCoeffs(models []Model) map[string][]float64 {
	m := map[string][]float64{}
	for _, model := range models {
		m[model.Name] = model.Coeffs
	}
	return m
}

// A chromosome and position in a genome
type ChrPos struct {
	Chr string
	Pos string
}

// A named chromosome and position in a genome
type ProbeChrPos struct {
	Probe string
	ChrPos
}

// Get probe name, chromosome, and position from a tab-separated table where line[0] is probe name, line[4] is chr, and line[5] is pos
func ReadProbeChrPos(rcm ReadCloserMaker) ([]ProbeChrPos, error) {
	h := handle("ReadProbeInfo: %w")

	r, e := rcm.NewReadCloser()
	defer r.Close()
	cr := csvh.CsvIn(r)
	cr.Comma = rune(',')
	cr.LazyQuotes = false

	pcps := []ProbeChrPos{}

	line, e := cr.Read()
	if e != nil { return nil, h(e) }

	for line, e = cr.Read(); e != io.EOF; line, e = cr.Read() {
		if e != nil { return nil, h(e) }

		if len(line) < 5 {
			return nil, h(fmt.Errorf("len(line) %v < 6", len(line)))
		}

		pcps = append(pcps, ProbeChrPos{line[0], ChrPos{line[4], line[5]}})
	}
	return pcps, nil
}

// Get a mapping from genome position to probe name
func MapChrPosToProbe(pcps []ProbeChrPos) map[ChrPos]string {
	m := map[ChrPos]string{}
	for _, pcp := range pcps {
		m[pcp.ChrPos] = pcp.Probe
	}
	return m
}

// Remove the remainder of val % step from val (good for finding window start points)
func Truncate(val int, step int) int {
	return (val / step) * step
}

// Find all of the named probes that fit in each tiled window
func MapChrPosWinToProbe(pcps []ProbeChrPos, winsize int) map[ChrPos][]string {
	m := map[ChrPos][]string{}
	for _, pcp := range pcps {
		pos, e := strconv.ParseInt(pcp.Pos, 0, 64)
		if e != nil { panic(e) }
		chrpos := ChrPos{pcp.Chr, fmt.Sprint(Truncate(int(pos), winsize))}
		m[chrpos] = append(m[chrpos], pcp.Probe)
	}
	return m
}

// Find all of the probes on a chromosome
func MapChrToProbes(pcps []ProbeChrPos) map[string][]string {
	m := map[string][]string{}
	for _, pcp := range pcps {
		m[pcp.Chr] = append(m[pcp.Chr], pcp.Probe)
	}
	return m
}

// Everything produced during an F test
type FTestResult struct {
	Name1 string
	Name2 string
	Count1 float64
	Count2 float64
	Mean1 float64
	Mean2 float64
	Sd1 float64
	Sd2 float64
	F float64
	Df1 float64
	Df2 float64
	P float64
}

// Read the results of an F test from a line in a table (the order of tab-separated
// entries is the same as the order of fields in FTestResult)
func ParseFTestResult(line string) (FTestResult, error) {
	r := FTestResult{}
	_, e := fmt.Sscanf(line, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
		&r.Name1, &r.Name2,
		&r.Count1, &r.Count2,
		&r.Mean1, &r.Mean2,
		&r.Sd1, &r.Sd2,
		&r.F,
		&r.Df1, &r.Df2,
		&r.P,
	)
	return r, e
}

// Read all F test results from a table using ParseFTestResult
func ReadFTestResults(r io.Reader) ([]FTestResult, error) {
	h := handle("ReadFTestResults: %w")
	results := []FTestResult{}
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)

	for s.Scan() {
		result, e := ParseFTestResult(s.Text())
		if e != nil { return nil, h(e) }
		results = append(results, result)
	}
	return results, nil
}

// The scaled standard deviation difference between two datasets
type ScaledFTest struct {
	Name1 string
	Name2 string
	ScaledSdDiff float64
}

// From a set of named coefficient lists, average all of map[ent][1] (the slopes)
func MeanSlope(probes []string, probeToCoeffMap map[string][]float64) (float64, error) {
	if len(probes) < 1 {
		return 0, fmt.Errorf("MeanSlope: No probes")
	}

	sum := 0.0
	count := 0.0
	for _, probe := range probes {
		coeffs, ok := probeToCoeffMap[probe]
		if !ok {
			// return 0, fmt.Errorf("MeanSlope: probe %v not in map", probe)
			continue
		}
		slope := coeffs[1]
		sum += slope
		count++
	}

	return sum / count, nil
}

// Convert an F test result into a scaled standard deviation difference, where
// the scale factor is (-slope / 2.0)
func ScaleSdDiff(ftest FTestResult, slope float64) float64 {
	diff := ftest.Sd2 - ftest.Sd1
	scaleFactor := -slope / 2.0
	return diff * scaleFactor
}

// Convert an F test result to the scaled mean difference -- (mean2 - mean1) * slope
func ScaleMeanDiff(ftest FTestResult, slope float64) float64 {
	diff := ftest.Mean2 - ftest.Mean1
	return diff * slope
}

// Generate a scaled F test based on the combination of the f test result, a
// map of chromosome to probe, and a map of probe to coefficient (on a per-chromosome basis)
func ScaleFTestPerChrom(ftest FTestResult, chrToProbeMap map[string][]string, probeToCoeffMap map[string][]float64) (ScaledFTest, error) {
	h := handle("ScaleFTestPerChrom: %w")

	namefields := strings.Split(ftest.Name2, "_")
	if len(namefields) < 2 {
		return ScaledFTest{}, h(fmt.Errorf("len(namefields) < 2"))
	}
	chr := namefields[1]

	probes, ok := chrToProbeMap[chr]
	if !ok {
		return ScaledFTest{}, h(fmt.Errorf("chr %v not in map", chr))
	}

	meanSlope, e := MeanSlope(probes, probeToCoeffMap)
	if e != nil { return ScaledFTest{}, h(e) }

	return ScaledFTest {
		ftest.Name1,
		ftest.Name2,
		ScaleSdDiff(ftest, meanSlope),
	}, nil
}

// Run ScaleFTestPerChrom for a set of F test results
func ScaleFTestsPerChrom(ftests []FTestResult, chrToProbeMap map[string][]string, probeToCoeffMap map[string][]float64) ([]ScaledFTest, error) {
	h := handle("ScaleFTestsPerChrom: %w")
	scaled := []ScaledFTest{}

	for _, ftest := range ftests {
		scaledone, e := ScaleFTestPerChrom(ftest, chrToProbeMap, probeToCoeffMap)
		if e != nil { return nil, h(e) }
		scaled = append(scaled, scaledone)
	}

	return scaled, nil
}

// Run ScaleFTest on each chromosome-position pair
func ScaleFTestPerChrPos(ftest FTestResult, chrPosToProbeMap map[ChrPos][]string, probeToCoeffMap map[string][]float64) (ScaledFTest, error) {
	h := handle("ScaleFTestPerChrom: %w")

	namefields := strings.Split(ftest.Name2, "_")
	if len(namefields) < 3 {
		return ScaledFTest{}, h(fmt.Errorf("len(namefields) < 2"))
	}
	chrPos := ChrPos{namefields[1], namefields[2]}

	probes, ok := chrPosToProbeMap[chrPos]
	if !ok {
		return ScaledFTest{}, h(fmt.Errorf("chrPos %v not in map", chrPos))
	}

	meanSlope, e := MeanSlope(probes, probeToCoeffMap)
	if e != nil { return ScaledFTest{}, h(e) }

	return ScaledFTest {
		ftest.Name1,
		ftest.Name2,
		ScaleSdDiff(ftest, meanSlope),
	}, nil
}

// Run ScaleFTestPerChrPos for each of a set of FTestResults
func ScaleFTestsPerChrPos(ftests []FTestResult, chrPosToProbeMap map[ChrPos][]string, probeToCoeffMap map[string][]float64) ([]ScaledFTest, error) {
	h := handle("ScaleFTestsPerChrom: %w")
	scaled := []ScaledFTest{}

	for _, ftest := range ftests {
		scaledone, e := ScaleFTestPerChrPos(ftest, chrPosToProbeMap, probeToCoeffMap)
		if e != nil { return nil, h(e) }
		scaled = append(scaled, scaledone)
	}

	return scaled, nil
}

// Write scaled F tests to a tab-separated table
func WriteScaled(w io.Writer, scaled []ScaledFTest) error {
	for _, s := range scaled {
		_, e := fmt.Fprintf(w, "%v\t%v\t%v\n", s.Name1, s.Name2, s.ScaledSdDiff)
		if e != nil { return fmt.Errorf("WriteScaled: %w", e) }
	}
	return nil
}

// Write both the basic F test and the scaled F test to a tab-separated table
func WriteScaled2(w io.Writer, ftests []FTestResult, scaled []ScaledFTest) error {
	h := handle("WriteScaled2: %w")
	if len(ftests) != len(scaled) {
		return h(fmt.Errorf("len(ftests) %v != len(scaled) %v", len(ftests), len(scaled)));
	}
	for i, s := range scaled {
		r := ftests[i]
		if r.Name1 != s.Name1 {
			return h(fmt.Errorf("r.Name1 %v != s.Name1 %v", r.Name1, s.Name1))
		}
		if r.Name2 != s.Name2 {
			return h(fmt.Errorf("r.Name2 %v != s.Name2 %v", r.Name2, s.Name2))
		}
		_, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			r.Name1, r.Name2,
			r.Count1, r.Count2,
			r.Mean1, r.Mean2,
			r.Sd1, r.Sd2,
			r.F,
			r.Df1, r.Df2,
			r.P,
			s.ScaledSdDiff,
		)
		if e != nil { return h(e) }
	}
	return nil
}

// Get the name to provide for the identity column. If unwindowed, it should be
// indiv_chrom_tissue_poswin. Otherwise, it should be "indiv_chrom_tissue"
func GetName2(winsize int) string {
	name2 := "indiv_chrom_tissue"
	if winsize != -1 {
		name2 = "indiv_chrom_tissue_poswin"
	}
	return name2
}

// If you're doing a T test, the stat is "t", otherwise "f"
func GetStat(ttest bool) string {
	stat := "f"
	if ttest {
		stat = "t"
	}
	return stat
}

// If you're doing a T test, scale means; otherwise, scale standard deviations
func GetDiff(ttest bool) string {
	diff := "scaled_sd_diff"
	if ttest {
		diff = "scaled_mean_diff"
	}
	return diff
}

// print the header for the F or T test results
func PrintHead(w io.Writer, ttest bool, outfmt string, winsize int) error {
	if outfmt == "2" {
		return PrintHeadOf2(w, ttest, winsize)
	}
	return PrintHeadOf1(w, ttest, winsize)
}

// Print the header using the outfmt == "1" setting
func PrintHeadOf1(w io.Writer, ttest bool, winsize int) error {
	name2 := GetName2(winsize)
	diff := GetDiff(ttest)

	_, e := fmt.Fprintf(w, "%v\t%v\t%v",
		"control_tissue", name2,
		diff,
	)
	return e
}

// Print the header using the outfmt == "2" setting
func PrintHeadOf2(w io.Writer, ttest bool, winsize int) error {
	name2 := GetName2(winsize)
	stat := GetStat(ttest)
	diff := GetDiff(ttest)

	_, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v",
		"control_tissue", name2,
		"count1", "count2",
		"mean1", "mean2",
		"sd1", "sd2",
		stat,
		"df1", "df2",
		"p",
		diff,
	)
	return e
}

// Run the entire F test scaling program on the command line
func RunScaleFTests() {
	pheadp := flag.Bool("ph", false, "Print header for output format")
	probepp := flag.String("p", "", "probe info path")
	modelpp := flag.String("m", "", "probe model path")
	winsizep := flag.Int("w", -1, "window size if using chrpos")
	ttestp := flag.Bool("t", false, "Do per-chromosome t-test instead of f-test")
	ofp := flag.String("of", "", "output format (currently supporting 1 or 2, default 1)")
	flag.Parse()

	if *pheadp {
		PrintHead(os.Stdout, *ttestp, *ofp, *winsizep)
		return
	}

	if *probepp == "" && !*ttestp {
		panic(fmt.Errorf("missing -p"))
	}
	if *modelpp == "" {
		panic(fmt.Errorf("missing -m"))
	}

	ftests, e := ReadFTestResults(os.Stdin)
	if e != nil { panic(e) }

	models, e := ReadModelPath(MaybeGzPath(*modelpp))
	if e != nil { panic(e) }

	var scaled []ScaledFTest

	if *ttestp {
		if len(models) != 1 {
			panic(fmt.Errorf("len(models) %v != 1)", len(models)))
		}
		scaled, e = ScaleTTestsPerChrom(ftests, models[0].Coeffs[1])
		if e != nil { panic(e) }
	} else if *winsizep == -1 {
		probeset, e := ReadProbeChrPos(MaybeGzPath(*probepp))
		if e != nil { panic(e) }

		chrtoprobe := MapChrToProbes(probeset)
		scaled, e = ScaleFTestsPerChrom(ftests, chrtoprobe, MapProbeToCoeffs(models))
		if e != nil { panic(e) }
	} else {
		probeset, e := ReadProbeChrPos(MaybeGzPath(*probepp))
		if e != nil { panic(e) }

		chrpostoprobe := MapChrPosWinToProbe(probeset, *winsizep)
		scaled, e = ScaleFTestsPerChrPos(ftests, chrpostoprobe, MapProbeToCoeffs(models))
		if e != nil { panic(e) }
	}

	stdout := bufio.NewWriter(os.Stdout)
	defer stdout.Flush()

	if *ofp == "2" {
		e = WriteScaled2(stdout, ftests, scaled)
	} else {
		e = WriteScaled(stdout, scaled)
	}
	if e != nil { panic(e) }
}

// Scale one FTestResult using ScaleMeanDiff
func ScaleTTestPerChrom(ftest FTestResult, slope float64) (ScaledFTest, error) {
	return ScaledFTest {
		ftest.Name1,
		ftest.Name2,
		ScaleMeanDiff(ftest, slope),
	}, nil
}

// Run ScaleTTestPerChrom on each of ftests
func ScaleTTestsPerChrom(ftests []FTestResult, slope float64) ([]ScaledFTest, error) {
	h := handle("ScaleFTestsPerChrom: %w")
	scaled := []ScaledFTest{}

	for _, ftest := range ftests {
		scaledone, e := ScaleTTestPerChrom(ftest, slope)
		if e != nil { return nil, h(e) }
		scaled = append(scaled, scaledone)
	}

	return scaled, nil
}

