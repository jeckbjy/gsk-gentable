package base

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidIDIndex   = errors.New("invalid id index") // id索引只能为0
	ErrUnknownType      = errors.New("unknown type")
	ErrUnknownTypeInfo  = errors.New("unknown type info")
	ErrInvalidEnum      = errors.New("invalid enum")
	ErrNotFoundEnum     = errors.New("not found enum")
	ErrDuplicateEnum    = errors.New("duplicate enum")
	ErrInvalidColumn    = errors.New("invalid column")
	ErrInvalidJoinIndex = errors.New("invalid join index")
)

const (
	ROW_NAME = int(0)
	ROW_TYPE = 1
	ROW_NOTE = 2
)

// Sheet 一张数据表
// 前三行是数据表头,第一行列名,第二行类型,第三行注释
// 支持安列名合并成一个LIST数组,要求:NAME_{INDEX},限制只能是数值类型,索引从0递增
// 数据结构支持MAP,LIST,可以指定类型,如MAP[LONG,LONG], LIST[INT],不指定则默认INT类型,不能嵌套复合类型,
// 复杂结构,DATE,TIME,WEEK,ENUM
// 类型不区分大小写
// 限制:SID,NID必须是第一列,且不能重复,加载代码会自动添加映射,否则不会添加
// 支持默认值处理,布尔，数值默认为0
// TODO:是否需要支持注释,首字母以#开始?
type Sheet struct {
	ID     string     // 表名
	Names  []string   // 列名
	Types  []*Type    // 类型
	Notes  []string   // 注释
	Values [][]string // 数据
	merger *_Merger   // 用于合并数据
}

func (s *Sheet) GetTypes() []string {
	results := make([]string, len(s.Types))
	for i, t := range s.Types {
		results[i] = t.String()
	}

	return results
}

func (s *Sheet) HasIdType() bool {
	return isIdentity(s.Types[0].T0)
}

func (s *Sheet) Init(sheetID string) {
	s.ID = sheetID
}

// Append 添加一行数据
func (s *Sheet) Append(index int, cells []string) error {
	switch index {
	case ROW_NAME:
		// 校验是否需要合并列 NAME_INDEX
		s.merger = &_Merger{}
		names, err := s.merger.Build(cells)
		if err != nil {
			return err
		}
		s.Names = names
	case ROW_TYPE:
		if !s.merger.CheckLen(len(cells)) {
			return ErrInvalidColumn
		}
		// 先合并
		cells = s.merger.Join(cells)
		// parse type
		for i := 0; i < len(cells); i++ {
			t := &Type{}
			if err := t.Parse(cells[i]); err != nil {
				return err
			}
			if isIdentity(t.T0) && i != 0 {
				return ErrInvalidIDIndex
			}
			s.Types = append(s.Types, t)
		}

		return nil

	case ROW_NOTE:
		s.Notes = s.merger.Join(cells)
	default:
		if !s.merger.CheckLen(len(cells)) {
			return ErrInvalidColumn
		}

		// 修正数据
		for i, v := range cells {
			idx := s.merger.GetIndex(i)
			if data, err := s.fixData(v, s.Types[idx]); err != nil {
				return err
			} else {
				cells[i] = data
			}
		}

		cells = s.merger.Join(cells)
		s.Values = append(s.Values, cells)
	}

	return nil
}

// 修正空值,
func (s *Sheet) fixData(data string, info *Type) (string, error) {
	if data == "" {
		if isNumber(info.T0) {
			return "0", nil
		}

		return "", nil
	}

	switch info.T0 {
	case TYPE_ENUM:
		// 转换enum
		if info.Enums == nil {
			return "", ErrInvalidEnum
		}
		v, ok := info.Enums[data]
		if !ok {
			return "", ErrNotFoundEnum
		}
		return v, nil
	case TYPE_DATE:
		// 以00补足不全的数据 2018-05-03 11:38:30
		tkn := strings.FieldsFunc(data, func(r rune) bool {
			return r == '-' || r == ':' || r == ' '
		})
		for i := len(tkn); i < 6; i++ {
			tkn = append(tkn, "00")
		}
		result := fmt.Sprintf("%s-%s-%s %s:%s:%s",
			tkn[0], tkn[1], tkn[2],
			tkn[3], tkn[4], tkn[5])

		return result, nil
	case TYPE_WEEK:
		//5-12:00:00
		if len(data) == 0 {
			return data, nil
		}
		tkn := strings.FieldsFunc(data, func(r rune) bool {
			return r == ':'
		})
		for i := len(tkn); i < 3; i++ {
			tkn = append(tkn, "00")
		}
		result := fmt.Sprintf("%s:%s:%s", tkn[0], tkn[1], tkn[2])
		return result, nil
	}

	return data, nil
}

// 全部处理完后,校验数据是否合法,ID不能重复,不能为空
func (s *Sheet) Verify() error {
	if len(s.Values) == 0 {
		return nil
	}

	// 校验id是否重复,是否为空
	if isIdentity(s.Types[0].T0) {
		idMap := make(map[string]int)
		for i := 0; i < len(s.Values); i++ {
			id := s.Values[i][0]
			if id == "" {
				return fmt.Errorf("empty id, index %+v in sheet %+v", i, s.ID)
			}
			if _, ok := idMap[id]; ok {
				return fmt.Errorf("duplicate id,index %+v, id %+v in sheet %+v", i, id, s.ID)
			}
		}
	}

	return nil
}
