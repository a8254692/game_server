// log
package log

import (
	"fmt"
	"log"
	"os"
)

// log句柄对象
type Stdout struct {
	FirstTag string      //输出标记字符
	Logger   *log.Logger //logger对象
	sign     string      //左边字符标记
}

func (s *Stdout) Init(firstTag string, sign string) {
	s.FirstTag = firstTag
	s.sign = sign
	s.Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
}

func (s *Stdout) Log(v ...interface{}) {
	s.Logger.Println(s.sign, s.FirstTag, s.sign, fmt.Sprint(v...))
}
