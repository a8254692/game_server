package pprof

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

//文件前缀名
const PPROF_FRONT = "PPROF"

//文件后缀名
const PPROF_EXTENSION = ".prof"

//文件
var PPROF_FILE *os.File

var aaa []int

//写入标记
var PPROF_FLAG = true

var PPROF_RUNING_FLAG = false

func Init(proProt string) {
	if PPROF_FLAG {
		//Startpprof()

	}

	go func() {
		http.ListenAndServe("0.0.0.0:"+proProt, nil)
	}()
}

//开启性能监控
func Startpprof() {

	if PPROF_RUNING_FLAG {
		return
	}
	fileName := PPROF_FRONT + strconv.FormatInt(time.Now().Unix(), 10) + PPROF_EXTENSION
	var err error
	PPROF_FILE, err = os.Create(fileName)
	if err != nil {
		return
	}
	pprof.StartCPUProfile(PPROF_FILE)

	PPROF_RUNING_FLAG = true
}

//关闭性能监控
func Endpprof() {

	if PPROF_RUNING_FLAG {
		pprof.StopCPUProfile()
		PPROF_FILE.Close()
		PPROF_RUNING_FLAG = false
	}

}
