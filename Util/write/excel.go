package write

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tealeg/xlsx"
)

//////////////////////////////////////EXCEL文件写入类///////////////////////////////////////
type writeEXCELFile struct {
	File  *xlsx.File
	Sheet *xlsx.Sheet
}

func (this *writeEXCELFile) GetFileName(name string) string {

	if !strings.HasSuffix(name, Extension_EXCEL) {
		return name + Extension_EXCEL
	}
	return name
}

func (this *writeEXCELFile) create() {
	this.File = xlsx.NewFile()
	var err error
	this.Sheet, err = this.File.AddSheet("Sheet1")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (this *writeEXCELFile) WriteLine(contents ...interface{}) {
	row := this.Sheet.AddRow()
	for i := 0; i < len(contents); i++ {
		value := fmt.Sprint(contents[i])
		cell := row.AddCell()
		cell.Value = value
	}
}

func (this *writeEXCELFile) Save(path string, name string) string {
	if _, err := os.Stat(path); err != nil {
		os.MkdirAll(path, 777)
	}
	fileName := filepath.Join(path, this.GetFileName(name))
	err := this.File.Save(fileName)
	if err != nil {
		fmt.Println(err.Error())
	}
	return fileName
}
