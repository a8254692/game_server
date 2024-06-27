package write

//写入文件接口
type Write interface {
	create()                              //创建或者打开文件
	WriteLine(contents ...interface{})    //写入
	Save(path string, name string) string //保存
}

type WriteFile_Type int //写入类型

const (
	WriteFile_Type_TXT   WriteFile_Type = iota //txt
	WriteFile_Type_EXCEL                       //Excel
)

const (
	Extension_EXCEL = ".xlsx" //Excel后缀名
	Extension_TXT   = ".txt"  //txt后缀名
)

//写入文件工具
//参数1，路径名
//参数2，文件名，不需要后缀名
//参数3，写入类型
//返回值，Write接口
//Write接口使用方式：
//使用WriteLine写入一行数据，可以输入任意数量的字符串，字符串会根据具体的写入类型进行分割
//输入结束，使用Save接口保存数据
func CreateWriteFile(typ WriteFile_Type) Write {
	switch typ {
	case WriteFile_Type_TXT:
		writer := writeTXTFile{}
		writer.create()
		return &writer
	case WriteFile_Type_EXCEL:
		writer := writeEXCELFile{}
		writer.create()
		return &writer
	}
	return nil
}


