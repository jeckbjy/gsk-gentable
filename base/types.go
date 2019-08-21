package base

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// 支持的数据类型
const (
	TYPE_ANY  = iota // 未知错误类型
	TYPE_SID         // 字符ID(str),要求不能重复,不能为空
	TYPE_NID         // 数值ID(int),要求不能重复,不能为空
	TYPE_INT         // 整数(int)
	TYPE_UINT        // 非负整数(uint)
	TYPE_I64         // 长精度整数()
	TYPE_U64         // 非负长整数
	TYPE_F32         // 单精度浮点数
	TYPE_F64         // 双精度浮点数
	TYPE_STR         // 字符类型
	TYPE_BOOL        // 布尔类型,可以是true,false,0,1
	TYPE_DATE        // 日期可以不填写,默认为00,如:2018-05-03 11:38:30
	TYPE_WEEK        // 日期:比如,每周5-12:00:00表示每周五12点,可以不填写周,则纯表示时间
	TYPE_ENUM        // 枚举类型,需要指定映射,ENUM[a:1,b:2,c:3]
	TYPE_MAP         // 字典类型,支持指定类型:MAP[STRING,STRING],禁止多层嵌套
	TYPE_LIST        // 数组,支持指定类型:LIST[INT],禁止多层嵌套,TODO:求和?
)

var typeName = [...]string{
	"ANY", "SID", "NID", "INT", "UINT", "INT64", "UINT64", "FLOAT", "DOUBLE", "STRING", "BOOL", "DATE", "WEEK", "ENUM", "MAP", "LIST",
}

var typeMap = map[string]int{
	"STRING":  TYPE_STR,
	"STR":     TYPE_STR,
	"SID":     TYPE_SID,
	"NID":     TYPE_NID,
	"INT":     TYPE_INT,
	"UINT":    TYPE_UINT,
	"LONG":    TYPE_I64,
	"ULONG":   TYPE_U64,
	"FLOAT":   TYPE_F32,
	"DOUBLE":  TYPE_F64,
	"BOOL":    TYPE_BOOL,
	"DATE":    TYPE_DATE,
	"WEEK":    TYPE_WEEK,
	"ENUM":    TYPE_ENUM,
	"MAP":     TYPE_MAP,
	"LIST":    TYPE_LIST,
	"I32":     TYPE_INT, // alias
	"U32":     TYPE_UINT,
	"I64":     TYPE_I64,
	"U64":     TYPE_U64,
	"INT64":   TYPE_I64,
	"UINT64":  TYPE_U64,
	"F32":     TYPE_F32,
	"F64":     TYPE_F64,
	"FLOAT32": TYPE_F32,
	"FLOAT64": TYPE_F64,
}

func isTypes(v int, types ...int) bool {
	for i := 0; i < len(types); i++ {
		if v == types[i] {
			return true
		}
	}

	return false
}

func isNumber(v int) bool {
	return (v >= TYPE_INT && v <= TYPE_F64) || v == TYPE_BOOL
}

func isIdentity(v int) bool {
	return v == TYPE_SID || v == TYPE_NID
}

// 类型信息
type Type struct {
	T0    int               // 主类型
	T1    int               // Map,List使用
	T2    int               // Map使用
	Enums map[string]string // 枚举类型
}

func (t *Type) String() string {
	switch t.T0 {
	case TYPE_MAP:
		return fmt.Sprintf("%s[%s,%s]", typeName[t.T0], typeName[t.T1], typeName[t.T2])
	case TYPE_LIST:
		return fmt.Sprintf("%s[%s]", typeName[t.T0], typeName[t.T1])
	default:
		return typeName[t.T0]
	}
}

func (t *Type) bind(v *int, s string) error {
	s = strings.TrimSpace(s)
	if vv, ok := typeMap[s]; ok {
		*v = vv
		return nil
	}

	return ErrUnknownType
}

func (t *Type) Parse(data string) error {
	data = strings.TrimSpace(data)
	data = strings.TrimRight(data, "]")
	data = strings.ToUpper(data)
	var head, tail string
	if index := strings.IndexByte(data, '['); index != -1 {
		head = data[:index]
		tail = data[index+1:]
	} else {
		head = data
		tail = ""
	}

	if err := t.bind(&t.T0, head); err != nil {
		return err
	}

	if tail == "" {
		if isTypes(t.T0, TYPE_LIST, TYPE_MAP) {
			t.T1 = TYPE_INT
			t.T2 = TYPE_INT
		}
		return nil
	}

	if !isTypes(t.T0, TYPE_LIST, TYPE_MAP, TYPE_ENUM) {
		return ErrUnknownTypeInfo
	}

	// 解析tail
	tokens := strings.FieldsFunc(tail, func(r rune) bool {
		return unicode.IsPunct(r)
	})

	if len(tokens) == 0 {
		return nil
	}

	switch t.T0 {
	case TYPE_LIST:
		return t.parseList(tokens)
	case TYPE_MAP:
		return t.parseMap(tokens)
	case TYPE_ENUM:
		return t.parseEnum(tokens)
	}

	return nil
}

func (t *Type) parseList(tokens []string) error {
	if len(tokens) != 1 {
		return ErrUnknownTypeInfo
	}
	if err := t.bind(&t.T1, tokens[0]); err != nil {
		return err
	}

	return nil
}

func (t *Type) parseMap(tokens []string) error {
	switch len(tokens) {
	case 1:
		if err := t.bind(&t.T1, tokens[0]); err != nil {
			return err
		}
		t.T2 = t.T1
	case 2:
		if err := t.bind(&t.T1, tokens[0]); err != nil {
			return err
		}
		if err := t.bind(&t.T2, tokens[1]); err != nil {
			return err
		}
	default:
		return ErrUnknownTypeInfo
	}

	return nil
}

func (t *Type) parseEnum(tokens []string) error {
	if len(tokens)%2 == 1 {
		return ErrInvalidEnum
	}

	exists := make(map[int]int)
	t.Enums = make(map[string]string)
	for i := 0; i < len(tokens); i += 2 {
		k := strings.TrimSpace(tokens[i])
		v, e := strconv.Atoi(strings.TrimSpace(tokens[i+1]))
		if e != nil {
			return ErrInvalidEnum
		}

		// value有重复
		if _, ok := exists[v]; ok {
			return ErrDuplicateEnum
		}

		exists[v] = 0
		t.Enums[k] = strconv.Itoa(v)
	}

	return nil
}
