package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

	"gentable/base"
	"gentable/code"
	"gentable/conf"
	"gentable/data"
	"gentable/input"
)

// Executor 数据处理,分成两个阶段:
// 1:对每个表加载,转换,并输出数据和代码
// 2:所有表格处理结束后,再执行代码管理器的输出
// 问题:init中switch硬编码
// TODO:print log info
type Executor struct {
	cfg       *conf.Config
	loader    input.Builder
	inputOpts *input.Options
	codeOpts  *code.Options
	wg        sync.WaitGroup
	mux       sync.Mutex
	outputs   []*OutputBuilder
	errors    []error
}

func (e *Executor) Build(cfg *conf.Config, task *conf.Task) error {
	if err := e.init(cfg, task); err != nil {
		return err
	}

	// build sheet
	invokers, err := e.loader.Load(e.inputOpts)
	if err != nil {
		return err
	}

	for _, fn := range invokers {
		sfn := fn
		e.run(func() error {
			sheet, err := sfn()
			if err != nil {
				return err
			}

			for _, ob := range e.outputs {
				ob.Build(e, sheet)
			}

			return nil
		})
	}

	e.wg.Wait()

	//build manager,不会处理太多,不用并发处理了
	for _, ob := range e.outputs {
		if len(ob.sheets) == 0 || len(ob.cbuilders) == 0 {
			continue
		}
		for _, builder := range ob.cbuilders {
			err := builder.BuildAll(ob.sheets, ob.dtype, e.codeOpts)
			e.tryAddError(err)
		}
	}

	e.dump()

	return nil
}

func (e *Executor) dump() {
	for _, err := range e.errors {
		fmt.Println(err)
	}
}

func (e *Executor) init(cfg *conf.Config, task *conf.Task) error {
	e.cfg = cfg
	e.codeOpts = &code.Options{}
	if cfg.Options != nil {
		opts := e.codeOpts
		opts.GoPkg = cfg.Options.GoPkg
	}

	// 构建task
	for _, c := range task.Outputs {
		ob := &OutputBuilder{}
		err := ob.Init(c)
		if err != nil {
			e.errors = append(e.errors, err)
			continue
		}
		e.outputs = append(e.outputs, ob)
	}

	// 构建loader
	loader := input.Get(task.Type)
	if loader == nil {
		return fmt.Errorf("cannot find loader, %+v", task.Type)
	}

	// 初始化Options
	opts := &input.Options{}
	opts.Source = task.Source
	opts.Sheets = task.Sheets

	// build other inputOpts
	switch task.Type {
	case input.TypeGdoc:
		auth, err := input.LoadAuthFromFile(task.Auth)
		if err != nil {
			return err
		}

		opts.Auth = auth
	}

	e.loader = loader
	e.inputOpts = opts

	return nil
}

func (e *Executor) run(fn func() error) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		err := fn()
		e.tryAddError(err)
	}()
}

func (e *Executor) tryAddError(err error) {
	if err != nil {
		e.mux.Lock()
		e.errors = append(e.errors, err)
		e.mux.Unlock()
	}
}

/////////////////////////////////////////////
// Output channel
/////////////////////////////////////////////
type OutputBuilder struct {
	dtype     string
	filters   map[string]bool
	dbuilder  data.Builder
	cbuilders []code.Builder
	sheets    []*base.Sheet
	mux       sync.Mutex
}

func (o *OutputBuilder) Init(c *conf.Output) error {
	o.dtype = c.Data
	for _, s := range c.Sheets {
		o.filters[s] = true
	}

	// build data builder
	dbuilder := data.Get(c.Data)
	if dbuilder == nil {
		return fmt.Errorf("can not find data builder, %+v", c.Data)
	}

	if err := os.MkdirAll(path.Join("output", o.dtype, "data"), os.ModePerm); err != nil {
		return err
	}

	o.dbuilder = dbuilder

	errs := ""
	for _, name := range c.Code {
		builder := code.Get(name)
		if builder != nil {
			_ = os.MkdirAll(path.Join("output", o.dtype, builder.Type()), os.ModePerm)
			o.cbuilders = append(o.cbuilders, builder)
		} else {
			errs += fmt.Sprintf("cannot find code builder, %+v\n", name)
		}
	}

	if errs != "" {
		return errors.New(errs)
	}

	return nil
}

func (o *OutputBuilder) Build(e *Executor, sheet *base.Sheet) {
	if len(o.filters) > 0 {
		if _, ok := o.filters[sheet.ID]; !ok {
			return
		}
	}

	o.mux.Lock()
	o.sheets = append(o.sheets, sheet)
	o.mux.Unlock()

	e.run(func() error {
		fmt.Printf("build data, %+v, %+v\n", o.dtype, sheet.ID)
		return o.dbuilder.Build(sheet)
	})

	for _, cb := range o.cbuilders {
		tmp := cb
		e.run(func() error {
			fmt.Printf("build code, %+v, %+v, %+v\n", o.dtype, tmp.Type(), sheet.ID)
			return tmp.Build(sheet, o.dtype, e.codeOpts)
		})
	}
}
