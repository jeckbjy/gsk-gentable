package code

import (
	"bytes"
	"fmt"
	"gentable/code/gotpl"
	"io/ioutil"
	"os"
	"sync"
	"text/template"

	"gentable/base"

	"github.com/iancoleman/strcase"
)

type GoBuilder struct {
	tpl *template.Template
	mux sync.Mutex
}

func (b *GoBuilder) Type() string {
	return "go"
}

var goType = [...]string{
	base.TYPE_SID:  "string",
	base.TYPE_NID:  "int",
	base.TYPE_INT:  "int",
	base.TYPE_UINT: "uint",
	base.TYPE_I64:  "int64",
	base.TYPE_U64:  "uint64",
	base.TYPE_F32:  "float32",
	base.TYPE_F64:  "float64",
	base.TYPE_STR:  "string",
	base.TYPE_BOOL: "bool",
}

func (b *GoBuilder) toBaseType(t int) string {
	return goType[t]
}

func (b *GoBuilder) toType(types []*base.Type) []string {
	results := make([]string, len(types))
	for i, t := range types {
		switch t.T0 {
		case base.TYPE_LIST:
			t1 := b.toBaseType(t.T1)
			results[i] = "[]" + t1
		case base.TYPE_MAP:
			t1 := b.toBaseType(t.T1)
			t2 := b.toBaseType(t.T2)
			results[i] = fmt.Sprintf("map[%s]%s", t1, t2)
		case base.TYPE_DATE:
			results[i] = "Time"
		case base.TYPE_WEEK:
			results[i] = "Week"
		case base.TYPE_ENUM:
			results[i] = b.toBaseType(base.TYPE_INT)
		default:
			results[i] = b.toBaseType(t.T0)
		}
	}

	return results
}

// getTableTemplate table的template可以cache,更加高效
func (b *GoBuilder) getTableTemplate() (*template.Template, error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.tpl == nil {
		tpl, err := template.New("gotpl").Parse(gotpl.TplTable)
		if err != nil {
			return nil, base.Error(err)
		}
		b.tpl = tpl
	}

	return b.tpl, nil
}

func (b *GoBuilder) getAnnotation() string {
	return "// " + Warning
}

func (b *GoBuilder) toTag(dtype string) string {
	switch dtype {
	case "csv", "tsv", "psv":
		return "csv"
	default:
		return dtype
	}
}

func (b *GoBuilder) Build(sheet *base.Sheet, dtype string, opts *Options) error {
	type TField struct {
		Name string
		Type string
		Tags string
	}
	type TTable struct {
		Package string
		Class   string
		File    string    // 文件名
		IDType  string    // ID类型
		IDName  string    // 名字
		Fields  []*TField // 所有字段
	}
	pkg := opts.GetPkg()
	cls := strcase.ToCamel(sheet.ID)
	filename := fmt.Sprintf("%s.%s", sheet.ID, dtype)
	names := toCamel(sheet.Names)
	types := b.toType(sheet.Types)
	nmax := maxLen(names)
	tmax := maxLen(types)
	fields := make([]*TField, 0, len(types))
	idType := "int"
	idName := ""
	if sheet.HasIdType() {
		idType = b.toBaseType(sheet.Types[0].T0)
		idName = names[0]
	}

	tag := b.toTag(dtype)
	for i, n := range sheet.Names {
		f := &TField{
			Name: fmt.Sprintf("%-*s", nmax, names[i]),
			Type: fmt.Sprintf("%-*s", tmax, types[i]),
			Tags: fmt.Sprintf("`%s:\"%s,%+v\"`", tag, n, i),
		}
		fields = append(fields, f)
	}

	tbl := TTable{
		Package: pkg,
		Class:   cls,
		File:    filename,
		IDType:  idType,
		IDName:  idName,
		Fields:  fields,
	}

	tpl, err := b.getTableTemplate()
	if err != nil {
		return base.Error(err)
	}

	outfile := outFile(dtype, "go", strcase.ToSnake(sheet.ID)+"_tbl.go")
	return generate(tpl, b.getAnnotation(), &tbl, outfile)
}

func (b *GoBuilder) BuildAll(sheets []*base.Sheet, dtype string, opts *Options) error {
	_ = b.buildLib(dtype, opts)

	type TSheet struct {
		Name     string // 驼峰表名
		AlignKey string // 对齐后表名,XXXName
	}
	type TTables struct {
		Package string
		Sheets  []*TSheet
	}
	tbl := TTables{
		Package: opts.GetPkg(),
	}

	nmax := 0
	for _, s := range sheets {
		s := &TSheet{Name: strcase.ToCamel(s.ID)}
		if len(s.Name) > nmax {
			nmax = len(s.Name)
		}

		tbl.Sheets = append(tbl.Sheets, s)
	}

	nmax += 4
	for _, s := range tbl.Sheets {
		s.AlignKey = fmt.Sprintf("%-*s", nmax, s.Name+"Name")
	}

	// 只会用一次,不需要cache
	tpl, err := template.New("gomgr").Parse(gotpl.TplManager)
	if err != nil {
		return base.Error(err)
	}

	outfile := outFile(dtype, "go", "zzmanager.go")
	return generate(tpl, b.getAnnotation(), &tbl, outfile)
}

func (b *GoBuilder) buildLib(dtype string, opts *Options) error {
	pkg := fmt.Sprintf("package %s\n", opts.GetPkg())

	buf := bytes.Buffer{}
	buf.WriteString(pkg)
	buf.WriteString(gotpl.LibComm)
	outfile := outFile(dtype, "go", "zzlib.go")
	_ = ioutil.WriteFile(outfile, buf.Bytes(), os.ModePerm)

	buf = bytes.Buffer{}
	buf.WriteString(pkg)
	tag := b.toTag(dtype)
	switch tag {
	case "json":
		buf.WriteString(gotpl.LibJson)
	case "csv":
		buf.WriteString(gotpl.LibCsv)
	default:
		return fmt.Errorf("not support, %+v", dtype)
	}

	outfile = outFile(dtype, "go", fmt.Sprintf("zzlib_%s.go", dtype))
	return ioutil.WriteFile(outfile, buf.Bytes(), os.ModePerm)
}
