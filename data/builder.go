package data

import (
	"gentable/base"
	"path"
)

func init() {
	Add(&JsonBuilder{})
	Add(newDsvBuilder(kTypeCsv))
	Add(newDsvBuilder(kTypeTsv))
	Add(newDsvBuilder(kTypePsv))
}

var gBuilderMap = make(map[string]Builder)

func Add(builder Builder) {
	gBuilderMap[builder.Type()] = builder
}

func Get(name string) Builder {
	return gBuilderMap[name]
}

// Builder 数据输出到 data/{type} 目录下
type Builder interface {
	Type() string
	Build(sheet *base.Sheet) error
}

func outFile(dtype string, sheetID string) string {
	return path.Join("output", dtype, "data", sheetID+"."+dtype)
}
