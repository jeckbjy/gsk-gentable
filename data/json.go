package data

import (
	"encoding/json"
	"gentable/base"
	"io/ioutil"
	"os"
)

func NewJsonBuilder() *JsonBuilder {
	return &JsonBuilder{}
}

type JsonBuilder struct {
}

func (b *JsonBuilder) Type() string {
	return "json"
}

func (b *JsonBuilder) Build(sheet *base.Sheet) error {
	type KVMap map[string]interface{}
	table := make([]KVMap, 0, len(sheet.Values))
	column := len(sheet.Names)
	for _, v := range sheet.Values {
		data := make(KVMap)
		for i := 0; i < column; i++ {
			key := sheet.Names[i]
			val := v[i]
			data[key] = val
		}
		table = append(table, data)
	}

	result, err := json.MarshalIndent(table, "", "\t")
	if err != nil {
		return err
	}

	output := outFile(b.Type(), sheet.ID)
	return ioutil.WriteFile(output, []byte(result), os.ModePerm)
}
