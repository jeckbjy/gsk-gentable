package main

import (
	"fmt"
	"gentable/base"
	"gentable/cmd/auth"
	"gentable/conf"
	"log"
	"os"
	"sync"
)

const kDefaultConfig = "config.hcl"

func main() {
	if len(os.Args) == 1 || os.Args[1] == "help" {
		fmt.Println("usage: gentable [task1,task2]")
		fmt.Println("usage: gentable auth [filename default credentials.json]")
		return
	}

	switch os.Args[1] {
	case "auth":
		runAuth()
	default:
		runTask()
	}
}

func runAuth() {
	log.Print("run auth\n")
	filename := ""
	if len(os.Args) > 2 {
		filename = os.Args[2]
	}

	if err := auth.Build(filename); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func runTask() {
	tasks := make([]string, 0)
	for i := 1; i < len(os.Args); i++ {
		tasks = append(tasks, os.Args[i])
	}

	// load config
	cfg := &conf.Config{}
	if err := cfg.Load(kDefaultConfig); err != nil {
		fmt.Println(base.Error(err))
		return
	}

	// 并发执行任务
	wg := sync.WaitGroup{}
	for _, name := range tasks {
		task, ok := cfg.Tasks[name]
		if !ok {
			fmt.Printf("cannot find task:%+v\n", name)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			ex := &Executor{}
			if err := ex.Build(cfg, task); err != nil {
				log.Println(base.Error(err))
			}
		}()
	}

	wg.Wait()
}
