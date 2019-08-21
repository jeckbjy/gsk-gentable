package input

import (
	"errors"
	"fmt"
	"gentable/base"
	"os"

	"github.com/tealeg/xlsx"
)

type XlsxBuilder struct {
}

func (b *XlsxBuilder) Type() string {
	return TypeXlsx
}

func (b *XlsxBuilder) Load(opts *Options) ([]LoadFunc, error) {
	results := make([]LoadFunc, 0)

	file, err := os.Open(opts.Source)
	if err != nil {
		return nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	checker := Checker{}
	checker.Init(opts.Sheets)

	if !fi.IsDir() {
		// load file
		xfile, err := xlsx.OpenFile(opts.Source)
		if err != nil {
			return nil, err
		}

		for _, sheet := range xfile.Sheets {
			if !checker.Check(sheet.Name) {
				tmp := sheet
				results = append(results, func() (*base.Sheet, error) {
					return b.parse(tmp)
				})
			}
		}
	} else {
		// load from directory
	}

	return results, nil
}

func (b *XlsxBuilder) parse(sheet *xlsx.Sheet) (*base.Sheet, error) {
	result := &base.Sheet{}
	result.Init(sheet.Name)

	if len(sheet.Rows) < 2 {
		return nil, errors.New("no enough data")
	}

	column := len(sheet.Rows[0].Cells)
	for i, row := range sheet.Rows {
		cells := make([]string, 0, column)
		for _, cell := range row.Cells {
			data := fmt.Sprint(cell)
			cells = append(cells, data)
		}
		err := result.Append(i, cells)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
