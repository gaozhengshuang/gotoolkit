/// @file ./logwapper.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-11-21

package log

import (
	"fmt"
	_"flag"
	_"log"
	_"path"
	_"os"
	"github.com/cihub/seelog"
)


var gLogger seelog.LoggerInterface


// log 作为初始默认logger
func init() {
	//fmt.Println("logwapper.init()")
	//
	////
	//file := "./defaultlog/log.log"
	//if merr := os.MkdirAll(path.Dir(file), os.ModePerm); merr != nil {
	//	panic(merr)
	//}

	//logFile, err := os.OpenFile("./defaultlog/log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)  // 文件追加方式打开
	//if ( err != nil )   {
	//	panic("create file error !")
	//}

	////
	//dlogger := log.New(logFile,"[Debug]",log.LstdFlags)

}

//func Init(logdir string, logname string) bool {
func Init(logdir, logname string, loglevel ...string) bool {
	if logdir == "" { logdir = "./" }
	if logname == "" { logname = "default.log" }

	// check level
	loglvl := "info"
	if len(loglevel) != 0 {
		loglvl = loglevel[0]
		switch loglvl {
		case "trace":
		case "debug":
		case "info":
		case "error":
		case "critical":
		default:
			panic(fmt.Sprintf("loglevel should 'trace' 'debug' 'info' 'error' 'critical'"))
		}
	}

	logConfig := `
	<seelog minlevel="` + loglvl + `">
		<outputs formatid="gamelog">
			<buffered size="10000" flushperiod="1000">
				<rollingfile type="date" filename="` + logdir + "/" + logname + `" datepattern="20060102-15"/>
			</buffered>

			<filter levels="trace,debug,info,warn,error,critical">
				<console />
			</filter>
		</outputs>

		<formats>
			<format id="gamelog" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
		</formats>
	</seelog>
	`
	logger, err := seelog.LoggerFromConfigAsBytes([]byte(logConfig))
	if err != nil {
		panic(fmt.Sprintf("init logwapper fail [%s]", err))
		return false
	}
	gLogger = logger
	seelog.ReplaceLogger(gLogger)
	return true
}

func Flush() {
	seelog.Flush()
}


var Trace = seelog.Tracef
var Debug = seelog.Debugf
var Warn  = seelog.Warnf
var Info  = seelog.Infof
var Error = seelog.Errorf
var Fatal = seelog.Criticalf

//func Debug(format string, a ...interface{}) {
//	seelog.Debugf(format, a...)
//}
//
//func Info(format string, a ...interface{}) {
//	seelog.Infof(format, a...)
//}
//
//func Trace(format string, a ...interface{}) {
//	seelog.Tracef(format, a...)
//}
//
//func Warning(format string, a ...interface{}) {
//	seelog.Warnf(format, a...)
//}
//
//func Error(format string, a ...interface{}) {
//	seelog.Errorf(format, a...)
//}
//
//func Fatal(format string, a ...interface{}) {
//	seelog.Criticalf(format, a...)
//}
//
//func Enter() {
//	seelog.Info("")
//}

