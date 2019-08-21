package code

const goLibComm = `
import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// 5-00:00:00
// weekday(8)|hour(8)|minute(8)|second(8)|clock(24)
type Week struct {
	d uint64
}

func (w *Week) Set(week, h, m, s int) {
	clock := uint64(h*3600 + m*60 + s)
	w.d = uint64(week+1)<<48 | uint64(h)<<40 | uint64(m)<<32 | uint64(s)<<24 | clock
}

// -1 mean not set,same sa time.Weekday
func (w Week) Weekday() int {
	return int((w.d>>48)&0xFF) - 1
}

func (w Week) Hour() int {
	return int((w.d >> 40) & 0xFF)
}

func (w Week) Minute() int {
	return int((w.d >> 32) & 0xFF)
}

func (w Week) Second() int {
	return int((w.d >> 24) & 0xFF)
}

func (w Week) Clock() int {
	return int(w.d & 0x00FFFFFF)
}

func storeValue(str string, val reflect.Value) error {
	switch val.Kind() {
	case reflect.String:
		val.SetString(str)
	case reflect.Bool:
		value, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		val.SetBool(value)
	case reflect.Int, reflect.Int32, reflect.Int64:
		// Parse the value to an int
		value, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		val.SetInt(value)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		val.SetUint(value)
	case reflect.Float32, reflect.Float64:
		// Parse the value to an float
		value, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		val.SetFloat(value)
	case reflect.Slice:
		if !isNumberType(val.Type().Elem().Kind()) {
			return fmt.Errorf("is not number type")
		}

		// TODO: remove empty??
		tkn := strings.FieldsFunc(str, func(r rune) bool {
			return r == '|' || r == ','
		})

		result := reflect.MakeSlice(val.Type(), len(tkn), len(tkn))
		for i, t := range tkn {
			t = strings.TrimSpace(t)
			e := result.Index(i)
			if err := storeValue(t, e); err != nil {
				return err
			}
		}
		val.Set(result)
	case reflect.Map:
		if !isNumberType(val.Type().Elem().Kind()) {
			return fmt.Errorf("is not number type")
		}

		tkn := strings.FieldsFunc(str, func(r rune) bool {
			// 范围有点广
			return unicode.IsPunct(r)
		})

		if len(tkn)%2 != 0 {
			return fmt.Errorf("bad map len")
		}

		typ := val.Type()
		// 每次都是new?
		keyv := reflect.New(typ.Key())
		valv := reflect.New(typ.Elem())
		result := reflect.MakeMap(typ)
		for i := 0; i < len(tkn); i += 2 {
			if err := storeValue(strings.TrimSpace(tkn[i]), keyv.Elem()); err != nil {
				return err
			}
			if err := storeValue(strings.TrimSpace(tkn[i+1]), valv.Elem()); err != nil {
				return err
			}
			result.SetMapIndex(keyv.Elem(), valv.Elem())
		}

		val.Set(result)
	case reflect.Struct:
		switch val.Interface().(type) {
		case time.Time:
			// 固定格式,Date
			tm, err := time.Parse("2006-01-02 15:04:05", str)
			if err != nil {
				return err
			}
			val.Set(reflect.ValueOf(tm))
		case Week:
			return parseWeek(str, val)
		default:
			return errors.New("not support struct")
		}
	default:
		return fmt.Errorf("not support:%+v", val.Kind())
	}

	return nil
}

func parseWeek(str string, val reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	weekStr := ""
	timeStr := str
	idx := strings.IndexByte(str, '-')
	if idx != -1 {
		weekStr = strings.TrimSpace(str[:idx])
		timeStr = str[idx+1:]
	}
	w := -1
	if weekStr != "" {
		w = mustAtoi(weekStr)
	}

	tkn := strings.Split(timeStr, ":")
	if len(tkn) != 3 {
		return fmt.Errorf("parse week fail, %+v", str)
	}
	h := mustAtoi(strings.TrimSpace(tkn[0]))
	m := mustAtoi(strings.TrimSpace(tkn[1]))
	s := mustAtoi(strings.TrimSpace(tkn[2]))
	week := Week{}
	week.Set(w, h, m, s)
	val.Set(reflect.ValueOf(week))
	return
}

func isNumberType(kind reflect.Kind) bool {
	if kind >= reflect.Int && kind <= reflect.Float64 && kind != reflect.Uintptr {
		return true
	}

	return false
}

func mustAtoi(s string) int {
	i, e := strconv.Atoi(s)
	if e != nil {
		panic(e)
	}
	return i
}
`

const goLibCsv = `
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

const goLibJson = `
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
