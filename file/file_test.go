/// @file file_test.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-09

package utfile_test
import (
	"testing"
	"fmt"
	"gitee.com/jntse/gotoolkit/file"
)

func TestFilePathScanner(t *testing.T) {
	scanner := utfile.NewFilePathScanner()
	scanner.Init("./")
	err := scanner.Scan()
	if err != nil {	
		fmt.Println(err)
		return
	}
	
	for _, k := range scanner.Files {
		fmt.Println(k)
	}
}


