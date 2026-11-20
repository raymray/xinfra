package xerror

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/raymray/xinfra/xerror/internal/stacktrace"
)

type stack string

func callers(skip ...int) stack {
	s := defaultSkip
	if len(skip) > 0 {
		s = skip[0]
	}
	return stack(stacktrace.Take(s, stacktrace.Full))
}

// writeStackTo 将错误的堆栈信息写入 buf，返回递增后的 index。
// multiError 必须优先于 Error 检查，否则 errors.AsType 会穿透 Unwrap 误匹配。
func writeStackTo(buf *bytes.Buffer, index int, err error) int {
	if m, ok := errors.AsType[*multiError](err); ok {
		if m.text != "" {
			_, _ = fmt.Fprintf(buf, "%d. %s\n", index, m.text)
			index++
			if m.stack != "" {
				buf.WriteString(string(m.stack))
				buf.WriteByte('\n')
			}
		}
		for _, er := range m.errs {
			index = writeStackTo(buf, index, er)
		}
		return index
	}
	if e, ok := errors.AsType[*Error](err); ok {
		for e != nil {
			_, _ = fmt.Fprintf(buf, "%d. %s\n", index, e.text)
			index++
			if e.stack != "" {
				buf.WriteString(string(e.stack))
				buf.WriteByte('\n')
			}
			if e.err != nil {
				if next, ok2 := errors.AsType[*Error](e.err); ok2 {
					e = next
				} else {
					_, _ = fmt.Fprintf(buf, "%d. %s\n", index, e.err.Error())
					return index + 1
				}
			} else {
				return index
			}
		}
	}
	_, _ = fmt.Fprintf(buf, "%d. %s\n", index, err.Error())
	return index + 1
}
