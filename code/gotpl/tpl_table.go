package gotpl

const TplTable = `
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
