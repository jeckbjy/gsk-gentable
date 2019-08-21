package code

const goTableTpl = `
{{ $ClassTbl := printf "%sTbl" .Class -}}
package {{.Package}}

type {{.Class}} struct {
	{{- range .Fields}}
	{{.Name}} {{.Type}} {{.Tags}}
	{{- end}}
}

type {{$ClassTbl}} struct {
	version int
	list []*{{.Class}}
	dict map[{{.IDType}}]*{{.Class}}
}

func (t *{{$ClassTbl}}) At(index int) *{{.Class}} {
	return t.list[index]
}

func (t *{{$ClassTbl}}) Get(id {{.IDType}}) *{{.Class}} {
	return t.dict[id]
}

func (t *{{$ClassTbl}}) Has(id {{.IDType}}) bool {
	if _, ok := t.dict[id]; ok {
		return true
	}
	return false
}

func (t *{{$ClassTbl}}) Version() int {
	return t.version
}

func (t *{{$ClassTbl}}) Name() string {
	return "{{.Class}}"
}

func (t *{{$ClassTbl}}) File() string {
	return "{{.File}}"
}

func (t *{{$ClassTbl}}) Load(data []byte) error {
	list := make([]*{{.Class}}, 0, len(t.list))
	if err := Unmarshal(data, &list); err != nil {
		return err
	}

	t.version++
	t.list = list
	t.dict = make(map[{{.IDType}}]*{{.Class}})

	{{- if .IDName}}
	for _, v := range list {
		t.dict[v.{{.IDName}}] = v
	}
	{{- end}}

	return nil
}
`

// 加载表格,支持加载全部,加载单个
// 支持从内存加载,支持从文件加载
const goManagerTpl = `
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
