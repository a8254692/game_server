// log
package log

import (
	"BilliardServer/Common"
	"fmt"
	"os"
)

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

var logMod string
var infoLogHandle Handle
var errorLogHandle Handle
var waringLogHandle Handle
var consoleLogger Stdout

func Init(folderName string, mode string) {
	logMod = mode

	if logMod != Common.ModeProd {
		infoLogHandle.Init(folderName, "infors", "|")
		consoleLogger.Init("print", "-")
	}

	errorLogHandle.Init(folderName, "errors", "*")
	waringLogHandle.Init(folderName, "waring", "-")
}

// 输出普通信息
func Info(v ...interface{}) {
	if logMod != Common.ModeProd {
		infoLogHandle.Log(fmt.Sprint(v...))
	}
}

// 输出错误信息
func Error(v ...interface{}) {
	errorLogHandle.Log(fmt.Sprint(v...))
}

// 输出警告信息
func Waring(v ...interface{}) {
	waringLogHandle.Log(fmt.Sprint(v...))
}

func Print(v ...interface{}) {
	if logMod != Common.ModeProd {
		consoleLogger.Log(fmt.Sprint(v...))
	}
}
