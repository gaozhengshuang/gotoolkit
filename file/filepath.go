/// @file filepath.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-09

package utfile
import (
	"os"
	"path/filepath"
)

type FilePathScanner struct {
	root string
	exts []string		// 过滤列表，后缀名
	Files []string
}

func NewFilePathScanner() *FilePathScanner {
	f := &FilePathScanner{ Files: make([]string,0) }
	return f
}

func (f* FilePathScanner) Init(path string) {
	f.root = path
}

func (f* FilePathScanner) Num() int {
	return len(f.Files)
}

func (f* FilePathScanner) Scan() error {
	err := filepath.Walk(f.root, f.walkExec)
	return err
}

// 限定文件后缀名
func (f* FilePathScanner) Keep(ext ...string) {
	f.exts = append(f.exts, ext...)
}

func (f* FilePathScanner) IsExclude(path string) bool {
	if len(f.exts) == 0 { return true }

	ext := filepath.Ext(path)
	for _, v := range f.exts {
		if v == ext { return true }
	}
	return false
}

func (f* FilePathScanner) walkExec(path string, info os.FileInfo, err error) error {
	if info == nil     { return err }
	if info.IsDir()    { return nil }
	//fmt.Printf("name=%s size=%d byte path=%s\n", info.Name(), info.Size(), path)

	if f.IsExclude(path) {
		f.Files = append(f.Files, path)
	}
	return nil
}

