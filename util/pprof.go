/// @file pprof.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util
import "os"
import "fmt"
import "net/http"
import _"net/http/pprof"
import "runtime/pprof"

var importuse1 = os.Create
var importuse2 = fmt.Printf
var importuse3 = pprof.StartCPUProfile

func OpenPProf(handle func()) {
	/*/
	handle()
	/*/
	os.MkdirAll("./pprof",0777)
	name := fmt.Sprintf("./pprof/cpu_profile.%d.prof",os.Getpid())
	ppfile, err := os.Create(name)
	if err != nil { panic(err) }
	pprof.StartCPUProfile(ppfile)
	handle()
	defer ppfile.Close()
	defer pprof.StopCPUProfile()
	//*/
}

func OpenNetPProf(ip string, port int32, handle func()) {
	/*/
	handle()
	/*/
	addr := fmt.Sprintf("%s:%d", ip, port)
	fmt.Printf("net pprof listening[%s]...\n", addr)
	FunNetPProfListen := func(addr string) {
		fmt.Println(http.ListenAndServe(addr, nil))
	}
	go FunNetPProfListen(addr)
	handle()
	//*/
}

