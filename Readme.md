## SpStat

A statistical package for calculating classical parametric statistics without
reading the entire dataset into memory, using single-pass statistical
approaches where possible.

## Introduction

Nearly all summary statistics can be calculated by reading through a set of
data one or a few times. Unfortunately, most statistical programs require that
an entire dataset be loaded into memory before any summary statistics are
calculated. This is often the faster approach, but it puts strict limits on the
size of datasets that can be summarized. Since many modern genomics datasets
are much too big to fit in main memory, this package calculates classical
statistics on streams of data.

## Library

The self-documented library of functions is available in Go language using the following import:

```go
import (
	"github.com/jgbaldwinbrown/spstat/pkg"
)
```

## Executables

### ttest

```
Usage of ttest:
  -bloodcol string
    	name of column listing control samples as "blood"
  -i string
    	input .gz file
  -testcol string
    	column to use for all test
  -v string
    	value column name
```

### ftest

```
Usage of ftest:
  -bloodcol string
    	name of column listing control samples as "blood"
  -i string
    	input .gz file
  -testcol string
    	column to use for all test
  -v string
    	value column name
```

### bloodnorm

```
Usage of bloodnorm:
  -i string
    	input .gz file
  -s string
    	column to subtract
  -v string
    	value column name
```

### others

More coming soon!
