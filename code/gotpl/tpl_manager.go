package gotpl

const TplManager = `
package {{.Package}}

import (
	"fmt"
	"io/ioutil"
	"path"
)

const (
{{- range .Sheets}}
	{{.AlignKey}} = "{{.Name}}"
{{- end}}
)

func init() {
	{{- range .Sheets}}
	add(&{{.Name}}Tbl{})
	{{- end}}
}

var gTableMap = make(map[string]ITable)

type ITable interface {
	Version() int
	Name() string
	File() string
	Load(data []byte) error
}

func add(tbl ITable) {
	gTableMap[tbl.Name()] = tbl
}

func GetTable(name string) ITable {
	return gTableMap[name]
}

func LoadFromFile(basedir string) error {
	errors := ""
	for _, t := range gTableMap {
		filename := path.Join(basedir, t.File())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := t.Load(data); err != nil {
			errors += err.Error()
		}
	}

	if len(errors) != 0 {
		return fmt.Errorf("%s", errors)
	}

	return nil
}
`
