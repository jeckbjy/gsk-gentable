package conf

import (
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

type Output struct {
	Data   string   `hcl:"data"`   // 需要输出数据的类型,如[csv,tsv,json]
	Code   []string `hcl:"code"`   // 需要输出代码的类型,如[go]
	Sheets []string `hcl:"sheets"` // 需要输出的表名,空表示全部
}

// Task一个导出任务包括一个输入源和1个或多个输出源
type Task struct {
	Auth    string    `hcl:"auth"`    // gdoc配置地址,默认为空,读取oauth.json配置
	Type    string    `hcl:"type"`    // 类型[gdoc,xlsx]
	Name    string    `hcl:"name"`    // 任务名
	Source  string    `hcl:"source"`  // 可以是文件路径,也可以是SpeadsheetID
	Sheets  []string  `hcl:"sheets"`  // 需要加载的表名,空表示全部
	Outputs []*Output `hcl:"outputs"` // 多个输出源,比如服务器,客户端分别导出不同的表,不同的类型
}

type Options struct {
	GoPkg string `hcl:"gopkg"` // go包名,默认tbl
}

type Config struct {
	Auth    string           `hcl:"auth"` // gdoc配置地址,默认为空,读取oauth.json配置
	Type    string           `hcl:"type"` // 类型[gdoc,xlsx]
	Options *Options         `hcl:"options"`
	Tasks   map[string]*Task `hcl:"tasks"`
}

func (c *Config) Load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return c.Unmarshal(string(data))
}

func (c *Config) Unmarshal(data string) error {
	if err := hcl.Unmarshal([]byte(data), c); err != nil {
		return err
	}

	// merge global config
	for _, task := range c.Tasks {
		merge(&task.Auth, c.Auth)
		merge(&task.Type, c.Type)
	}

	return nil
}

func merge(dst *string, src string) {
	if *dst == "" {
		*dst = src
	}
}
