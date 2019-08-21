package base

import (
	"fmt"
	"path/filepath"
	"runtime"
)

type werror struct {
	err  error
	file string
	line int
}

func (e *werror) Error() string {
	return fmt.Sprintf("%s, [%s:%d]", e.err.Error(), e.file, e.line)
}

func Error(err error) error {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		return &werror{
			err:  err,
			file: filepath.Base(file),
			line: line,
		}
	}

	return err
}
