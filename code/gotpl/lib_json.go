package gotpl

const LibJson = `
import (
	"encoding/json"
	"reflect"
	"strings"
)

func Unmarshal(data []byte, dest interface{}) error {
	records := make([]map[string]string, 0, 0)
	if err := json.Unmarshal(data, records); err != nil {
		return err
	}

	//Create the slice the put the values in
	//Get the reflected value of dest
	destRv := reflect.ValueOf(dest).Elem()
	// Create a new reflected value containing a slice:
	sliceRv := reflect.MakeSlice(destRv.Type(), len(records), len(records))

	elemType := destRv.Type().Elem().Elem()

	fieldKey := make([]string, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		tag := elemType.Field(i).Tag.Get("json")
		key := ""
		if len(tag) > 0 {
			pos := strings.LastIndexByte(tag, ',')
			if pos != -1 {
				key = tag[:pos]
			} else {
				key = tag
			}
		}

		fieldKey[i] = key
	}

	for i, record := range records {
		item := sliceRv.Index(i)
		value := reflect.New(elemType)
		item.Set(value)
		for j := 0; j < len(fieldKey); j++ {
			key := fieldKey[j]
			if key == "" {
				continue
			}
			fieldRv := value.Elem().Field(j)
			rawVal := record[key]
			if rawVal == "" {
				continue
			}
			_ = storeValue(rawVal, fieldRv)
		}
	}

	destRv.Set(sliceRv)
	return nil
}
`
