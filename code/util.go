package code

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/iancoleman/strcase"
)

func maxLen(strs []string) int {
	l := 0
	for _, s := range strs {
		if len(s) > l {
			l = len(s)
		}
	}

	return l
}

func toCamel(strs []string) []string {
	results := make([]string, 0, len(strs))
	for _, s := range strs {
		results = append(results, strcase.ToCamel(s))
	}

	return results
}

func outFile(dtype string, ctype string, filename string) string {
	return path.Join("output", dtype, ctype, filename)
}

// generate 调用模板生成最终文件
func generate(tpl *template.Template, annotation string, data interface{}, target string) error {
	out := bytes.Buffer{}
	out.WriteString(annotation)
	if err := tpl.Execute(&out, data); err != nil {
		return err
	}

	return ioutil.WriteFile(target, out.Bytes(), os.ModePerm)
}
