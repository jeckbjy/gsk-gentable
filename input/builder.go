package input

import (
	"gentable/base"
	"unicode"
)

const (
	TypeGdoc = "gdoc"
	TypeXlsx = "xlsx"
)

func init() {
	Add(&GDocBuilder{})
	Add(&XlsxBuilder{})
}

var gBuilderMap = make(map[string]Builder)

func Add(loader Builder) {
	gBuilderMap[loader.Type()] = loader
}

func Get(name string) Builder {
	return gBuilderMap[name]
}

type Options struct {
	Source string
	Sheets []string
	Auth   *GDocAuthConf
}

type LoadFunc func() (*base.Sheet, error)

// Builder 用于加载数据,可以从gdoc,也可以从本地目录或者本地文件
// 不同的数据源可能需要不同的认证参数,所有需要用到的参数都集中到Options中
// 不同的数据源的格式也需要转化成Sheet格式,为了提高并发,Load返回一个处理回调函数数组,而不是直接返回Sheet
// 一些限制规则:表名必须是字符开始,否则不导出,可以把不需要导出的表以下划线开始
type Builder interface {
	Type() string
	Load(opts *Options) ([]LoadFunc, error)
}

// 校验sheet是否需要处理
type Checker struct {
	sheets map[string]bool
}

func (c *Checker) Init(sheets []string) {
	c.sheets = make(map[string]bool)
	for _, s := range sheets {
		c.sheets[s] = true
	}
}

// Check 校验是否需要加载Sheet, false不需要,true需要
func (c *Checker) Check(sheet string) bool {
	if len(sheet) == 0 || !unicode.IsLetter(rune(sheet[0])) {
		return false
	}

	// 全部都需要加载
	if len(c.sheets) == 0 {
		return true
	}

	return c.sheets[sheet]
}
