package path

import (
	"os"
	"path/filepath"
	"strings"
)

const Download_Path = "download"
const Download_CDKEY_Path = "download/cdkey"

//获取根目录
func GetRootPath() string {
	root, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return filepath.Dir(os.Args[0])
	}
	return root
}

//获取下载目录
func GetDownloadPath() string {
	root := GetRootPath()
	download := filepath.Join(root, Download_Path)

	_, err := os.Stat(download)
	if err != nil {
		os.MkdirAll(download, 0777)
	}
	return download
}

//路径是否存在
func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

//获取一个文件的下载链接
//形式"download/xxx.txt"
func GetFileURL(absPath string) string {

	rootPath := GetRootPath()

	relativePath := strings.Replace(absPath, rootPath, "", -1)
	relativePath = strings.Replace(relativePath, string(filepath.Separator), "/", -1)
	return relativePath

}


//删除文件
func RemoveFile(path string) bool {
	del := os.Remove(path)
	if del != nil {
		return false
	}
	return true
}