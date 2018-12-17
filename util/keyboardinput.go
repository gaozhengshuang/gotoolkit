/// @file keyboardinput.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util
import "os"
import "bufio"
import "strings"
import _ "fmt"
import "gitee.com/jntse/gotoolkit/log"

type KeyBordInput struct {
	C chan string
	start bool
}

func (t *KeyBordInput) Init() {
	t.C = make(chan string, 1)
}

func (t *KeyBordInput) Start() {
	if t.start == true { panic("KeyBordInput Start Already") }
	t.start = true
	log.Info("starting keyboard controler...")
	go t.run()
}

func (t *KeyBordInput) Ch() <-chan string {
	return t.C
}

func (t *KeyBordInput) Push(cmd string) {
	t.C <- cmd
}

// Deprecated !  
// use Push() instead
func (t *KeyBordInput) Insert(cmd string) {
	t.C <- cmd
}

// 键盘监控协程
func (t *KeyBordInput) run() {
	//*/
	defer log.Info("closing keyboard controler...")
	inputReader := bufio.NewReader(os.Stdin)
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			log.Error("keyboard err=%s", err)
			continue
		}

		switch input {
		case "exit\n": log.Info("start shutdown...")
		case "quit\n": log.Info("start shutdown...")
		default:	log.Info("keyboard input cmd=%s", input)	// remove '\n'
		}

		input = strings.TrimSpace(input)
		t.C <- input
	}
	/*/
		strings.TrimSpace("")		// do nothing, just for compile
		bufio.NewReader(os.Stdin)	// do nothing, just for compile
	//*/
}

