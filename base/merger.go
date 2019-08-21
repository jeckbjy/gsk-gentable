package base

import (
	"regexp"
	"strconv"
	"strings"
)

// NAME_INDEX
var indexedNameReg = regexp.MustCompile("^[a-z0-9]+_[0-9]$")

// 合并信息
type _Merger struct {
	Index []int // 用于记录索引合并后映射关系,负数表示需要合并的索引,正数表示可以直接赋值
	Need  bool  // 是否需要合并
	Len   int   // 合并后列长度
}

// 通过原索引获取合并后索引
func (m *_Merger) GetIndex(i int) int {
	index := m.Index[i]
	if index < 0 {
		return -index
	}

	return index
}

// 校验原始长度是否合法
func (m *_Merger) CheckLen(length int) bool {
	return len(m.Index) == length
}

func (m *_Merger) Build(names []string) ([]string, error) {
	m.Index = make([]int, len(names))
	results := make([]string, 0)
	nameMap := make(map[string]int)
	for i, v := range names {
		if !indexedNameReg.MatchString(v) {
			m.Index[i] = len(results)
			results = append(results, v)
			continue
		}
		pos := strings.LastIndexByte(v, '_')
		name := v[:pos]
		index, _ := strconv.Atoi(v[pos+1:])
		if idx, ok := nameMap[name]; !ok {
			if index != 0 {
				return nil, ErrInvalidJoinIndex
			}
			// 不存在，新建
			idx = len(results)
			results = append(results, name)
			m.Index[i] = idx
			nameMap[name] = idx
		} else {
			m.Index[i] = -idx
		}
	}

	m.Need = len(results) != len(names)
	m.Len = len(results)
	return results, nil
}

func (m *_Merger) Join(values []string) []string {
	if !m.Need {
		return values
	}

	results := make([]string, m.Len)
	for i, idx := range m.Index {
		if idx < 0 {
			posi := -idx
			results[posi] += "," + values[i]
		} else {
			results[idx] = values[i]
		}
	}

	return results
}
