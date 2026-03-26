package xlog

import (
	"context"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	MAX_DEPTH = 10
)

// Caller returns returns a Valuer that returns a pkg/file:line description of the caller.
func Caller(depth int) log.Valuer {
	return func(context.Context) interface{} {
		d := depth
		var file string
		var line int
		for i := 0; i < MAX_DEPTH; i++ {
			_, file, line, _ = runtime.Caller(d)
			if strings.LastIndex(file, "/log/filter.go") > 0 {
				d++
				continue
			}
			if strings.LastIndex(file, "/log/helper.go") > 0 {
				d++
				continue
			}
			break
		}
		_, file, line, _ = runtime.Caller(d)
		idx := strings.LastIndexByte(file, '/')
		return file[idx+1:] + ":" + strconv.Itoa(line)
	}
}
