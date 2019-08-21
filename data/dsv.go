package data

import (
	"bytes"
	"encoding/csv"
	"gentable/base"
	"io/ioutil"
	"os"
)

const (
	kTypeCsv = "csv"
	kTypeTsv = "tsv"
	kTypePsv = "psv"
)

func newDsvBuilder(dtype string) Builder {
	var comma rune
	switch dtype {
	case kTypeCsv:
		comma = ','
	case kTypeTsv:
		comma = '\t'
	case kTypePsv:
		comma = '|'
	}

	return &DsvBuilder{comma: comma, dtype: dtype}
}

type DsvBuilder struct {
	comma rune
	dtype string
}

func (b *DsvBuilder) Type() string {
	return b.dtype
}

func (b *DsvBuilder) Build(sheet *base.Sheet) error {
	builder := bytes.Buffer{}
	w := csv.NewWriter(&builder)
	if err := w.Write(sheet.Names); err != nil {
		return err
	}
	w.Comma = b.comma

	if err := w.Write(sheet.GetTypes()); err != nil {
		return err
	}

	if err := w.Write(sheet.Notes); err != nil {
		return err
	}

	if err := w.WriteAll(sheet.Values); err != nil {
		return err
	}

	data := builder.Bytes()
	output := outFile(b.dtype, sheet.ID)
	return ioutil.WriteFile(output, data, os.ModePerm)
}
