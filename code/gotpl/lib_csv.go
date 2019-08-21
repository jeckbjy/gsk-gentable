package gotpl

const LibCsv = `
import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal parse slice from data
func Unmarshal(data []byte, dest interface{}) error {
	typ := reflect.TypeOf(dest)
	if typ.Elem().Kind() != reflect.Slice {
		return errors.New("destination is not a slice")
	}

	if typ.Elem().Elem().Kind() != reflect.Ptr {
		return errors.New("destination element is not ptr")
	}

	source := bytes.NewReader(data)
	reader := csv.NewReader(source)
	records, err := reader.ReadAll()
	if err != nil && err != io.EOF {
		return err
	}

	// no data
	if len(records) < 4 {
		return nil
	}

	records = records[3:]

	//Create the slice the put the values in
	//Get the reflected value of dest
	destRv := reflect.ValueOf(dest).Elem()
	// Create a new reflected value containing a slice:
	sliceRv := reflect.MakeSlice(destRv.Type(), len(records), len(records))

	elemType := destRv.Type().Elem().Elem()

	fieldIdx := make([]int, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		tag := elemType.Field(i).Tag.Get("csv")
		idx := -1
		if len(tag) > 0 {
			pos := strings.LastIndexByte(tag, ',')
			if pos != -1 {
				idxStr := tag[pos+1:]
				if r, err := strconv.Atoi(idxStr); err == nil {
					idx = r
				}
			}
		}

		if idx != -1 && idx >= len(records[0]) {
			idx = -1
		}
		fieldIdx[i] = idx
	}

	for i, record := range records {
		item := sliceRv.Index(i)
		value := reflect.New(elemType)
		item.Set(value)
		for j := 0; j < len(fieldIdx); j++ {
			idx := fieldIdx[j]
			if idx == -1 {
				continue
			}
			fieldRv := value.Elem().Field(j)
			rawVal := record[idx]
			_ = storeValue(rawVal, fieldRv)
		}
	}

	destRv.Set(sliceRv)

	return nil
}
`
