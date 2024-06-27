package stack

import (
	"BilliardServer/Util/log"
	"runtime"
	"strings"
)

var haveContains []string //必须包含的前缀
var notContains []string  //必须排除的后缀

// 初始化
func InitPrint(have []string, not []string) {
	haveContains = have
	notContains = not
}

// 打印调用堆栈
func PrintCallStack() {
	//定义是否包含某个字符的函数
	funHaveContains := func(s string, ss []string) bool {
		for i, _ := range ss {
			if !strings.Contains(s, ss[i]) {
				return false
			}
		}

		return true
	}

	//定义非包含函数
	funNotContains := func(s string, ss []string) bool {
		for i, _ := range ss {
			if strings.Contains(s, ss[i]) {
				return false
			}
		}

		return true
	}
	log.Error("-----------stack start---------------")
	for skip := 2; ; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		if funHaveContains(file, haveContains) &&
			funNotContains(file, notContains) {
			//获取短文件名
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			log.Error(short, " line = ", line)
		}
	}
	log.Error("-----------stack end-----------------")
}
