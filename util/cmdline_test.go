/// @file cmdline_test.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util_test
import (
	"testing"
	"gitee.com/jntse/gotoolkit/util"
)

func TestCmdlineParse(t *testing.T) {

	type CmdlineArgument struct {
		//Conf        string
		//LogName     string
		//LogPath     string
	}

	args := &CmdlineArgument{}
	util.DoCmdLineParse(args)
	//if args.Conf == ""      { panic("-----input '-conf' cmdline argu-----") }
	//if args.LogName == ""   { panic("-----input '-logname' cmdline argu-----") }
	//if args.LogPath == ""   { panic("-----input '-logpath' cmdline argu-----") }
}

