package write

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

///////////////////////////////////////TXT写入类///////////////////////////////////////
type writeTXTFile struct {
	Content string
}

func (this *writeTXTFile) GetFileName(name string) string {

	if !strings.HasSuffix(name, Extension_TXT) {
		return name + Extension_TXT
	}
	return name
}

func (this *writeTXTFile) create() {

	this.Content = ""

}

func (this *writeTXTFile) WriteLine(contents ...interface{}) {
	content := ""
	for i := 0; i < len(contents); i++ {
		if i == len(contents)-1 {
			content += fmt.Sprint(contents[i]) + "\r\n"
		} else {
			content += fmt.Sprint(contents[i]) + ","
		}
	}
	this.Content += content
}

func (this *writeTXTFile) Save(path string, name string) string {

	var err error
	if _, err := os.Stat(path); err != nil {
		os.MkdirAll(path, 777)
	}

	file, err := os.Create(filepath.Join(path, this.GetFileName(name)))
	if err != nil {
		fmt.Println(err.Error())
	}
	file.WriteString(this.Content)
	file.Sync()
	file.Close()
	return filepath.Join(path, this.GetFileName(name))
}
